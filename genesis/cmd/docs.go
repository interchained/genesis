package starportcmd

import (
	"github.com/spf13/cobra"
	"github.com/interchained/genesis/docs"
	"github.com/interchained/genesis/genesis/pkg/localfs"
	"github.com/interchained/genesis/genesis/pkg/markdownviewer"
)

func NewDocs() *cobra.Command {
	c := &cobra.Command{
		Use:   "docs",
		Short: "Show Genesis Starport docs",
		Args:  cobra.ExactArgs(0),
		RunE:  docsHandler,
	}
	return c
}

func docsHandler(cmd *cobra.Command, args []string) error {
	path, cleanup, err := localfs.SaveTemp(docs.Docs)
	if err != nil {
		return err
	}
	defer cleanup()

	return markdownviewer.View(path)
}
