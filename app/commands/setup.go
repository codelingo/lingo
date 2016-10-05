package commands

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	utilConfig "github.com/codelingo/lingo/app/util/common/config"

	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
)

// TODO(waigani) add a quick-start cmd with --dry-run flag that: inits git,
// inits lingo, sets up auth and register's repo with CodeLingo platform.

const (
	gitUsernameCfgPath     = "gitserver.user.username"
	gitUserPasswordCfgPath = "gitserver.user.password"
)

func init() {
	register(&cli.Command{
		Name:   "setup",
		Usage:  "Configure the lingo tool for the current machine",
		Action: setupLingoAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "username",
				Usage: "Set the username of the lingo user",
			},
			cli.StringFlag{
				Name:  "token",
				Usage: "Set the token of the lingo user",
			},
		},
		// TODO(waigani) docs on username and token

	}, false, homeRq, configRq)
}

func setupLingoAction(c *cli.Context) {
	username, err := setupLingo(c)
	if err != nil {
		util.OSErrf(err.Error())
	}

	fmt.Sprintln("CodeLingo user set to %q", username)
}

func setupLingo(c *cli.Context) (string, error) {
	if err := util.MaxArgs(c, 0); err != nil {
		return "", errors.Trace(err)
	}
	// recieve generated token from server
	// open codelingo.io/authclient/<token>
	// user verifies on website
	// recieve username, password and same token from server
	// but for now, we'll grab them from flags.
	username := c.String("username")
	if username == "" {

		fmt.Print("enter username: ")
		fmt.Scanln(&username)
	}
	if username == "" {
		return "", errors.New("username cannot be empty")
		//return "", errors.New("username not set")
	}

	password := c.String("token")
	if password == "" {
		fmt.Print("enter user-token:")
		fmt.Scanln(&password)
	}
	if password == "" {
		return "", errors.New("token cannot be empty")
	}

	// check token matches again
	// store username and password in config

	// Check user is logged into platform. If so, generate token.
	// https://github.com/gogits/go-gogs-client/wiki/Users#create-a-access-token

	// set creds in auth config
	if err := setLingoUser(username, password); err != nil {
		return "", errors.Trace(err)
	}

	cfgDir, err := util.ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	platConfig, err := utilConfig.Platform()
	if err != nil {
		return "", errors.Trace(err)
	}

	gitaddr, err := platConfig.GitServerAddr()
	if err != nil {
		return "", errors.Trace(err)
	}

	authCfg, err := util.AuthConfig()
	// TODO(waigani) don't use string lit here
	credFilename, err := authCfg.Get("gitserver.credentials_filename")
	if err != nil {
		return "", errors.Trace(err)
	}

	// Set creds in github.
	// TODO(waigani) these should all be consts.
	gitCredFile := filepath.Join(cfgDir, credFilename)
	out, err := gitCMD("config",
		"--global", fmt.Sprintf("credential.%s.helper", gitaddr),
		fmt.Sprintf("store --file %s", gitCredFile),
	)
	if err != nil {
		return "", errors.Annotate(err, out)
	}

	cfg, err := util.AuthConfig()
	if err != nil {
		// TODO(waigani) handle different errors
		cfg, err = util.CreateAuthConfig()
		if err != nil {
			return "", errors.Trace(err)
		}
	}

	gitUsername, err := cfg.Get(gitUsernameCfgPath)
	if err != nil {
		return "", errors.Trace(err)
	}
	gitPassword, err := cfg.Get(gitUserPasswordCfgPath)
	if err != nil {
		return "", errors.Trace(err)
	}

	gitprotocol, err := platConfig.GitServerProtocol()
	if err != nil {
		return "", errors.Trace(err)
	}

	githost, err := platConfig.GitServerHost()
	if err != nil {
		return "", errors.Trace(err)
	}

	gitport, err := platConfig.GitServerPort()
	if err != nil {
		return "", errors.Trace(err)
	}

	data := []byte(fmt.Sprintf("%s://%s:%s@%s:%s", gitprotocol, gitUsername, gitPassword, githost, gitport))
	// write creds to config file
	if err := ioutil.WriteFile(gitCredFile, data, 0755); err != nil {
		return "", errors.Trace(err)
	}

	return username, nil
}

func gitCMD(args ...string) (out string, err error) {
	cmd := exec.Command("git", args...)
	b, err := cmd.CombinedOutput()
	out = strings.TrimSpace(string(b))
	return out, errors.Annotate(err, out)
}

// write auth config if it's missing
func getOrCreateAuthConfig() (*config.Config, error) {
	cfg, err := util.AuthConfig()
	if err != nil {
		// TODO(waigani) handle different errors
		return util.CreateAuthConfig()
	}
	return cfg, nil
}

func setLingoUser(username, password string) error {
	cfg, err := getOrCreateAuthConfig()
	if err != nil {
		return errors.Trace(err)
	}

	// TODO(waigani) check if currentuser is already set and abort. Require a --reset flag to reset.
	if err := cfg.Set("all."+gitUsernameCfgPath, username); err != nil {
		return errors.Trace(err)
	}
	return cfg.Set("all."+gitUserPasswordCfgPath, password)

}
