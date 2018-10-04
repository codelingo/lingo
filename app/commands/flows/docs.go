package flows

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

var DocsCmd = cli.Command{
	Name:        "docs",
	Usage:       "Generate documentation from Tenets",
	Subcommands: cli.Commands{PullRequestCmd},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  util.OutputFlg.String(),
			Usage: "File to save found results to.",
		},
		cli.StringFlag{
			Name:  "template, t",
			Value: "default",
			Usage: "The template to use when generating docs.",
		},
	},
	Description: `
""$ lingo search <filename>" .
`[1:],
	Action: docsAction,
}

func docsAction(ctx *cli.Context) {
	docs, err := docsCMD(ctx)
	if err != nil {

		// Debugging
		util.Logger.Debugw("docsAction", "err_stack", errors.ErrorStack(err))

		util.FatalOSErr(err)
		return
	}

	print(docs)
}

func docsCMD(cliCtx *cli.Context) (string, error) {
	return "docs", nil
}
