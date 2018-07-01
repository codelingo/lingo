package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "run",
		Usage:  "Run the given flow in the current directory.",
		Action: runAction,
		Subcommands: []cli.Command{
			reviewCommand,
			searchCommand,
			codemodCommand,
		},
	}, false, false)
}

func runAction(ctx *cli.Context) {
	if err := run(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
}

func run(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("Failed to run flow - no flow given.")
	} else {
		invalidFlowName := c.Args()[0]
		return errors.Errorf("Failed to run flow - '%s' is not installed. View your installed flows by running `$ lingo list flows`.", invalidFlowName)
	}
}
