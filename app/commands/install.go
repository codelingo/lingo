package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"

	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "install",
		Usage:  "Install a Flow",
		Action: installAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "owner",
				Value: "codelingo",
				Usage: "Owner of the Flow",
			},
		},
	}, false, false)
}

func installAction(ctx *cli.Context) {
	if err := install(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Printf("Success! You can now run the installed flow with `lingo run %s`\n", ctx.Args()[0])
}

func install(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("Failed to install Flow - no Flow given.")
	}

	home, err := util.LingoHome()
	if err != nil {
		return errors.Trace(err)
	}

	ownerName := c.String("owner")
	flowName := args[0]
	flowPath := fmt.Sprintf("%s/flows/%s/%s", home, ownerName, flowName)

	if err := os.MkdirAll(flowPath, 0755); err != nil {
		return errors.Trace(err)
	}

	fileUrl := fmt.Sprintf("https://github.com/codelingo/codelingo/raw/master/flows/%s/%s/cmd", ownerName, flowName)
	return errors.Trace(DownloadFile(flowPath+"/cmd", fileUrl))
}

func DownloadFile(filepath string, url string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return errors.Trace(os.Chmod(filepath, 0755))

}
