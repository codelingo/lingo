package config

import (
	"path/filepath"
	"github.com/juju/errors"
	"github.com/codelingo/lingo/service/config"
	"github.com/codelingo/lingo/app/util"
	"io/ioutil"
	"os"
)

const(
	gitCredentialFilename = "gitserver.credentials_filename"
	gitUserName = "gitserver.user.username"
	gitPassword = "gitserver.user.password"
)

type authConfig struct {
	*config.FileConfig
}

func AuthInDir(dir string) (*authConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	aCfgPath := filepath.Join(dir, AuthCfgFile)
	aCfg, err := cfg.New(aCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &authConfig{
		aCfg,
	}, nil
}

func Auth() (*authConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return AuthInDir(configHome)
}

func CreateAuthFileInDir(dir string, overwrite bool) error {
	aCfgFilePath := filepath.Join(dir, AuthCfgFile)
	if _, err := os.Stat(aCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(aCfgFilePath, []byte(AuthTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create auth config")
		}
	}
	return nil
}

func CreateAuthFile() error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return CreateAuthFileInDir(configHome, false)
}

func (a *authConfig) Dump() (map[string]interface{}, error) {
	keyMap := make(map[string]interface{})

	var authDumpConsts = []string{
		gitCredentialFilename,
		gitUserName,
		gitPassword,
	}

	for _, aCon := range authDumpConsts {
		newMap, err := a.GetAll(aCon)
		if err != nil {
			return nil, errors.Trace(err)
		}
		for k, v := range newMap {
			keyMap[k] = v
		}
	}

	return keyMap, nil
}

func (a *authConfig) GetGitCredentialsFilename() (string, error) {
	return a.GetValue(gitCredentialFilename)
}

func (a *authConfig) GetGitUserName() (string, error) {
	return a.GetValue(gitUserName)
}

func (a *authConfig) SetGitUserName(userName string) error {
	return a.Set(gitUserName, userName)
}

func (a *authConfig) GetGitUserPassword() (string, error) {
	return a.GetValue(gitPassword)
}

func (a *authConfig) SetGitUserPassword(userPassword string) error {
	return a.Set(gitPassword, userPassword)
}


var AuthTmpl = `
all:
  gitserver:
    credentials_filename: git-credentials
    user:
      password: ""
      username: ""
dev:
  gitserver:
    credentials_filename: git-credentials-dev
onprem:
  gitserver:
    credentials_filename: git-credentials-onprem
test:
  gitserver:
    credentials_filename: git-credentials-test`[1:]
