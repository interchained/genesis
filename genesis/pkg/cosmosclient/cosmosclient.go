package cosmosclient

import (
	"io"
	"os"

	"github.com/interchained/cosmos-sdk/client"
	"github.com/interchained/cosmos-sdk/client/flags"
	"github.com/interchained/cosmos-sdk/client/tx"
	cryptocodec "github.com/interchained/cosmos-sdk/crypto/codec"
	sdk "github.com/interchained/cosmos-sdk/types"
	"github.com/interchained/cosmos-sdk/types/tx/signing"
	authtypes "github.com/interchained/cosmos-sdk/x/auth/types"
	staking "github.com/interchained/cosmos-sdk/x/staking/types"
	"github.com/interchained/gpn/app/params"
	rpchttp "github.com/interchained/genesismint/rpc/client/http"
)

const (
	defaultGasAdjustment = 1.0
	defaultGasLimit      = 300000
)

// NewContext creates a new client context.
func NewContext(
	c *rpchttp.HTTP,
	out io.Writer,
	chainID,
	home string,
) client.Context {
	encodingConfig := params.MakeEncodingConfig()
	authtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	sdk.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	staking.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return client.Context{}.
		WithChainID(chainID).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithOutput(out).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(home).
		WithClient(c).
		WithSkipConfirmation(true)
}

// NewFactory creates a new tx factory.
func NewFactory(clientCtx client.Context) tx.Factory {
	return tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(defaultGasLimit).
		WithGasAdjustment(defaultGasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(clientCtx.TxConfig)
}
