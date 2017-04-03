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
		configRq,
	)
}

func updateAction(ctx *cli.Context) {

	// Write pre-update client defaults to CLHOME/configs/defaults/<version>/*.yaml
	err := writeConfigDefaults()
	if err != nil {
		util.OSErr(err)
		return
	}

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

	// Write post-update client defaults to CLHOME/configs/defaults/<version>/*.yaml
	err = writeConfigDefaults()
	if err != nil {
		util.OSErr(err)
		return
	}

	reset := ctx.Bool("reset-configs")
	err = updateClientConfigs(reset)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func writeConfigDefaults() error {
	err := config.CreateAuthDefaultFile()
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreatePlatformDefaultFile()
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreateVersionDefaultFile()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
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
		fmt.Printf("Your client & configs are already on the latest version (%v).\n", common.ClientVersion)
		// TODO:(emersonwood) Does anything more need to happen here? ie. should the user be prompted to update anyway or made aware of `lingo update --reset-configs`?
		return nil
	}

	// TODO:(emersonwood) Store old configs in a struct here

	// Overwrite existing configs with new client config templates
	err = config.CreateAuthFile(true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreatePlatformFile(true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreateVersionFile(true)
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
