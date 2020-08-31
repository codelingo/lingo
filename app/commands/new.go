package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"
	"github.com/urfave/cli"

	"github.com/juju/errors"
)

// TODO(waigani) add a quick-start cmd with --dry-run flag that: inits git,
// inits lingo, sets up auth and register's repo with CodeLingo platform.

func init() {
	register(&cli.Command{
		Name:   "init",
		Usage:  "Create a codelingo.yaml file in the current directory.",
		Action: newLingoAction,
	}, false, false, verify.VCSRq, verify.VersionRq)
}

// TODO(waigani) set lingo-home flag and test init creates correct home dir.

func newLingoAction(ctx *cli.Context) {
	if err := newLingo(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Println("Success! A codelingo.yaml file has been written in the current directory. Edit it with your editor of choice to get started writing Specs.")

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
specs:
  - name: template-spec   # Template for a Spec using the Codelingo Review Action
    actions:
      codelingo/review:
        comment: This will be commented on any code which matches any fact tagged with the '@review comment' decorator.
    query: |
      import codelingo/ast/<language>   # Replace <language> with the relevent language for your Spec eg. codelingo/ast/go

      # Begin Query here, at-least one fact must be decorated with '@review comment' for an automated code-review
      # See https://www.codelingo.io/specs for examples
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
