package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"runtime"
	"os/exec"
)

func init() {
	register(&cli.Command{
		Name:   "hub",
		Usage:  "Create a .lingo file in the current directory.",
		Action: hubAction,
	}, false)
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
