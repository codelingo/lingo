package commands

import (
	"flag"

	"github.com/blang/semver"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

func init() {
	// TODO(waigani) support upgrades
	// TODO(waigani) support specific versions

	// NOTE: temporary solution – `lingo update` ⟼ `lingo setup --keed-creds`
	register(&cli.Command{
		Name:   "update",
		Usage:  "Update the lingo client to the latest release.",
		Flags:  []cli.Flag{},
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

	// NOTE: temporary solution – `lingo update` ⟼ `lingo setup --keep-creds`
	fset := flag.NewFlagSet("update aliased to setup keep-creds", flag.ContinueOnError)
	f := cli.BoolFlag{
		Name:  "keep-creds",
		Usage: "Preserves existing credentials (if present)",
	}
	f.Apply(fset)
	ctx = cli.NewContext(ctx.App, fset, ctx.Parent())
	ctx.Set("keep-creds", "true")
	setupLingoAction(ctx)

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

func latestVersion() (*semver.Version, error) {
	svc, err := service.New()
	if err != nil {
		return nil, errors.Trace(err)
	}

	versionString, err := svc.LatestClientVersion()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return semver.New(versionString)
}
