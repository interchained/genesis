package app

import (
	"embed"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/plush"
	"github.com/gobuffalo/plushgen"
	"github.com/interchained/genesis/genesis/pkg/xgenny"
	"github.com/interchained/genesis/genesis/pkg/xstrings"
	"github.com/interchained/genesis/genesis/templates/testutil"
)

var (
	//go:embed stargate/* stargate/**/*
	fsStargate embed.FS

	tpl = xgenny.NewEmbedWalker(fsStargate, "stargate/")
)

// New returns the generator to scaffold a new Cosmos SDK app
func New(opts *Options) (*genny.Generator, error) {
	g := genny.New()
	if err := g.Box(tpl); err != nil {
		return g, err
	}
	ctx := plush.NewContext()
	ctx.Set("ModulePath", opts.ModulePath)
	ctx.Set("AppName", opts.AppName)
	ctx.Set("OwnerAndRepoName", opts.OwnerAndRepoName)
	ctx.Set("OwnerName", opts.OwnerName)
	ctx.Set("BinaryNamePrefix", opts.BinaryNamePrefix)
	ctx.Set("AddressPrefix", opts.AddressPrefix)
	ctx.Set("title", strings.Title)

	// Used for proto package name
	ctx.Set("formatOwnerName", xstrings.FormatUsername)

	if err := testutil.Register(ctx, g, opts.AppPath); err != nil {
		return g, err
	}

	g.Transformer(plushgen.Transformer(ctx))
	g.Transformer(genny.Replace("{{appName}}", opts.AppName))
	g.Transformer(genny.Replace("{{binaryNamePrefix}}", opts.BinaryNamePrefix))
	return g, nil
}
