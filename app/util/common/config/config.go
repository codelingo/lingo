package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"gopkg.in/yaml.v1"
)

const (
	DefaultsCfgFile = "defaults.yaml"
	ServicesCfgFile = "services.yaml"
	PlatformCfgFile = "platform.yaml"
	VersionCfgFile  = "version.yaml"
	AuthCfgFile     = "auth.yaml"
	EnvCfgFile 	= "lingo-current-env"
)

// Load assumes cfgFilename is relative to $LINGO_HOME. It loads the config
// data into values.
func Load(cfgFilename string, values interface{}) error {
	cfgPath, err := fullCfgPath(cfgFilename)
	if err != nil {
		return errors.Trace(err)
	}

	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return errors.Errorf("problem reading %s: %v", cfgFilename, err)
	}

	return errors.Annotatef(yaml.Unmarshal(data, values), "problem unmarshalling %s", cfgFilename)

}

// Edit assumes cfgFilename is relative to $LINGO_HOME and opens that
// file with the provided editor.
func Edit(cfgFilename, editor string) error {
	cfgPath, err := fullCfgPath(cfgFilename)
	if err != nil {
		return errors.Trace(err)
	}

	cmd, err := util.OpenFileCmd(editor, cfgPath, 0)
	if err != nil {
		return errors.Trace(err)
	}

	if err = cmd.Start(); err != nil {
		return errors.Trace(err)
	}
	if err = cmd.Wait(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func fullCfgPath(filename string) (string, error) {
	cfgHome, err := util.ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}
	return filepath.Join(cfgHome, filename), nil
}
