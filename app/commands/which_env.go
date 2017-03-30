package commands

import (
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	serviceConfig "github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"path/filepath"
	"fmt"
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
	fmt.Println(env)

	return nil
}
