package config

import (
"path/filepath"

"github.com/juju/errors"

"github.com/codelingo/lingo/service/config"
	"github.com/codelingo/lingo/app/util"
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

func (a *authConfig) GetGitCredentialsFilename() (string, error) {
	return a.Get("gitserver.credentials_filename")
}

func (a *authConfig) GetGitUserName() (string, error) {
	return a.Get("gitserver.user.username")
}

func (a *authConfig) SetGitUserName(userName string) error {
	return a.Set("gitserver.user.username", userName)
}

func (a *authConfig) GetGitUserPassword() (string, error) {
	return a.Get("gitserver.user.password")
}

func (a *authConfig) SetGitUserPassword(userPassword string) error {
	return a.Set("gitserver.user.password", userPassword)
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
