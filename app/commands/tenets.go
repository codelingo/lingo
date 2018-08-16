package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"io/ioutil"
	"net/http"
)

func init() {
	register(&cli.Command{
		Name:   "tenets",
		Usage:  "List Tenets",
		Action: listTenetsAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  util.InstalledFlg.String(),
				Usage: "Only show installed Tenets",
			},
			cli.StringFlag{
				Name:  util.OwnerFlg.String(),
				Usage: "List all Tenets of the given owner",
			},
			cli.StringFlag{
				Name:  util.NameFlg.String(),
				Usage: "Describe the named Tenets",
			},
			cli.StringFlag{
				Name:  util.BundleFlg.String(),
				Usage: "List all Tenets of the given bundle",
			},
		},
	}, false, true)
}

func listTenetsAction(ctx *cli.Context) {
	err := listTenets(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func listTenets(ctx *cli.Context) error {
	owner := ctx.String("owner")
	bundle := ctx.String("bundle")
	name := ctx.String("name")

	baseTenetURL := baseDiscoveryURL + "tenets"
	url := baseTenetURL + "/lingo_tenets.yaml"
	switch {
	case name != "":

		if owner == "" {
			return errors.New("owner flag must be set")
		}

		if bundle == "" {
			return errors.New("bundle flag must be set")
		}
		url = fmt.Sprintf("%s/%s/%s/%s/codelingo.yaml",
			baseTenetURL, owner, bundle, name)
	case bundle != "":
		if owner == "" {
			return errors.New("owner flag must be set")
		}
		url = fmt.Sprintf("%s/%s/%s/lingo_bundle.yaml",
			baseTenetURL, owner, bundle)

	case owner != "":
		url = fmt.Sprintf("%s/%s/lingo_owner.yaml",
			baseTenetURL, owner)
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
