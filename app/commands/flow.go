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
