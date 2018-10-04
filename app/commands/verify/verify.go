package verify

import (
	"bytes"
	"context"
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
	"github.com/codelingo/lingo/app/util/common/config"
	utilConfig "github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

type Require int

const (
	// all cmds have the base Requirement
	BaseRq Require = iota
	DotLingoRq
	AuthRq
	HomeRq
	ConfigRq
	VCSRq
	VersionRq
)

func (r Require) String() string {
	switch r {
	case DotLingoRq:
		return "codelingo.yaml"
	case AuthRq:
		return "authentication"
	case HomeRq:
		return "lingohome"
	case ConfigRq:
		return "config"
	case VCSRq:
		return "git"
	}
	return "unknown"
}

// Verifies that the Requirement is met
func (r Require) Verify() error {
	switch r {
	case BaseRq:
		return nil
	case VersionRq:
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
	case VCSRq:
		return verifyVCS()
	case DotLingoRq:
		return verifyDotLingo()
	case AuthRq:
		return verifyAuth()
	case HomeRq:
		return verifyLingoHome()
	case ConfigRq:
		return verifyConfig()
	}

	return errors.Errorf("unknown require type %d", r)
}

func (r Require) HelpMsg() string {
	switch r {
	case DotLingoRq:
		return "run `$ lingo init`"
	case AuthRq:
		return "run `$ lingo config setup`"
	case HomeRq:
		return "run `$ lingo config setup`"
	case ConfigRq:
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
	for filename := range common.LingoFilenames {
		if _, err := os.Stat(filename); err == nil {
			return nil
		}
	}
	return errors.New("no lingo file found, please run `$ lingo init`")
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

	err = CreateConfigDefaultFiles(configDefaults)
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

const MissingConfigError string = "Could not get %s config. Please run `lingo config setup`."

func verifyClientVersion() error {
	// TODO: revisit platform version checking.
	vCfg, err := utilConfig.Version()
	if err != nil {
		// TODO(waigani) don't throw error away before checking type.
		return errors.New(fmt.Sprintf(MissingConfigError, "version"))
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
	if duration.Hours() >= 4 {
		fmt.Println("Checking for updates...")

		latest, found, err := selfupdate.DetectLatest("codelingo/lingo")
		if err != nil {
			return errors.Trace(err)
		}

		err = vCfg.SetClientVersionLastChecked(time.Now().UTC().String())
		if err != nil {
			return errors.Trace(err)
		}

		if !found {
			return nil
		}

		err = vCfg.SetClientLatestVersion(latest.Version.String())
		if err != nil {
			return errors.Trace(err)
		}

		v := semver.MustParse(common.ClientVersion)
		if latest.Version.Equals(v) {
			return nil
		}

		if v.LT(latest.Version) {
			fmt.Printf("A new client version (%s) is available. Run 'lingo update' to update.\n", latest.Version.String())
		}
	}

	return nil
}

func latestVersion() (*semver.Version, error) {
	ctx, _ := util.UserCancelContext(context.Background())
	versionString, err := service.LatestClientVersion(ctx)
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

func CreateConfigDefaultFiles(dir string) error {
	err := config.CreateAuthFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreatePlatformFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	err = config.CreateVersionFileInDir(dir, true)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
