package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

// TODO(waigani) add a quick-start cmd with --dry-run flag that: inits git,
// inits lingo, sets up auth and register's repo with CodeLingo platform.

func init() {
	register(&cli.Command{
		Name:   "init",
		Usage:  "Create a codelingo.yaml file in the current directory.",
		Action: newLingoAction,
	}, false, false, vcsRq, versionRq)
}

// TODO(waigani) set lingo-home flag and test init creates correct home dir.

func newLingoAction(ctx *cli.Context) {
	if err := newLingo(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Println("Success! A codelingo.yaml file has been written in the current directory. Edit it with your editor of choice to get started writing Tenets.")

}

func newLingo(c *cli.Context) error {
	if err := util.MaxArgs(c, 1); err != nil {
		return errors.Trace(err)
	}

	_, err := writeDotLingoToCurrentDir(c)
	return errors.Trace(err)
}

func writeDotLingoToCurrentDir(c *cli.Context) (string, error) {
	cfgPath, err := getCfgPath(c)
	if err != nil {
		return "", errors.Trace(err)
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return "", errors.Errorf("codelingo.yaml file already exists: %q", cfgPath)
	}

	return cfgPath, writeDotLingo(cfgPath)
}

func writeDotLingo(cfgPath string) error {
	lingoSrc := `
tenets:
  - name: find-funcs
    doc: Example tenet that finds all functions.
    flows:
      codelingo/review:
        comment: This is a function, but you probably already knew that.
    query: |
      import codelingo/ast/common

      @ review.comment
      common.func({depth: any})
`[1:]

	return errors.Trace(ioutil.WriteFile(cfgPath, []byte(lingoSrc), 0644))
}

func getCfgPath(c *cli.Context) (string, error) {
	// Create a new tenet config file at either the provided directory or
	// a location from flags or environment, or the current directory
	cfgPath, _ := filepath.Abs(util.DesiredTenetCfgPath(c))
	if len(c.Args()) > 0 {
		cfgPath, _ = filepath.Abs(c.Args()[0])

		// Check that it exists and is a directory
		if pathInfo, err := os.Stat(cfgPath); os.IsNotExist(err) {
			return "", errors.Annotatef(err, "directory %q not found", cfgPath)
		} else if !pathInfo.IsDir() {
			return "", errors.Errorf("%q is not a directory", cfgPath)
		}

		// Use default config filename
		cfgPath = filepath.Join(cfgPath, util.DefaultTenetCfgPath)
	}
	return cfgPath, nil
}
