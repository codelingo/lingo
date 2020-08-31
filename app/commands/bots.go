package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"github.com/urfave/cli"
)

func init() {
	register(&cli.Command{
		Name:   "bots",
		Usage:  "List Bots",
		Action: listBotsAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  util.InstalledFlg.String(),
				Usage: "List Bots used in current project",
			},
			cli.StringFlag{
				Name:  util.OwnerFlg.String(),
				Usage: "List all Bots of the given owner",
			},
			cli.StringFlag{
				Name:  util.NameFlg.String(),
				Usage: "Describe the named Bot",
			},
		},
	}, false, true)
}

func listBotsAction(ctx *cli.Context) {
	err := listBots(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func listBots(ctx *cli.Context) error {
	owner := ctx.String("owner")
	name := ctx.String("name")

	baseBotURL := baseDiscoveryURL + "bots"
	url := baseBotURL + "/lingo_bots.yaml"
	switch {
	case name != "":

		if owner == "" {
			return errors.New("owner flag must be set")
		}

		url = fmt.Sprintf("%s/%s/%s/lingo_bot.yaml",
			baseBotURL, owner, name)

	case owner != "":
		url = fmt.Sprintf("%s/%s/lingo_owner.yaml",
			baseBotURL, owner)
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
