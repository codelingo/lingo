package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codelingo/lingo/app/util"
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
	case vcsRq:
		verifyVCS()
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

	authCfg, err := util.ReadAuthConfig()
	if err != nil {
		return errors.Trace(err)
	}

	if authCfg.CurrentUserToken == "" {
		return errors.New("no current user token found, please log in: `$ lingo login`")
	}

	return nil
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

	return errors.New("no configs set up yet")

	// lHome, err := util.LingoHome()
	// if err != nil {
	// 	return errors.Trace(err)
	// }

	// configsHome := filepath.Join(lHome, "configs")
	// if _, err := os.Stat(configsHome); os.IsNotExist(err) {
	// 	err := os.MkdirAll(configsHome, 0775)
	// 	if err != nil {
	// 		errors.Annotatef(err, "WARNING: could not create configs directory")
	// 	}
	// }

	// servicesCfg := filepath.Join(configsHome, config.ServicesCfgFile)
	// if _, err := os.Stat(servicesCfg); os.IsNotExist(err) {
	// 	err := ioutil.WriteFile(servicesCfg, []byte(config.ServicesTmpl), 0644)
	// 	if err != nil {
	// 		fmt.Printf("WARNING: could not create services config: %v \n", err)
	// 	}
	// }

	// defaultsCfg := filepath.Join(configsHome, config.DefaultsCfgFile)
	// if _, err := os.Stat(defaultsCfg); os.IsNotExist(err) {
	// 	err := ioutil.WriteFile(defaultsCfg, []byte(config.DefaultsTmpl), 0644)
	// 	if err != nil {
	// 		fmt.Printf("WARNING: could not create services config: %v \n", err)
	// 	}
	// }
}
