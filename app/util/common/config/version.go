package config

import (
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"time"
	"fmt"
	"path/filepath"
	"github.com/codelingo/lingo/app/util"
	"io/ioutil"
	"os"
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

func CreateVersion(overwrite bool) error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	vCfgFilePath := filepath.Join(configHome, VersionCfgFile)
	if _, err := os.Stat(vCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(vCfgFilePath, []byte(VersionTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create version config")
		}
	}

	return nil
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
	return v.SetForEnv("client.version_updated", version, "all")
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
