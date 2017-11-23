package commands

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	utilConfig "github.com/codelingo/lingo/app/util/common/config"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"
	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

// TODO(waigani) add a quick-start cmd with --dry-run flag that: inits git,
// inits lingo, sets up auth and register's repo with CodeLingo platform.

func init() {
	register(&cli.Command{
		Name:   "setup",
		Usage:  "Configure the lingo tool for the current environment on this machine.",
		Action: setupLingoAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "username",
				Usage: "Set the username of the lingo user.",
			},
			cli.StringFlag{
				Name:  "token",
				Usage: "Set the token of the lingo user.",
			},
			cli.BoolFlag{
				Name:  "keep-creds",
				Usage: "Preserves existing credentials (if present).",
			},
		},
		// TODO(waigani) docs on username and token

	}, false, homeRq, configRq)
}

func setupLingoAction(c *cli.Context) {
	username, err := setupLingo(c)
	if err != nil {
		util.OSErr(err)
	}

	fmt.Println("Success! CodeLingo user set to", username)
}

// TODO(waigani) dry run this
func setupLingo(c *cli.Context) (string, error) {

	if err := util.MaxArgs(c, 0); err != nil {
		return "", errors.Trace(err)
	}

	platConfig, err := utilConfig.Platform()
	if err != nil {
		return "", errors.Trace(err)
	}

	webAddr, err := platConfig.WebSiteAddress()
	if err != nil {
		return "", errors.Trace(err)
	}

	authConfig, err := commonConfig.Auth()
	if err != nil {
		return "", errors.Trace(err)
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
			username, err = authConfig.GetGitUserName()
			if err != nil && !strings.Contains(err.Error(), `"username" not found`) {
				// TODO(waigani) Check error type
				errors.Annotate(err, "setup --keep-creds failed.")
				return "", errors.Trace(err)
			}
		}

		if password == "" {
			password, err = authConfig.GetGitUserPassword()
			if err != nil && !strings.Contains(err.Error(), `"password" not found`) {
				// TODO(waigani) Check error type
				errors.Annotate(err, "setup --keep-creds failed.")
				return "", errors.Trace(err)
			}
		}
	}
	// TODO (Junyu) set proper Perforce Username and password
	// Prompt for user
	if username == "" || password == "" {
		lingoTokenAddr := "http://" + webAddr + "/lingo-token"
		fmt.Println("Please sign in to " + lingoTokenAddr + " to generate a new Token linked with your CodeLingo User account.")
	}

	if username == "" {
		fmt.Print("Enter Your CodeLingo Username: ")
		fmt.Scanln(&username)
	}
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	if password == "" {
		fmt.Print("Enter User-Token:")

		// If pwd can be executed,this is an Unix-like shell (i.e. GitBash)
		// If it runs into an error (i.e. Windows Command Prompt)
		// use terminal instead of gopass package to read password
		cmd := exec.Command("bash", "-c", "/usr/bin/pwd")
		var byt []byte
		if err := cmd.Run(); err != nil{
			byt, err = terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "", errors.Trace(err)
			}
		} else{
			byt, err = gopass.GetPasswd()
			if err != nil {
				return "", errors.Trace(err)
			}
		}

		password = string(byt)
		fmt.Println("")
	}

	if password == "" {
		return "", errors.New("token cannot be empty")
	}

	// TODO(waigani) check token matches again

	// Set username & password in auth.yaml
	if err := authConfig.SetGitUserName(username); err != nil {
		return "", errors.Trace(err)
	}
	if err := authConfig.SetGitUserPassword(password); err != nil {
		return "", errors.Trace(err)
	}

	// TODo (Junyu) md5 is not secure. All communication should be over tls
	p4Password := strings.ToUpper(GetMD5Hash(password))
	if err := authConfig.SetP4UserName(username); err != nil {
		return "", errors.Trace(err)
	}
	if err := authConfig.SetP4UserPassword(p4Password); err != nil {
		return "", errors.Trace(err)
	}

	credFilename, err := authConfig.GetGitCredentialsFilename()
	if err != nil {
		return "", errors.Trace(err)
	}

	gitaddr, err := platConfig.GitServerAddr()
	if err != nil {
		return "", errors.Trace(err)
	}
	cfgDir, err := util.ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	// Set creds in github.
	// TODO(waigani) these should all be consts.
	gitCredFile := filepath.Join(cfgDir, credFilename)
	gitCfgFile := gitCredFile
	if runtime.GOOS == "windows" {
		b, err := exec.Command("pwd").CombinedOutput()
		if err == nil {
			// Potential bug if the username contains "\\" which is unlikely to happen
			gitCredFile = strings.Replace(gitCredFile, "\\", "/", -1)
			gitCfgFile = strings.Replace(gitCredFile, "C:", "/C", 1)
		} else if !strings.Contains(err.Error(), "executable file not found") {
			return "", errors.Annotate(err, string(b))
		}
	}
	out, err := gitCMD("config",
		"--global", fmt.Sprintf("credential.%s.helper", gitaddr),
		fmt.Sprintf("store --file %s", gitCfgFile),
	)
	if err != nil {
		return "", errors.Annotate(err, out)
	}

	// Get git server details
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

	data := []byte(fmt.Sprintf("%s://%s:%s@%s%%3a%s", gitprotocol, username, password, githost, gitport))
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

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
