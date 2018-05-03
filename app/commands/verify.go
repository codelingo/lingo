package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
	utilConfig "github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
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
		outdated, err := VersionIsOutdated()
		if err != nil {
			if outdated {
				// Don't error, just warn
				// TODO(waigani) use below and https://godoc.org/go.uber.org/zap/zapcore#NewConsoleEncoder
				// util.Logger.Warn(err.Error())
				fmt.Fprint(os.Stderr, err.Error(), "\n")

				return nil
			} else {
				return errors.Trace(err)
			}
		}
		return nil
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
		return "run `$ lingo config setup`"
	case homeRq:
		return "run `$ lingo config setup`"
	case configRq:
		return "run `$ lingo config setup`"
	}

	return fmt.Sprintf("unknown require type %d", r)
}

// TODO(waigani) allow outside repo if we're making a remote pr review. Get
// the repo owner/name from the PR URL.
func verifyVCS() error {
	cmd := exec.Command("git", "status")

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	var gitErr error
	if err := cmd.Run(); err != nil {
		gitErr = errors.New(errBuf.String())
	} else {
		return nil
	}
	cmd = exec.Command("p4", "status")
	if err := cmd.Run(); err != nil {
		return errors.Annotate(errors.Annotate(err, gitErr.Error()), "lingo cannot be used outside of a git or perforce repository")
	}
	return nil
}

func verifyAuth() error {

	errMsg := "Authentication failed. Please run `lingo config setup`."
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
			// TODO(waigani) use below and https://godoc.org/go.uber.org/zap/zapcore#NewConsoleEncoder
			// util.Logger.Warn("could not create scripts directory:", err.Error())
			fmt.Fprint(os.Stderr, "could not create scripts directory:", err.Error())
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

	configDefaults, err := util.ConfigDefaults()
	configDefaults = filepath.Join(configDefaults, common.ClientVersion)
	if err != nil {
		return errors.Trace(err)
	}
	if _, err := os.Stat(configDefaults); os.IsNotExist(err) {
		err := os.MkdirAll(configDefaults, 0775)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create default configs directory")
		}
	}

	envCfg := filepath.Join(configsHome, utilConfig.EnvCfgFile)
	if _, err := os.Stat(envCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(envCfg, []byte("paas"), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create env config")
		}
	}

	err = createConfigDefaultFiles(configDefaults)
	if err != nil {
		return errors.Trace(err)
	}

	err = utilConfig.CreateAuthFile()
	if err != nil {
		return errors.Trace(err)
	}

	err = utilConfig.CreatePlatformFile()
	if err != nil {
		return errors.Trace(err)
	}

	err = utilConfig.CreateVersionFile()
	if err != nil {
		return errors.Trace(err)
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

const missingConfigError string = "Could not get %s config. Please run `lingo config setup`."

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

	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	lastChecked, err := time.Parse(layout, lastCheckedString)
	if err != nil {
		return errors.Trace(err)
	}

	duration := time.Since(lastChecked)
	if duration.Hours() >= 24 {
		latestVersion, err := latestVersion()
		if err != nil {
			return errors.Trace(err)
		}

		err = vCfg.SetClientLatestVersion(latestVersion.String())
		if err != nil {
			return errors.Trace(err)
		}
		err = vCfg.SetClientVersionLastChecked(time.Now().UTC().String())
		if err != nil {
			return errors.Trace(err)
		}
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
		return errors.New("Your client is out of date. This may result in unexpected behaviour.")
	} else if compare > 0 {
		return errors.New("Your client is newer than the platform. This may result in unexpected behaviour.")
	} else {
		// Versions are equal: OK.
		return nil
	}
}

func latestVersion() (*semver.Version, error) {
	svc, err := service.New()
	if err != nil {
		return nil, errors.Trace(err)
	}

	versionString, err := svc.LatestClientVersion()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return semver.New(versionString)
}

func compareVersions(current, latest string) (int, error) {
	cv, err := semver.Make(current)
	if err != nil {
		return -1, errors.Trace(err)
	}

	lv, err := semver.Make(latest)
	if err != nil {
		return 1, errors.Trace(err)
	}

	return cv.Compare(lv), nil
}

func VersionIsOutdated() (bool, error) {
	err := verifyClientVersion()
	if err != nil {
		if strings.HasSuffix(err.Error(), "This may result in unexpected behaviour.") {
			return true, err
		} else {
			return false, errors.Trace(err)
		}
	}
	return false, err
}
