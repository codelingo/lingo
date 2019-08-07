package commands

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/vcs"
	"github.com/juju/errors"

	"os"
	"path/filepath"
)

func init() {
	register(&cli.Command{
		Name:   "list",
		Usage:  "List all the installed Actions and Tenets that can be found from the given directory.",
		Action: listAllAction,
		Subcommands: []cli.Command{
			{
				Name:   "actions",
				Usage:  "Only list the installed Actions.",
				Action: listLocalFlowsAction,
			},
			{
				Name:   "tenets",
				Usage:  "Only list tenets that can be found from the given directory.",
				Action: listLocalTenetsAction,
			},
		},
	}, false, false, verify.VCSRq)
}

func listAllAction(ctx *cli.Context) {
	str, err := listAll(ctx)
	if err != nil {
		util.FatalOSErr(err)
	}
	fmt.Println(str)
}

func listAll(c *cli.Context) (string, error) {
	flows, err := listLocalFlows(c)
	if err != nil {
		return "", errors.Trace(err)
	}
	tenets, err := listLocalTenets(c)
	if err != nil {
		return "", errors.Trace(err)
	}
	allStr := flows + "\n" + tenets
	return allStr, nil
}

func listLocalFlowsAction(ctx *cli.Context) {
	str, err := listLocalFlows(ctx)
	if err != nil {
		util.FatalOSErr(err)
	}
	fmt.Println(str)
}

func listLocalFlows(c *cli.Context) (string, error) {
	str := `Actions:
  - review
  - search`
	return str, nil
}

func listLocalTenetsAction(ctx *cli.Context) {
	str, err := listLocalTenets(ctx)
	if err != nil {
		util.FatalOSErr(err)
	}
	fmt.Println(str)
}

func listLocalTenets(c *cli.Context) (string, error) {
	_, repo, err := vcs.New()
	if err != nil {
		return "", errors.Trace(err)
	}

	dir := ""
	if len(c.Args()) > 0 {
		dir, err = filepath.Abs(c.Args()[0])
		if err != nil {
			return "", errors.Trace(err)
		}
	} else {
		dir, err = os.Getwd()
		if err != nil {
			return "", errors.Trace(err)
		}
	}

	dls, err := repo.GetDotlingoFilepathsInDir(dir)
	if err != nil {
		return "", errors.Trace(err)
	}

	// TODO: need parse bot to get individual tenet names
	tenetsStr := "Tenets:"
	for _, dl := range dls {
		tenetsStr += fmt.Sprintf("\n  - %s", dl)
	}

	return tenetsStr, nil
}
