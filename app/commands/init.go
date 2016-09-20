package commands

import (
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
		Usage:  "create a .lingo config file in the current directory",
		Action: initLingo,
	}, false, vcsRq)
}

var intro = `
# Add and configure tenets following these guidelines: http://codelingo/lingo/getting-started#the-lingo-file.

`[1:]

// TODO(waigani) set lingo-home flag and test init creates correct home dir.

func initLingo(c *cli.Context) {
	if err := util.MaxArgs(c, 1); err != nil {
		util.OSErrf(err.Error())
		return
	}

	cfgPath, err := getCfgPath(c)
	if err != nil {
		util.OSErrf(err.Error())
		return
	}
	if _, err := os.Stat(cfgPath); err == nil {
		util.OSErrf("Already initialised. Using %q", cfgPath)
		return
	}

	if err := writeDotLingo(cfgPath); err != nil {
		util.OSErrf(err.Error())
		return
	}
	fmt.Printf("Successfully initialised. Lingo config file written to %q\n", cfgPath)
}

func writeDotLingo(cfgPath string) error {
	defaultDotLingo := util.DotLingo{
		Lexicons: []string{
			"codelingo/common as _",
		},
		Tenets: []util.Tenet{
			{
				Name:    "find-funcs",
				Doc:     "Example tenet that finds all functions.",
				Comment: "This is a function, but you probably already knew that.",
				Match: `
<func
`[1:],
			},
		},
	}
	byt, err := yaml.Marshal(defaultDotLingo)
	if err != nil {
		return errors.Trace(err)
	}

	// add comment to file
	byt = append(byt, []byte(intro)...)

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
