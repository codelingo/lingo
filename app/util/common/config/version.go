package config

import (
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
)

type versionConfig struct {
	*config.Config
}

func Version() (*versionConfig, error) {
	cfgPath, err := fullCfgPath(VersionCfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg, err := config.New(cfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &versionConfig{
		Config: cfg,
	}, nil
}

func (v *versionConfig) ClientVersion() (string, error) {
	return v.Get("client.version")
}

func (v *versionConfig) SetClientVersion(cv string) error {
	// TODO: Use `hashicorp/go-version` package for comparing and setting semvers
	// https://github.com/hashicorp/go-version
	return v.Set("all.client.version", cv)
}

var VersionTmpl = `
all:
  client:
    version: `[1:] + common.ClientVersion + `
`
