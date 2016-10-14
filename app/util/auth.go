package util

import (
	"path/filepath"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/service/config"
)

func CreateAuthConfig() (*config.Config, error) {
	// if config exists, return it
	if cfg, err := AuthConfig(); err == nil {
		return cfg, nil
	}

	cfgPath, err := authCfgPath()
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg, err := config.Create(cfgPath, nil, 0755)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err := cfg.Set("all.gitserver.credentials_filename", "git-credentials"); err != nil {
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if err := cfg.Set("dev.gitserver.credentials_filename", "git-credentials-dev"); err != nil {
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if err := cfg.Set("test.gitserver.credentials_filename", "git-credentials-test"); err != nil {
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return cfg, nil
}

// TODO(waigani) move this to util/common/config. Follow the platform config example.
func AuthConfig() (*config.Config, error) {
	cfgPath, err := authCfgPath()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return config.New(cfgPath)
}

// func AssertAuthConfigExists() error {
// 	cfgPath, err := authCfgPath()
// 	if err != nil {
// 		return errors.Trace(err)
// 	}

// 	_, err := os.Stat(cfgPath)
// 	return err
// }

func authCfgPath() (string, error) {
	hm, err := ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	return filepath.Join(hm, "auth.yaml"), nil
}
