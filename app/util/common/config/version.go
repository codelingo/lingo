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

const (
	clientVerLatest = "client.version_latest"
	clientVerLastChecked = "client.version_last_checked"
	clientVerUpdated = "client.version_updated"
)

type versionConfig struct {
	*config.FileConfig
}

func version(basepath string) (*versionConfig, error) {
	envFile := filepath.Join(basepath, EnvCfgFile)
	cfg := config.New(envFile)

	vCfgPath := filepath.Join(basepath, VersionCfgFile)
	vCfg, err := cfg.New(vCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &versionConfig{
		vCfg,
	}, nil
}

func Version() (*versionConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return version(configHome)
}

func VersionDefault(ver string) (*versionConfig, error) {
	configDefaults, err := util.ConfigDefaults()
	if err != nil {
		return nil, errors.Trace(err)
	}
	dir := filepath.Join(configDefaults, ver)
	return version(dir)
}

func createVersionFile(basepath string, overwrite bool) error {
	vCfgFilePath := filepath.Join(basepath, VersionCfgFile)
	if _, err := os.Stat(vCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(vCfgFilePath, []byte(VersionTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create version config")
		}
	}
	return nil
}

func CreateVersionFile(overwrite bool) error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return createVersionFile(configHome, overwrite)
}

func CreateVersionDefaultFile() error {
	configDefaults, err := util.ConfigDefaults()
	if err != nil {
		return errors.Trace(err)
	}
	return createVersionFile(configDefaults, true)
}

func (v *versionConfig) Dump() (map[string]interface{}, error) {
	keyMap := make(map[string]interface{})
	// TODO: Implement
	return keyMap, nil
}

func (v *versionConfig) ClientLatestVersion() (string, error) {
	return v.Get(clientVerLatest)
}

func (v *versionConfig) SetClientLatestVersion(version string) error {
	return v.Set(clientVerLatest, version)
}

func(v *versionConfig) ClientVersionLastChecked() (string, error) {
	return v.Get(clientVerLastChecked)
}

func (v *versionConfig) SetClientVersionLastChecked(timeString string) error {
	return v.Set(clientVerLastChecked, timeString)
}

func (v *versionConfig) ClientVersionUpdated() (string, error) {
	return v.Get(clientVerUpdated)
}

func (v *versionConfig) SetClientVersionUpdated(version string) error {
	return v.SetForEnv("all", clientVerUpdated, version)
}

var VersionTmpl = fmt.Sprintf(`
all:
  client:
    version_latest: %v
    version_last_checked: %v
    version_updated: %v
`, common.ClientVersion, time.Now().UTC().AddDate(0, 0, -1), common.ClientVersion)[1:]
