package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/app/util/common/config"
	servConf "github.com/codelingo/lingo/service/config"
	"fmt"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"strings"
	"os"
	"path/filepath"
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

	configDefaults, err := util.ConfigDefaults()
	if err != nil {
		util.OSErr(err)
		return
	}
	configDefaults = filepath.Join(configDefaults, common.ClientVersion)

	err = createConfigDefaultFiles(configDefaults)
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
	err = createConfigDefaultFiles(configDefaults)
	if err != nil {
		util.OSErr(err)
		return
	}

	reset := ctx.Bool("reset-configs")
	err = updateClientConfigs(configDefaults, reset)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func updateClientConfigs(configDefaults string, reset bool) error {

	versCfg, err := config.Version()
	if err != nil {
		return errors.Trace(err)
	}

	versionUpdated, err := versCfg.ClientVersionUpdated()
	if err != nil {
		return errors.Trace(err)
	}

	if versionUpdated == common.ClientVersion && !reset {
		fmt.Printf("Your client & configs are already on the latest version (%v).\n", common.ClientVersion)
		// TODO:(emersonwood) Does anything more need to happen here? ie. should the user be prompted to update anyway or made aware of `lingo update --reset-configs`?
		return nil
	}

	configUpdates, err := util.ConfigUpdates()
	if err != nil {
		return errors.Trace(err)
	}

	err = createConfigUpdateFiles(configUpdates)
	if err != nil {
		return errors.Trace(err)
	}

	// Add all valid user key:values to update configs
	if !reset {
		authDump, platDump, versDump, err := dumpCurrentConfigs()
		if err != nil {
			return errors.Trace(err)
		}

		authUpdateCfg, err := config.AuthInDir(configUpdates)
		if err != nil {
			return errors.Trace(err)
		}
		authDefaultCfg, err := config.AuthInDir(configDefaults)
		if err != nil {
			return errors.Trace(err)
		}
		err = mergeConfigs(authUpdateCfg.FileConfig, authDefaultCfg.FileConfig, authDump)
		if err != nil {
			return errors.Trace(err)
		}

		platUpdateCfg, err := config.PlatformInDir(configUpdates)
		if err != nil {
			return errors.Trace(err)
		}
		platDefaultCfg, err := config.PlatformInDir(configDefaults)
		if err != nil {
			return errors.Trace(err)
		}
		err = mergeConfigs(platUpdateCfg.FileConfig, platDefaultCfg.FileConfig, platDump)
		if err != nil {
			return errors.Trace(err)
		}

		versUpdateCfg, err := config.VersionInDir(configUpdates)
		if err != nil {
			return errors.Trace(err)
		}
		versDefaultCfg, err := config.VersionInDir(configDefaults)
		err = mergeConfigs(versUpdateCfg.FileConfig, versDefaultCfg.FileConfig, versDump)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err = replaceConfigFiles(configUpdates)
	if err != nil {
		return errors.Trace(err)
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

func dumpCurrentConfigs() (map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	authCfg, err := config.Auth()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	authDump, err := authCfg.Dump()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	platCfg, err := config.Platform()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	platDump, err := platCfg.Dump()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	versCfg, err := config.Version()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	versDump, err := versCfg.Dump()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	return authDump, platDump, versDump, nil
}

func createConfigDefaultFiles(dir string) error {
	err := config.CreateAuthFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreatePlatformFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreateVersionFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func createConfigUpdateFiles(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create update configs directory")
		}
	}

	err := config.CreateAuthFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreatePlatformFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}
	err = config.CreateVersionFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func mergeConfigs(currCfg, oldDefCfg *servConf.FileConfig, keyMap map[string]interface{}) error {
	for k, v := range keyMap {
		keyList := strings.Split(k, ".")
		if len(keyList) < 2 {
			// Not useful.. ignore
			continue
		}
		env := keyList[0]
		key := strings.Join(keyList[1:], ".")

		defVal, err := oldDefCfg.GetForEnv(env, key)
		if err != nil && !strings.HasPrefix(err.Error(), "Could not find value") {
			fmt.Println(err)
			continue
		}

		if v != defVal {
			err = currCfg.SetForEnv(env, key, v)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

	}
	return nil
}

func replaceConfigFiles(configUpdates string) error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	authUpdate := filepath.Join(configUpdates, config.AuthCfgFile)
	authHome := filepath.Join(configHome, config.AuthCfgFile)
	err = os.Rename(authUpdate, authHome)
	if err != nil {
		return errors.Trace(err)
	}

	platUpdate := filepath.Join(configUpdates, config.PlatformCfgFile)
	platHome := filepath.Join(configHome, config.PlatformCfgFile)
	err = os.Rename(platUpdate, platHome)
	if err != nil {
		return errors.Trace(err)
	}

	versUpdate := filepath.Join(configUpdates, config.VersionCfgFile)
	versHome := filepath.Join(configHome, config.VersionCfgFile)
	err = os.Rename(versUpdate, versHome)
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Remove(configUpdates)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
