package chaincmdrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/interchained/genesis/genesis/pkg/chaincmd"
	"github.com/interchained/genesis/genesis/pkg/cmdrunner/step"
	"github.com/interchained/genesis/genesis/pkg/cosmosver"
)

// Start starts the blockchain.
func (r Runner) Start(ctx context.Context, args ...string) error {
	return r.run(
		ctx,
		runOptions{wrappedStdErrMaxLen: 50000},
		r.chainCmd.StartCommand(args...),
	)
}

// LaunchpadStartRestServer start launchpad rest server.
func (r Runner) LaunchpadStartRestServer(ctx context.Context, apiAddress, rpcAddress string) error {
	return r.run(
		ctx,
		runOptions{wrappedStdErrMaxLen: 50000},
		r.chainCmd.LaunchpadRestServerCommand(apiAddress, rpcAddress),
	)
}

// Init inits the blockchain.
func (r Runner) Init(ctx context.Context, moniker string) error {
	return r.run(ctx, runOptions{}, r.chainCmd.InitCommand(moniker))
}

// KV holds a key, value pair.
type KV struct {
	key   string
	value string
}

// NewKV returns a new key, value pair.
func NewKV(key, value string) KV {
	return KV{key, value}
}

// LaunchpadSetConfigs updates configurations for a launchpad app.
func (r Runner) LaunchpadSetConfigs(ctx context.Context, kvs ...KV) error {
	for _, kv := range kvs {
		if err := r.run(
			ctx,
			runOptions{},
			r.chainCmd.LaunchpadSetConfigCommand(kv.key, kv.value),
		); err != nil {
			return err
		}
	}
	return nil
}

var gentxRe = regexp.MustCompile(`(?m)"(.+?)"`)

// Gentx generates a genesis tx carrying a self delegation.
func (r Runner) Gentx(
	ctx context.Context,
	validatorName,
	selfDelegation string,
	options ...chaincmd.GentxOption,
) (gentxPath string, err error) {
	b := &bytes.Buffer{}

	if err := r.run(ctx, runOptions{
		stdout: b,
		stderr: io.MultiWriter(b, os.Stderr),
		stdin:  os.Stdin,
	}, r.chainCmd.GentxCommand(validatorName, selfDelegation, options...)); err != nil {
		return "", err
	}

	return gentxRe.FindStringSubmatch(b.String())[1], nil
}

// CollectGentxs collects gentxs.
func (r Runner) CollectGentxs(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.CollectGentxsCommand())
}

// ValidateGenesis validates genesis.
func (r Runner) ValidateGenesis(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.ValidateGenesisCommand())
}

// UnsafeReset resets the blockchain database.
func (r Runner) UnsafeReset(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.UnsafeResetCommand())
}

// ShowNodeID shows node id.
func (r Runner) ShowNodeID(ctx context.Context) (nodeID string, err error) {
	b := &bytes.Buffer{}
	err = r.run(ctx, runOptions{stdout: b}, r.chainCmd.ShowNodeIDCommand())
	nodeID = strings.TrimSpace(b.String())
	return
}

// NodeStatus keeps info about node's status.
type NodeStatus struct {
	ChainID string
}

// Status returns the node's status.
func (r Runner) Status(ctx context.Context) (NodeStatus, error) {
	b := newBuffer()

	if err := r.run(ctx, runOptions{stdout: b, stderr: b}, r.chainCmd.StatusCommand()); err != nil {
		return NodeStatus{}, err
	}

	var chainID string

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return NodeStatus{}, err
	}

	switch r.chainCmd.SDKVersion() {
	case cosmosver.StargateZeroFourtyAndAbove:
		out := struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"NodeInfo"`
		}{}

		if err := json.Unmarshal(data, &out); err != nil {
			return NodeStatus{}, err
		}

		chainID = out.NodeInfo.Network
	default:
		out := struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		}{}

		if err := json.Unmarshal(data, &out); err != nil {
			return NodeStatus{}, err
		}

		chainID = out.NodeInfo.Network
	}

	return NodeStatus{
		ChainID: chainID,
	}, nil
}

// BankSend sends amount from fromAccount to toAccount.
func (r Runner) BankSend(ctx context.Context, fromAccount, toAccount, amount string) error {
	b := newBuffer()
	opt := []step.Option{
		r.chainCmd.BankSendCommand(fromAccount, toAccount, amount),
	}

	if r.chainCmd.KeyringPassword() != "" {
		input := &bytes.Buffer{}
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		opt = append(opt, step.Write(input.Bytes()))
	}

	if err := r.run(ctx, runOptions{stdout: b}, opt...); err != nil {
		if strings.Contains(err.Error(), "key not found") || // stargate
			strings.Contains(err.Error(), "unknown address") || // launchpad
			strings.Contains(b.String(), "item could not be found") { // launchpad
			return errors.New("account doesn't have any balances")
		}

		return err
	}

	out := struct {
		Code  int    `json:"code"`
		Error string `json:"raw_log"`
	}{}

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	if out.Code > 0 {
		return fmt.Errorf("cannot send tokens (SDK code %d): %s", out.Code, out.Error)
	}

	return nil
}

// Export exports the state of the chain into the specified file
func (r Runner) Export(ctx context.Context, exportedFile string) error {
	exportedState := &bytes.Buffer{}
	if err := r.run(ctx, runOptions{stdout: exportedState}, r.chainCmd.ExportCommand()); err != nil {
		return err
	}

	// Save the new state
	return ioutil.WriteFile(exportedFile, exportedState.Bytes(), 0644)
}

// EventSelector is used to query events.
type EventSelector struct {
	typ   string
	attr  string
	value string
}

// NewEventSelector creates a new event selector.
func NewEventSelector(typ, addr, value string) EventSelector {
	return EventSelector{typ, addr, value}
}

// Event represents a TX event.
type Event struct {
	Type       string
	Attributes []EventAttribute
	Time       time.Time
}

// EventAttribute holds event's attributes.
type EventAttribute struct {
	Key   string
	Value string
}

// QueryTxEvents queries tx events by event selectors.
func (r Runner) QueryTxEvents(
	ctx context.Context,
	selector EventSelector,
	moreSelectors ...EventSelector,
) ([]Event, error) {
	// prepare the slector.
	var list []string

	eventsSelectors := append([]EventSelector{selector}, moreSelectors...)

	for _, event := range eventsSelectors {
		list = append(list, fmt.Sprintf("%s.%s=%s", event.typ, event.attr, event.value))
	}

	query := strings.Join(list, "&")

	// execute the commnd and parse the output.
	b := newBuffer()

	if err := r.run(ctx, runOptions{stdout: b}, r.chainCmd.QueryTxEventsCommand(query)); err != nil {
		return nil, err
	}

	out := struct {
		Txs []struct {
			Logs []struct {
				Events []struct {
					Type  string `json:"type"`
					Attrs []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"attributes"`
				} `json:"events"`
			} `json:"logs"`
			TimeStamp string `json:"timestamp"`
		} `json:"txs"`
	}{}

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	var events []Event

	for _, tx := range out.Txs {
		for _, log := range tx.Logs {
			for _, e := range log.Events {
				var attrs []EventAttribute
				for _, attr := range e.Attrs {
					attrs = append(attrs, EventAttribute{
						Key:   attr.Key,
						Value: attr.Value,
					})
				}

				txTime, err := time.Parse(time.RFC3339, tx.TimeStamp)
				if err != nil {
					return nil, err
				}

				events = append(events, Event{
					Type:       e.Type,
					Attributes: attrs,
					Time:       txTime,
				})
			}
		}
	}

	return events, nil
}
