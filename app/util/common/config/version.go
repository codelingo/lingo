package config

import (
	"fmt"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	clientVerLatest      = "client.version_latest"
	clientVerLastChecked = "client.version_last_checked"
	clientVerUpdated     = "client.version_updated"
)

type versionConfig struct {
	*config.FileConfig
}

func VersionInDir(dir string) (*versionConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	vCfgPath := filepath.Join(dir, VersionCfgFile)
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
	return VersionInDir(configHome)
}

func CreateVersionFileInDir(dir string, overwrite bool) error {
	vCfgFilePath := filepath.Join(dir, VersionCfgFile)
	if _, err := os.Stat(vCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(vCfgFilePath, []byte(VersionTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create version config")
		}
	}
	return nil
}

func CreateVersionFile() error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return CreateVersionFileInDir(configHome, false)
}

func (v *versionConfig) Dump() (map[string]interface{}, error) {
	keyMap := make(map[string]interface{})

	var versDumpConsts = []string{
		clientVerLatest,
		clientVerLastChecked,
		clientVerUpdated,
	}

	for _, vCon := range versDumpConsts {
		newMap, err := v.GetAll(vCon)
		if err != nil {
			return nil, errors.Trace(err)
		}
		for k, v := range newMap {
			keyMap[k] = v
		}
	}

	return keyMap, nil
}

func (v *versionConfig) ClientLatestVersion() (string, error) {
	return v.GetValue(clientVerLatest)
}

func (v *versionConfig) SetClientLatestVersion(version string) error {
	return v.Set(clientVerLatest, version)
}

func (v *versionConfig) ClientVersionLastChecked() (string, error) {
	return v.GetValue(clientVerLastChecked)
}

func (v *versionConfig) SetClientVersionLastChecked(timeString string) error {
	return v.Set(clientVerLastChecked, timeString)
}

func (v *versionConfig) ClientVersionUpdated() (string, error) {
	return v.GetValue(clientVerUpdated)
}

func (v *versionConfig) SetClientVersionUpdated(version string) error {
	return v.SetForEnv("paas", clientVerUpdated, version)
}

var VersionTmpl = fmt.Sprintf(`
paas:
  client:
    version_latest: %v
    version_last_checked: %v
    version_updated: %v
`, common.ClientVersion, time.Now().UTC().AddDate(0, 0, -1), common.ClientVersion)[1:]
