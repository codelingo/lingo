package commands

import (
	"fmt"

	"github.com/blang/semver"

	"github.com/codegangsta/cli"
)

func init() {
	// TODO(waigani) support upgrades
	// TODO(waigani) support specific versions
	register(&cli.Command{
		Name:  "update",
		Usage: "update the lingo client to the latest release",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "prints the update steps without preforming them",
			},
			cli.BoolFlag{
				Name:  "check",
				Usage: "checks if a newer version is available",
			},
		},
		Action: updateAction,
	},
		homeRq,
	)
}

func updateAction(ctx *cli.Context) {

	fmt.Println("DISABLED: the update command will be enabled once the codelingo/lingo repository is public")

	// first check if an update is avaliable
	v, err := semver.Make(ClientVersion)
	if err != nil {
		fmt.Println(err.Error())
		return
		// errors.ErrorStack(err)
	}

	latest, err := latestVersion()
	if err != nil {
		fmt.Println(err.Error())
		return
		// errors.ErrorStack(err)
	}

	if v.GT(latest) {
		// no new versions available
		return
	}

	// CONTINUE HERE:
	// 1. detect local OS
	// 2. download new binary
	// 3. run upgrade steps from new binary
}

// TODO(waigani) once repo is public, check github API for latest release:
// https://api.github.com/repos/codelingo/lingo/releases/latest
func latestVersion() (semver.Version, error) {
	return semver.Make("1.2.3")
}
