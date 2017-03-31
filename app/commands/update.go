package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/app/util/common/config"
	"fmt"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"strings"
)

func init() {

	register(&cli.Command{
		Name:   "update",
		Usage:  "Update the lingo client to the latest release.",
		Flags:  []cli.Flag{
			cli.BoolFlag{
				Name: "reset",
				Usage: "Replace client configs with defaults.",
			},
			cli.BoolFlag{
				Name: "no-prompt",
				Usage: "Won't prompt to update client configs.",
			},
		},
		Action: updateAction,
	},
		false,
		homeRq,
		authRq,
	)

	// 	register(&cli.Command{
	// 	Name:  "update",
	// 	Usage: "update the lingo client to the latest release",
	// 	Flags: []cli.Flag{
	// 		cli.BoolFlag{
	// 			Name:  "dry-run",
	// 			Usage: "prints the update steps without preforming them",
	// 		},
	// 		cli.BoolFlag{
	// 			Name:  "check",
	// 			Usage: "checks if a newer version is available",
	// 		},
	// 	},
	// 	Action: updateAction,
	// },
	// 	false,
	// 	homeRq,
	// )
}

func updateAction(ctx *cli.Context) {

	// Check version against endpoint
	outdated, err := VersionIsOutdated()
	if err != nil {
		if outdated {
			fmt.Println("Your client is out of date. Please download and install the latest binary.")
			return
		} else {
			util.OSErr(err)
			return
		}
	}

	// Check version updated
	vCfg, err := config.Version()
	if err != nil {
		util.OSErr(err)
		return
	}

	verUpdtd, err := vCfg.ClientVersionUpdated()
	if err != nil {
		util.OSErr(err)
		return
	}

	compare, err := compareVersions(common.ClientVersion, verUpdtd)
	if err != nil {
		util.OSErr(err)
		return
	}

	reset := ctx.Bool("reset")
	if compare != 0 || reset {
		err := updateClientConfigs(reset)
		if err != nil {
			util.OSErr(err)
			return
		}
	} else {
		noPrompt := ctx.Bool("no-prompt")
		if !noPrompt {
			shouldUpdate := ""
			fmt.Print("Are you sure you want to update your client configs? (y/n): ")
			fmt.Scanln(&shouldUpdate)

			shouldUpdate = strings.ToLower(shouldUpdate)
			if shouldUpdate != "y" && shouldUpdate != "yes" {
				fmt.Println("Update aborted.")
				return
			}
		}

		err := updateClientConfigs(false)
		if err != nil {
			util.OSErr(err)
			return
		}
	}

	// fmt.Println("DISABLED: the update command will be enabled once the codelingo/lingo repository is public")

	// // first check if an update is avaliable
	// v, err := semver.Make(common.ClientVersion)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// 	// errors.ErrorStack(err)
	// }

	// latest, err := latestVersion()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// 	// errors.ErrorStack(err)
	// }

	// if v.GT(latest) {
	// 	// no new versions available
	// 	return
	// }

	// CONTINUE HERE:
	// 1. detect local OS
	// 2. download new binary
	// 3. run upgrade steps from new binary
}

func updateClientConfigs(reset bool) error {
	vCfg, err := config.Version()
	if err != nil {
		return errors.Trace(err)
	}

	if reset {
		fmt.Println("Resetting configs to defaults.")
	} else {
		fmt.Println("Updating configs.")
	}
	fmt.Println("Update complete... Setting version updated.")



	err = vCfg.SetClientVersionUpdated(common.ClientVersion)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println("Update success!")

	return nil
}
