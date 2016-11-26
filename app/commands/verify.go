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
	cfg, err := util.AuthConfig()
	if err != nil {
		return errors.Annotate(err, errMsg)
	}

	authFile, err := cfg.Get("gitserver.credentials_filename")
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
			errors.Annotatef(err, "WARNING: could not create configs directory")
		}
	}

	platformCfg := filepath.Join(configsHome, utilConfig.PlatformCfgFile)
	if _, err := os.Stat(platformCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(platformCfg, []byte(utilConfig.PlatformTmpl), 0644)
		if err != nil {
			fmt.Printf("WARNING: could not create platform config: %v \n", err)
		}
	}

	versionCfg := filepath.Join(configsHome, utilConfig.VersionCfgFile)
	if _, err := os.Stat(versionCfg); os.IsNotExist(err) {
		err := ioutil.WriteFile(versionCfg, []byte(utilConfig.VersionTmpl), 0644)
		if err != nil {
			fmt.Printf("WARNING: could not create version config: %v \n", err)
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

func verifyClientVersion() error {
	cfg, err := utilConfig.Version()
	if err != nil {
		return err
	}

	version, err := cfg.ClientVersion()
	if err != nil {
		return err
	}
	// TODO: Use `hashicorp/go-version` package for comparing and setting semvers
	// https://github.com/hashicorp/go-version
	if version != common.ClientVersion {
		return errors.New("Update required. Please run $ lingo setup --keep-creds.")
	}
	return nil
}
