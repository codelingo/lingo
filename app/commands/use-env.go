package commands

import (
	"fmt"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	serviceConfig "github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
)

const (
	websiteAddr      = "website.addr"
	platformAddr     = "platform.addr"
	flowAddr         = "flow.address"
	gitServerAddr    = "gitserver.remote.host"
	p4ServerAddr     = "p4server.remote.host"
	messagequeueAddr = "messagequeue.address.host"
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
	if newEnv == "onprem" {
		ip := ""
		fmt.Print("Enter the Platform IP: ")
		fmt.Scanln(&ip)
		cfg, err := commonConfig.Platform()
		if err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(websiteAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(platformAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(flowAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(gitServerAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(p4ServerAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(messagequeueAddr, ip); err != nil {
			return errors.Trace(err)
		}
	}
	fmt.Printf("Success! Environment set to '%v'.\n", newEnv)

	return nil
}
