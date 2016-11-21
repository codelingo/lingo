package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// TODO(waigani) start ingesting repo as soon as it's inited

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
	fmt.Printf("Successfully initialised. \n Added remote %q %s\n", remoteName, remoteAddr)
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

	// get the repo name, default to working directory name
	dir, err := os.Getwd()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	repoName := filepath.Base(dir)

	// TODO(benjamin-rood) check if repo name and contents are identical
	// If, so this should be a no-op and only the remote needs to be set.
	// ensure creation of distinct remote
	repoName = createDistinctRepoName(repoName, repo)

	if err := repo.CreateRemote(repoName); err != nil {
		return "", "", errors.Trace(err)
	}
	return repo.SetRemote(repoOwner, repoName)
}

func createDistinctRepoName(name string, repo backing.Repo) string {
	if exists, _ := repo.Exists(name); !exists {
		return name
	}
	// If repoName is present in GOGS, append the string with an incremented int
	return createDistinctRepoName(incrementRepoVersion(name), repo)
}

func incrementRepoVersion(r string) string {
	if ok, sep := repoNameIsVersion(r); ok {
		newRepoName := ""
		version, _ := strconv.Atoi(sep[len(sep)-1])
		version++
		suffix := strconv.Itoa(version)
		for i := 0; i < len(sep)-1; i++ {
			newRepoName += (sep[i] + "-")
		}
		newRepoName += suffix
		return newRepoName
	}
	return (r + "-1") // begin suffix at 1
}

func repoNameIsVersion(r string) (bool, []string) {
	sep := strings.Split(r, "-")
	if len(sep) > 1 {
		suffix := sep[len(sep)-1]
		if _, err := strconv.Atoi(suffix); err == nil {
			return true, sep
		}
	}
	return false, sep
}
