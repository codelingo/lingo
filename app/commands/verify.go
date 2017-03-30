package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
	utilConfig "github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"
	"time"
	"strings"
	"strconv"
)

type require int

const (
	// all cmds have the base requirement
	baseRq require = iota
	dotLingoRq
	authRq
	homeRq
	configRq
	vcsRq
	versionRq
)

func (r require) String() string {
	switch r {
	case dotLingoRq:
		return ".lingo"
	case authRq:
		return "authentication"
	case homeRq:
		return "lingohome"
	case configRq:
		return "config"
	case vcsRq:
		return "git"
	}
	return "unknown"
}

// Verifies that the requirement is met
func (r require) Verify() error {
	switch r {
	case baseRq:
		return nil
	case versionRq:
		return verifyClientVersion()
	case vcsRq:
		return verifyVCS()
	case dotLingoRq:
		return verifyDotLingo()
	case authRq:
		return verifyAuth()
	case homeRq:
		return verifyLingoHome()
	case configRq:
		return verifyConfig()
	}

	return errors.Errorf("unknown require type %d", r)
}

func (r require) HelpMsg() string {
	switch r {
	case dotLingoRq:
		return "run `$ lingo init`"
	case authRq:
		return "run `$ lingo auth`"
	case homeRq:
		return "run `$ lingo init`"
	case configRq:
		return "run `$ lingo init`"
	}

	return fmt.Sprintf("unknown require type %d", r)
}

// TODO(waigani) allow outside repo if we're making a remote pr review. Get
// the repo owner/name from the PR URL.
func verifyVCS() error {
	cmd := exec.Command("git", "status")

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "lingo cannot be used outside of a git repository")
	}

	if errBuf.Len() > 0 {
		return errors.New(errBuf.String())
	}
	return nil
}

func verifyAuth() error {

	errMsg := "Authentication failed. Please run `lingo setup`."
	authCfg, err := utilConfig.Auth()
	if err != nil {
		return errors.Annotate(err, errMsg)
	}

	authFile, err := authCfg.GetGitCredentialsFilename()
	if err != nil {
		return errors.Annotate(err, errMsg)
	} else if authFile == "" {
		return errors.New(errMsg)
	}

	cfgDir, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}

	_, err = os.Stat(filepath.Join(cfgDir, authFile))
	if os.IsNotExist(err) {
		return errors.New(errMsg)
	}
	return errors.Trace(err)
}

func verifyDotLingo() error {
	_, err := os.Stat(".lingo")
	return err

	// TODO(waigani) do this properly, return

	// if cfgPath, _ := common.TenetCfgPath(c); cfgPath == "" {
	// 	return errors.Wrap(common.ErrMissingDotLingo, errors.New("ui"))
	// }
	// return nil
}

func verifyLingoHome() error {
	// create lingo home if it doesn't exist
	lHome := util.MustLingoHome()
	if _, err := os.Stat(lHome); os.IsNotExist(err) {
		err := os.MkdirAll(lHome, 0775)
		if err != nil {
			return errors.Trace(err)
		}
	}

	tenetsHome := filepath.Join(lHome, "tenets")
	if _, err := os.Stat(tenetsHome); os.IsNotExist(err) {
		err := os.MkdirAll(tenetsHome, 0775)
		if err != nil {
			return errors.Trace(err)
		}
	}

	logsHome := filepath.Join(lHome, "logs")
	if _, err := os.Stat(logsHome); os.IsNotExist(err) {
		err := os.MkdirAll(logsHome, 0775)
		if err != nil {
			return errors.Trace(err)
		}
	}

	scriptsHome := filepath.Join(lHome, "scripts")
	if _, err := os.Stat(scriptsHome); os.IsNotExist(err) {
		err := os.MkdirAll(scriptsHome, 0775)
		if err != nil {
			fmt.Printf("WARNING: could not create scripts directory: %v \n", err)
		}
	}

	return nil
}

func verifyConfig() error {
	configsHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	if _, err := os.Stat(configsHome); os.IsNotExist(err) {
		err := os.MkdirAll(configsHome, 0775)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create configs directory")
		}
	}

	envCfg := filepath.Join(configsHome, utilConfig.EnvCfgFile)
	if _, err := os.Stat(envCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(envCfg, []byte("all"), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create env config")
		}
	}

	authCfg := filepath.Join(configsHome, utilConfig.AuthCfgFile)
	if _, err := os.Stat(authCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(authCfg, []byte(utilConfig.AuthTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create auth config")
		}
	}

	platformCfg := filepath.Join(configsHome, utilConfig.PlatformCfgFile)
	if _, err := os.Stat(platformCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(platformCfg, []byte(utilConfig.PlatformTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create platform config")
		}
	}

	versionCfg := filepath.Join(configsHome, utilConfig.VersionCfgFile)
	if _, err := os.Stat(versionCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(versionCfg, []byte(utilConfig.VersionTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create version config")
		}
	}

	// servicesCfg := filepath.Join(configsHome, utilConfig.ServicesCfgFile)
	// if _, err := os.Stat(servicesCfg); os.IsNotExist(err) {
	// 	err := ioutil.WriteFile(servicesCfg, []byte(utilConfig.ServicesTmpl), 0644)
	// 	if err != nil {
	// 		fmt.Printf("WARNING: could not create services config: %v \n", err)
	// 	}
	// }

	// defaultsCfg := filepath.Join(configsHome, utilConfig.DefaultsCfgFile)
	// if _, err := os.Stat(defaultsCfg); os.IsNotExist(err) {
	// 	err := ioutil.WriteFile(defaultsCfg, []byte(utilConfig.DefaultsTmpl), 0644)
	// 	if err != nil {
	// 		fmt.Printf("WARNING: could not create services config: %v \n", err)
	// 	}
	// }
	return nil
}

const missingConfigError string = "Could not get %s config. Please run `lingo setup`."

func verifyClientVersion() error {
	vCfg, err := utilConfig.Version()
	if err != nil {
		// TODO(waigani) don't throw error away before checking type.
		return errors.New(fmt.Sprintf(missingConfigError, "version"))
	}

	lastCheckedString, err := vCfg.ClientVersionLastChecked()
	if err != nil {
		return errors.Trace(err)
	}

	layout := "2006-01-02 15:04:05.000000000 -0700 MST"
	lastChecked, err := time.Parse(layout, lastCheckedString)
	if err != nil {
		return errors.Trace(err)
	}

	duration := time.Since(lastChecked)
	if duration.Hours() >= 24 {
		fmt.Println("TODO: Call platform version endpoint.")
		// Call on the platform endpoint if need be
		// Update version in config
		// Print needs update
	}

	latestVersion, err := vCfg.ClientLatestVersion()
	if err != nil {
		return errors.Trace(err)
	}

	compare, err := compareVersions(common.ClientVersion, latestVersion)
	if err != nil {
		return errors.Trace(err)
	}

	if compare < 0 {
		fmt.Println("Warning: Your client is out of date. This may result in unexpected behaviour.")
	} else if compare > 0 {
		fmt.Println("Warning: Your client is newer than the platform. This may result in unexpected behaviour.")
	} else {
		// Versions are equal: OK.
	}

	return nil
}

func compareVersions(v1, v2 string) (int, error) {
	v1Tokens := strings.Split(v1, ".")
	v2Tokens := strings.Split(v2, ".")

	var result int = 0

	for index, v1Token := range v1Tokens {
		if len(v2Tokens) > index {
			v2Token := v2Tokens[index]

			v1Int, err := strconv.Atoi(v1Token)
			if err != nil {
				return -1, errors.New("Could not convert all parts of the current version string to integers.")
			}
			v2Int, err := strconv.Atoi(v2Token)
			if err != nil {
				return 1, errors.New("Could not convert all parts of the latest version string to integers.")
			}

			if v1Int > v2Int {
				return 1, nil
			} else if v1Int < v2Int {
				return -1, nil
			} else {
				continue
			}
		} else {
			return 1, nil
		}
	}

	if result == 0 && len(v1Tokens) < len(v2Tokens) {
		result = -1
	}

	return result, nil
}
