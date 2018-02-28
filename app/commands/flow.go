package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
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
				},
			},
		},
	}, false)
}

func flowAction(ctx *cli.Context) {
	err := flow(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func flow(ctx *cli.Context) error {
	fmt.Println("TOOD: explain flows and how to use them.")
	return nil
}

func flowListAction(ctx *cli.Context) {
	err := flowList(ctx)
	if err != nil {
		util.OSErr(err)
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

	for _, flow := range availableFlows {
		fmt.Printf("- %s\n", flow)
	}

	return nil
}
