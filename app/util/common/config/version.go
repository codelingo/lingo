package config

import (
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"path/filepath"
	"github.com/codelingo/lingo/app/util"
)

type versionConfig struct {
	*config.FileConfig
}

func Version() (*versionConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	vCfgPath, err := fullCfgPath(VersionCfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	vCfg, err := cfg.New(vCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &versionConfig{
		vCfg,
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
