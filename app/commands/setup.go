package commands

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"golang.org/x/crypto/ssh/terminal"

	utilConfig "github.com/codelingo/lingo/app/util/common/config"

	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
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
			cli.BoolFlag{
				Name:  "keep-creds",
				Usage: "Preserves existing credentials (if present)",
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

	fmt.Println("Success! CodeLingo user set to", username)
}

// TODO(waigani) dry run this
func setupLingo(c *cli.Context) (string, error) {

	if err := util.MaxArgs(c, 0); err != nil {
		return "", errors.Trace(err)
	}

	versionConfig, err := utilConfig.Version()
	if err != nil {
		return "", errors.Trace(err)
	}
	if err := versionConfig.SetClientVersion(common.ClientVersion); err != nil {
		return "", errors.Trace(err)
	}

	platConfig, err := utilConfig.Platform()
	if err != nil {
		return "", errors.Trace(err)
	}

	webAddr, err := platConfig.Address()
	if err != nil {
		return "", errors.Trace(err)
	}

	authConfig, authErr := util.AuthConfig()
	if authErr != nil {
		return "", errors.Trace(authErr)
	}

	// recieve generated token from server
	// open codelingo.io/authclient/<token>
	// user verifies on website
	// recieve username, password and same token from server
	// but for now, we'll grab them from flags.
	username := c.String("username")
	password := c.String("token")

	// If --keep-creds is set, attempt to grab username/token
	// from authConfig
	if c.Bool("keep-creds") {
		if username == "" {
			username, authErr = authConfig.Get(gitUsernameCfgPath)
			if authErr != nil {
				if strings.Contains(authErr.Error(), `"username" not found`) {
					username = ""
				} else {
					errors.Annotate(authErr, "setup --keep-creds failed.")
					return "", errors.Trace(authErr)
				}
			}
		}

		if password == "" {
			password, authErr = authConfig.Get(gitUserPasswordCfgPath)
			if authErr != nil {
				if strings.Contains(authErr.Error(), `"password" not found`) {
					password = ""
				} else {
					errors.Annotate(authErr, "setup --keep-creds failed.")
					return "", errors.Trace(authErr)
				}
			}
		}
	}

	if username == "" {
		fmt.Print("Enter Your CodeLingo Username: ")
		fmt.Scanln(&username)
	}
	if username == "" {
		return "", errors.New("username cannot be empty")
		//return "", errors.New("username not set")
	}

	if password == "" {
		// Launch website to gen token
		lingoTokenAddr := "http://" + webAddr + "/lingo-token"
		fmt.Print("Enter User-Token:")
		if err := open.Run(lingoTokenAddr); err != nil {
			return "", errors.Trace(err)
		}
		byt, err := terminal.ReadPassword(0)
		if err != nil {
			return "", errors.Trace(err)
		}
		password = string(byt)
		fmt.Println("")
	}

	if password == "" {
		return "", errors.New("token cannot be empty")
	}

	// TODO(waigani) check token matches again
	// store username and password in config

	// set creds in auth config
	if err := setLingoUser(username, password); err != nil {
		return "", errors.Trace(err)
	}

	authConfig, err = getOrCreateAuthConfig()
	if err != nil {
		return "", errors.Trace(err)
	}

	cfgDir, err := util.ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	// TODO(waigani) don't use string lit here
	credFilename, err := authConfig.Get("gitserver.credentials_filename")
	if err != nil {
		return "", errors.Trace(err)
	}

	gitaddr, err := platConfig.GitServerAddr()
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

	gitUsername, err := authConfig.Get(gitUsernameCfgPath)
	if err != nil {
		return "", errors.Trace(err)
	}
	gitPassword, err := authConfig.Get(gitUserPasswordCfgPath)
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
