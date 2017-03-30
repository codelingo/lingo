package config

import (
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"time"
	"fmt"
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

func (v *versionConfig) ClientLatestVersion() (string, error) {
	return v.Get("client.version_latest")
}

func (v *versionConfig) SetClientLatestVersion(version string) error {
	return v.Set("client.version_latest", version)
}

func(v *versionConfig) ClientVersionLastChecked() (string, error) {
	return v.Get("client.version_last_checked")
}

func (v *versionConfig) SetClientVersionLastChecked(timeString string) error {
	return v.Set("client.version_last_checked", timeString)
}

func (v *versionConfig) ClientVersionUpdated() (string, error) {
	return v.Get("client.version_updated")
}

func (v *versionConfig) SetClientVersionUpdated(version string) error {
	return v.Set("client.version_updated", version)
}

func (v *versionConfig) ClientVersion() (string, error) {
	return v.Get("client.version")
}

func (v *versionConfig) SetClientVersion(cv string) error {
	// TODO: Use `hashicorp/go-version` package for comparing and setting semvers
	// https://github.com/hashicorp/go-version
	return v.Set("all.client.version", cv)
}

var VersionTmpl = fmt.Sprintf(`
all:
  client:
    version_latest: %v
    version_last_checked: %v
    version_updated: %v
`, common.ClientVersion, time.Now().UTC().AddDate(0, 0, -1), common.ClientVersion)[1:]
