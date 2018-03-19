package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "flow",
		Usage:  "Explains what flows are and how to use them.",
		Action: flowAction,
		Subcommands: []cli.Command{
			{
				Name:   "list",
				Usage:  "Shows available flows",
				Action: flowListAction,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  util.InstalledFlg.String(),
						Usage: "Only show installed flows",
					},
					cli.StringFlag{
						Name:  util.FormatFlg.String(),
						Usage: "The format for the output. Can be listed (default) or \"json\" encoded.",
					},
					cli.StringFlag{
						Name:  util.OutputFlg.String(),
						Usage: "A filepath to output lexicon data to. If the flag is not set, outputs to cli.",
					},
				},
			},
		},
	}, false)
}

func flowAction(ctx *cli.Context) {
	err := flow(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func flow(ctx *cli.Context) error {
	fmt.Println("Flow is an automated workflow.")
	return nil
}

func flowListAction(ctx *cli.Context) {
	err := flowList(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func flowList(ctx *cli.Context) error {
	// TODO: discovery layer
	availableFlows := []string{
		"codelingo/search",
		"codelingo/review",
	}

	if ctx.IsSet(util.InstalledFlg.Long) {
		// TODO: filter by installed here
	}

	err := outputBytes(ctx.String("output"), getFormat(ctx.String("format"), availableFlows))
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
