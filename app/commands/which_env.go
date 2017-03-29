package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	serviceConfig "github.com/codelingo/lingo/service/config"
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
		Usage:  "Show the current environment.",
		Action: whichEnvAction,

	}, false, homeRq, configRq)
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

	envFilepath := filepath.Join(configsHome, commonConfig.EnvCfgFile)
	cfg := serviceConfig.New(envFilepath)

	env, err := cfg.GetEnv()
	if err != nil {
		return errors.Trace(err)
	}

	err = outputString(ctx.String("output"), env)
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


