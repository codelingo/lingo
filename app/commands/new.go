package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
)

// TODO(waigani) add a quick-start cmd with --dry-run flag that: inits git,
// inits lingo, sets up auth and register's repo with CodeLingo platform.

func init() {
	register(&cli.Command{
		Name:   "init",
		Usage:  "Create a .lingo file in the current directory.",
		Action: newLingoAction,
	}, false, vcsRq, versionRq)
}

var intro = `
# Add and configure tenets following these guidelines: http://codelingo/lingo/getting-started#the-lingo-file.

`[1:]

// TODO(waigani) set lingo-home flag and test init creates correct home dir.

func newLingoAction(ctx *cli.Context) {
	if err := newLingo(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Println("Success! A .lingo file has been written in the current directory. Edit it with your editor of choice to get started writing Tenets.")

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
		return "", errors.Errorf(".lingo file already exists: %q", cfgPath)
	}

	return cfgPath, writeDotLingo(cfgPath)
}

func writeDotLingo(cfgPath string) error {
	// TODO: Language argument to specify query.
	defaultDotLingo := util.DotLingo{
		Tenets: []util.Tenet{
			{
				Bots: map[string]util.Bot{
					"codelingo/clair": {

						Name:    "find-funcs",
						Doc:     "Example tenet that finds all functions.",
						Comment: "This is a function, but you probably already knew that.",
					},
				},
				Query: `
import codelingo/ast/common/0.0.0

@ clair.comment
common.func({depth: any})
`[1:],
			},
		},
	}
	byt, err := yaml.Marshal(defaultDotLingo)
	if err != nil {
		return errors.Trace(err)
	}

	// Our syntax requires unorthodox extra double space. TODO: fix.
	byt = bytes.Replace(byt, []byte("\n"), []byte("\n  "), -1)

	// add comment to file
	// TODO(waigani) comments seem to cause a "corrupt patch" error, removing this for now.
	// byt = append(byt, []byte(intro)...)

	if err = ioutil.WriteFile(cfgPath, byt, 0644); err != nil {
		return errors.Trace(err)
	}
	return nil
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
