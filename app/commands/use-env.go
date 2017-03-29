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
		Name:   "use-env",
		Usage:  "Use the given environment.",
		Action: useEnvAction,

	}, false, homeRq, configRq)
}

func useEnvAction(ctx *cli.Context) {
	err := useEnv(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func useEnv(ctx *cli.Context) error {
	var err error
	switch len(ctx.Args()) {
	case 0:
		err = errors.New("Error: An environment value must be specified: `lingo use-env <env>`")
		return err
	case 1:
		// Success case
		break
	default:
		err = errors.New("Error: Only 1 environment value can be specified: `lingo use-env <env>`")
		return err
	}

	configsHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	newEnv := ctx.Args()[0]
	envFilepath := filepath.Join(configsHome, commonConfig.EnvCfgFile)

	cfg := serviceConfig.New(envFilepath)
	err = cfg.SetEnv(newEnv)
	if err != nil {
		return errors.Trace(err)
	}

	success := fmt.Sprintf("Success! Environment set to '%v'.", newEnv)
	err = outputString(ctx.String("output"), success)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}


