package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"
	"github.com/juju/errors"

	"github.com/codegangsta/cli"
)

func init() {
	register(&cli.Command{
		Name:   "init",
		Usage:  "Setup lingo for the current repository",
		Action: initRepoAction,
		// TODO(waigani) add --dry-run flag
	}, false, vcsRq)
}

// TODO(waigani) start injesting repo as soon as it's inited

func initRepoAction(ctx *cli.Context) {
	remoteName, remoteAddr, err := initRepo(ctx)
	if err != nil {

		// TODO(waigani) use error types
		if strings.Contains(err.Error(), "already exists") {
			util.OSErrf("lingo already initialized for this repository.")
		} else {
			util.OSErrf(err.Error())
		}
		return
	}
	fmt.Printf("Successfully initialised. \n Added remote %q %s\a Starting injest...\n", remoteName, remoteAddr)
}

func initRepo(ctx *cli.Context) (string, string, error) {

	repo := vcs.New(backing.Git)
	authCfg, err := util.AuthConfig()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	// TODO(waigani) Try to get owner and name from origin remote first.

	// get the repo owner name
	repoOwner, err := authCfg.Get("gitserver.user.username")
	if err != nil {
		return "", "", errors.Trace(err)
	}
	var repoOwnerCustom string
	fmt.Printf("repo owner(%s): ", repoOwner)
	fmt.Scanln(&repoOwnerCustom)
	if repoOwnerCustom != "" {
		repoOwner = repoOwnerCustom
	}

	// get the repo name, default to working directory name
	dir, err := os.Getwd()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	repoName := filepath.Base(dir)
	var repoNameCustom string
	fmt.Printf("repo name(%s): ", repoName)
	fmt.Scanln(&repoNameCustom)
	if repoNameCustom != "" {
		repoName = repoNameCustom
	}

	return repo.SetRemote(repoOwner, repoName)
}
