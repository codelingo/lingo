package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"os/exec"
	"runtime"
)

func init() {
	register(&cli.Command{
		Name:   "hub",
		Usage:  "Opens the CodeLingo Hub in your default browser.",
		Action: hubAction,
	}, false, true)
}

func hubAction(ctx *cli.Context) {
	if err := hub(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
}

func hub(c *cli.Context) error {
	return openBrowser("https://codelingo.io/hub")
}

func openBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	err := cmd.Start()
	return errors.Trace(err)
}
