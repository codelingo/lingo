package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/codelingo/lingo/app/util"

	"github.com/codegangsta/cli"
	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "actions",
		Usage:  "List Actions",
		Action: listFlowsAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  util.InstalledFlg.String(),
				Usage: "Only show installed Actions",
			},
			cli.StringFlag{
				Name:  util.OwnerFlg.String(),
				Usage: "List all Actions of the given owner",
			},
			cli.StringFlag{
				Name:  util.NameFlg.String(),
				Usage: "Describe the named Actions",
			},
		},
	}, false, true)
}

func listFlowsAction(ctx *cli.Context) {
	err := listFlows(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func listFlows(ctx *cli.Context) error {
	owner := ctx.String("owner")
	name := ctx.String("name")

	baseFlowURL := baseDiscoveryURL + "flows"
	url := baseFlowURL + "/lingo_flows.yaml"
	switch {
	case name != "":

		if owner == "" {
			return errors.New("owner flag must be set")
		}

		url = fmt.Sprintf("%s/%s/%s/lingo_flow.yaml",
			baseFlowURL, owner, name)

	case owner != "":
		url = fmt.Sprintf("%s/%s/lingo_owner.yaml",
			baseFlowURL, owner)
	}
	resp, err := http.Get(url)
	if err != nil {
		return errors.Trace(err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println(string(data))
	return nil
}
