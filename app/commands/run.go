package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"

	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:            "run",
		Usage:           "Run the given flow in the current directory.",
		Action:          runAction,
		SkipFlagParsing: true,
	}, false, false, verify.VersionRq)
}

func runAction(ctx *cli.Context) {
	if err := run(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
}

func run(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("Failed to run flow - no flow given.")
	}

	flowName := args[0]
	cmdPath, err := findInstalledCmd(flowName)
	if err != nil {
		return errors.Trace(err)
	}

	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		return errors.Errorf("Flow %[1]q not found. Try installing it with `lingo install %[1]s`", flowName)
	}

	var strArgs []string
	for _, arg := range args[1:] {
		strArgs = append(strArgs, arg)
	}

	return errors.Trace(runFlowCmd(cmdPath, strArgs))
}

func runFlowCmd(flowCmd string, args []string) error {
	cmd := exec.Command(flowCmd, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return errors.Trace(cmd.Run())
}

func findInstalledCmd(name string) (string, error) {

	// TODO(waigani) we should cycle through all installed owners if owner is
	// not specified.
	owner := "codelingo"

	parts := strings.Split(name, "/")
	switch {
	case len(parts) == 2:
		owner = parts[0]
		name = parts[1]
	case len(parts) > 2:
		return "", errors.Errorf("%q is not a valid Flow name", name)
	}

	home, err := util.LingoHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("%s/flows/%s/%s/cmd%s", home, owner, name, ext), nil
}
