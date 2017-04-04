package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/app/util/common/config"
	servConf "github.com/codelingo/lingo/service/config"
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
			/*
			TODO: Implement automatic download & install when client release becomes public.
			1. Prompt user to auto download / install.
			2. Download latest binary into temp folder.
			3. Install binary whilst this client is still running.
			4. Exit this client so new client can be loaded.
			5. Either prompt user to run `lingo update` or do it automatically somehow?
			*/
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

	versCfg, err := config.Version()
	if err != nil {
		return errors.Trace(err)
	}

	versionUpdated, err := versCfg.ClientVersionUpdated()
	if err != nil {
		return errors.Trace(err)
	}

	if versionUpdated == common.ClientVersion {
		fmt.Printf("Your client & configs are already on the latest version (%v).\n", common.ClientVersion)
		// TODO:(emersonwood) Does anything more need to happen here? ie. should the user be prompted to update anyway or made aware of `lingo update --reset-configs`?
		return nil
	}

	// Dump the configs
	authCfg, err := config.Auth()
	if err != nil {
		return errors.Trace(err)
	}
	authDump, err := authCfg.Dump()
	if err != nil {
		return errors.Trace(err)
	}

	platCfg, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}
	platDump, err := platCfg.Dump()
	if err != nil {
		return errors.Trace(err)
	}

	versDump, err := versCfg.Dump()
	if err != nil {
		return errors.Trace(err)
	}

	/*
	TODO: Either store these dumps in a temp place or don't overwrite the base config files until merging is complete.
	If any errors occur before merging is complete then this will leave the configs in a semi-merged state
	that can't be recovered from since the original dumps above will be lost.
	*/

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
		currAuthCfg, err := config.Auth()
		if err != nil {
			return errors.Trace(err)
		}
		oldAuthCfg, err := config.AuthDefault(versionUpdated)
		if err != nil {
			return errors.Trace(err)
		}
		err = mergeConfigs(currAuthCfg.FileConfig, oldAuthCfg.FileConfig, authDump)
		if err != nil {
			return errors.Trace(err)
		}

		currPlatCfg, err := config.Platform()
		if err != nil {
			return errors.Trace(err)
		}
		oldPlatCfg, err := config.PlatformDefault(versionUpdated)
		if err != nil {
			return errors.Trace(err)
		}
		err = mergeConfigs(currPlatCfg.FileConfig, oldPlatCfg.FileConfig, platDump)
		if err != nil {
			return errors.Trace(err)
		}

		currVerCfg, err := config.Version()
		if err != nil {
			return errors.Trace(err)
		}
		oldVerCfg, err := config.VersionDefault(versionUpdated)
		err = mergeConfigs(currVerCfg.FileConfig, oldVerCfg.FileConfig, versDump)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err = versCfg.SetClientVersionUpdated(common.ClientVersion)
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

func mergeConfigs(currCfg, oldCfg *servConf.FileConfig, keyMap map[string]interface{}) error {
	// TODO: Implement
	//for k, v := range keyMap {
	//	kList := strings.Split(k, ".")
	//	if len(kList) < 2 {
	//		// Not useful.. ignore
	//		continue
	//	}
	//	env := kList[0]
	//	key := strings.Join(kList[1:], ".")
	//
	//
	//}
	return nil
}
