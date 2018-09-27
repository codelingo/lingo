package commands

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	serviceConfig "github.com/codelingo/lingo/service/config"
	"github.com/howeyc/gopass"
	"github.com/juju/errors"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	websiteAddr      = "website.addr"
	platformAddr     = "platform.addr"
	flowAddr         = "flow.address"
	gitServerAddr    = "gitserver.remote.host"
	p4ServerAddr     = "p4server.remote.host"
	messagequeueAddr = "messagequeue.address.host"
	baseDiscoveryURL = "https://raw.githubusercontent.com/codelingo/codelingo/master/"
)

func init() {
	register(&cli.Command{
		Name:   "config",
		Usage:  "Show summary of config",
		Action: configAction,
		Subcommands: []cli.Command{
			{
				Name:   "env",
				Usage:  "Show the current environment.",
				Action: configEnvAction,
				Subcommands: []cli.Command{
					{
						Name:   "use",
						Usage:  "Use the given environment.",
						Action: useEnvAction,
					},
				},
			},
			{
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
			},
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "ip",
				Usage: "Set the IP of the OnPrem Platform.",
			},
		},
	}, false, false, homeRq, configRq)
}

func configAction(ctx *cli.Context) {
	err := showConfig(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func showConfig(ctx *cli.Context) error {
	username, err := getUsername()
	if err != nil {
		return errors.Trace(err)
	}
	env, err := getEnv()
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Printf(`Username: %s
Environment: %s
`, username, env)

	return nil
}

func useEnvAction(ctx *cli.Context) {
	err := useEnv(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func useEnv(ctx *cli.Context) error {
	var err error
	switch len(ctx.Args()) {
	case 0:
		err = errors.New("Error: An environment value must be specified: `lingo use-env <env>`")
		return err
	case 1:
		// Success case
		break
	default:
		err = errors.New("Error: Only 1 environment value can be specified: `lingo use-env <env>`")
		return err
	}

	configsHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	newEnv := ctx.Args()[0]
	envFilepath := filepath.Join(configsHome, commonConfig.EnvCfgFile)

	cfg := serviceConfig.New(envFilepath)
	err = cfg.SetEnv(newEnv)
	if err != nil {
		return errors.Trace(err)
	}
	if newEnv == "onprem" {
		ip := ctx.String("ip")
		if ip == "" {
			fmt.Print("Enter the Platform IP: ")
			fmt.Scanln(&ip)
		}
		cfg, err := commonConfig.Platform()
		if err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(websiteAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(platformAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(flowAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(gitServerAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(p4ServerAddr, ip); err != nil {
			return errors.Trace(err)
		}
		if err := cfg.Set(messagequeueAddr, ip); err != nil {
			return errors.Trace(err)
		}
	}
	fmt.Printf("Success! Environment set to '%v'.\n", newEnv)

	return nil
}

func configEnvAction(ctx *cli.Context) {
	err := configEnv(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func configEnv(ctx *cli.Context) error {
	env, err := getEnv()
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Println(env)

	return nil
}

func getEnv() (string, error) {
	configsHome, err := util.ConfigHome()
	if err != nil {
		return "", errors.Trace(err)
	}
	envFilepath := filepath.Join(configsHome, commonConfig.EnvCfgFile)
	cfg := serviceConfig.New(envFilepath)
	env, err := cfg.GetEnv()
	if err != nil {
		return "", errors.Trace(err)
	}

	return env, nil
}

func getUsername() (string, error) {
	authCfg, err := commonConfig.Auth()
	if err != nil {
		return "", errors.Trace(err)
	}

	// TODO: have a single CodeLingo username instead of using repo usernames
	for i := 0; i < 3; i++ {
		var username string
		var err error

		switch i {
		case 0:
			username, err = authCfg.GetGitUserName()
		case 1:
			username, err = authCfg.GetP4UserName()
		default:
			util.UserFacingWarning("Warning: username not set yet. Run `lingo config setup` to set your username.")
			username, err = "", nil
		}

		if err != nil {
			if strings.Contains(err.Error(), "Could not find value") {
				continue
			} else {
				return "", errors.Trace(err)
			}
		}

		return username, nil
	}

	return "", nil
}

func setupLingoAction(c *cli.Context) {
	username, err := setupLingo(c)
	if err != nil {
		util.FatalOSErr(err)
	}

	fmt.Println("Success! CodeLingo user set to", username)
}

// TODO(waigani) dry run this
func setupLingo(c *cli.Context) (string, error) {

	if err := util.MaxArgs(c, 0); err != nil {
		return "", errors.Trace(err)
	}

	platConfig, err := commonConfig.Platform()
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
		lingoTokenAddr := webAddr + "/settings/profile"
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
		if err := cmd.Run(); err != nil {
			byt, err = terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "", errors.Trace(err)
			}
		} else {
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

	gitAddr, err := platConfig.GitServerAddr()
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
		"--global", fmt.Sprintf("credential.%s.helper", gitAddr),
		fmt.Sprintf("store --file %s", gitCfgFile),
	)
	if err != nil {
		return "", errors.Annotate(err, out)
	}

	// Set git server details
	gitURL, err := url.Parse(gitAddr)
	if err != nil {
		return "", errors.Trace(err)
	}
	gitURL.User = url.UserPassword(username, password)
	data := []byte(gitURL.String())

	// Write creds to config file
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
