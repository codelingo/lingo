package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/app/util/common/config"
	"fmt"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

func init() {

	register(&cli.Command{
		Name:   "update",
		Usage:  "Update the lingo client to the latest release.",
		Flags:  []cli.Flag{
			cli.BoolFlag{
				Name: "reset-configs",
				Usage: "Replace client configs with defaults.",
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
			fmt.Println("Your client is out of date. Please download and install the latest binary from https://github.com/codelingo/lingo/releases")
			return
		} else {
			util.OSErr(err)
			return
		}
	}

	reset := ctx.Bool("reset-configs")
	err = updateClientConfigs(reset)
	if err != nil {
		util.OSErr(err)
		return
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

	vrsnUpdtd, err := vCfg.ClientVersionUpdated()
	if err != nil {
		return errors.Trace(err)
	}

	if vrsnUpdtd == common.ClientVersion {
		fmt.Printf("Your client & configs have already been updated to the latest version (%v).\n", common.ClientVersion)
		// TODO:(emersonwood) Does anything more need to happen here? ie. should the user be prompted to update anyway or made aware of `lingo update --reset-configs`?
		return nil
	}

	// TODO:(emersonwood) Store old configs in a struct here

	// Overwrite existing configs with store templates
	err = config.CreateAuth(true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreatePlatform(true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreateVersion(true)
	if err != nil {
		return errors.Trace(err)
	}

	if !reset {
		// TODO:(emersonwood) Restore values from old configs; discuss with Jesse
	}

	err = vCfg.SetClientVersionUpdated(common.ClientVersion)
	if err != nil {
		return errors.Trace(err)
	}

	if reset {
		fmt.Println("Configs reset successfully.")
	} else {
		fmt.Println("Configs updated successfully.")
	}

	return nil
}
