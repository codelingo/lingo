package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"fmt"
	"os"
)

func init() {
	register(&cli.Command{
		Name:   "which-env",
		Usage:  "Show the current environment",
		Action: whichEnvAction,

	}, false)
}

func whichEnvAction(ctx *cli.Context) {
	err := whichEnv(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func whichEnv(ctx *cli.Context) error {
	configsHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	envCfg := filepath.Join(configsHome, config.EnvCfgFile)

	env, err := ioutil.ReadFile(envCfg)
	if err != nil {
		if strings.Contains(err.Error(), "open /home/dev/.codelingo/configs/lingo-current-env: no such file or directory") {
			return errors.New("No lingo environment set. Please run `lingo use-env <env>` to set the environment.")
		}

		return errors.Trace(err)
	}

	trimmedEnv := strings.TrimSpace(string(env))
	err = outputString(ctx.String("output"), trimmedEnv)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func outputString(output string, content string) error {
	if !strings.HasSuffix(content, "\n") {
		content = content+"\n"
	}

	if output == "" {
		fmt.Print(content)
		return nil
	}

	outputPath, err := getFilePath(output)
	if err != nil {
		return errors.Trace(err)
	}

	if _, err := os.Stat(outputPath); err == nil {
		return errors.Trace(err)
	}

	return errors.Trace(ioutil.WriteFile(outputPath, []byte(content), 0644))
}


