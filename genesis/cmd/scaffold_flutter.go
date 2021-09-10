package starportcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/interchained/genesis/genesis/pkg/clispinner"
	"github.com/interchained/genesis/genesis/services/scaffolder"
)

// NewScaffoldFlutter scaffolds a Flutter app for a chain.
func NewScaffoldFlutter() *cobra.Command {
	c := &cobra.Command{
		Use:   "flutter",
		Short: "A Flutter app for your XRC20 token chain",
		Args:  cobra.NoArgs,
		RunE:  scaffoldFlutterHandler,
	}

	c.Flags().StringP(flagPath, "p", "./flutter", "path to scaffold content of the Flutter app")

	return c
}

func scaffoldFlutterHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("Scaffolding...")
	defer s.Stop()

	path, _ := cmd.Flags().GetString(flagPath)

	if err := scaffolder.Flutter(path); err != nil {
		return err
	}

	s.Stop()
	fmt.Printf("\n🎉 Scaffold a Flutter app.\n\n")

	return nil
}
