package commands

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"

	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "uninstall",
		Usage:  "uninstall an Action",
		Action: uninstallAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "owner",
				Value: "codelingo",
				Usage: "Owner of the Flow",
			},
		},
	}, false, false)
}

func uninstallAction(ctx *cli.Context) {
	if err := uninstall(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Printf("Success! %s Flow has been uninstalled.\n", ctx.Args()[0])
}

func uninstall(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("Failed to uninstall Flow - no Flow given.")
	}

	home, err := util.LingoHome()
	if err != nil {
		return errors.Trace(err)
	}

	ownerName := c.String("owner")
	flowName := args[0]
	flowPath := fmt.Sprintf("%s/flows/%s/%s", home, ownerName, flowName)

	return errors.Trace(os.Remove(flowPath + "/cmd"))
}
