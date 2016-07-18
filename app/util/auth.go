package util

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/juju/errors"
)

type authConfig struct {
	CurrentUserToken string `yaml:"currentusertoken"`
}

func ReadAuthConfig() (*authConfig, error) {
	hm, err := LingoHome()
	if err != nil {
		return nil, errors.Trace(err)

	}

	authFile := filepath.Join(hm, "auth")
	b, err := ioutil.ReadFile(authFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := yaml.Unmarshal(b, authConfig); err != nil {
		return err
	}

	return authConfig, nil
}
