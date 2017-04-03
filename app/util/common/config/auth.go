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

func Auth() (*authConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	aCfgPath, err := fullCfgPath(AuthCfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	aCfg, err := cfg.New(aCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &authConfig{
		aCfg,
	}, nil
}

func createAuthFile(basepath string, overwrite bool) error {
	aCfgFilePath := filepath.Join(basepath, AuthCfgFile)
	if _, err := os.Stat(aCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(aCfgFilePath, []byte(AuthTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create auth config")
		}
	}
	return nil
}

func CreateAuthFile(overwrite bool) error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return createAuthFile(configHome, overwrite)
}

func CreateAuthDefaultFile() error {
	configDefaults, err := util.ConfigDefaults()
	if err != nil {
		return errors.Trace(err)
	}
	return createAuthFile(configDefaults, true)
}

func (a *authConfig) GetGitCredentialsFilename() (string, error) {
	return a.Get(gitCredentialFilename)
}

func (a *authConfig) GetGitUserName() (string, error) {
	return a.Get(gitUserName)
}

func (a *authConfig) SetGitUserName(userName string) error {
	return a.Set(gitUserName, userName)
}

func (a *authConfig) GetGitUserPassword() (string, error) {
	return a.Get(gitPassword)
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
