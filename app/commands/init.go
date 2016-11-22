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

	if err = repo.AssertNotTracked(); err != nil {
		// TODO (benjamin-rood): Check the error type
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

	// TODO(benjamin-rood) Check if repo name and contents are identical.
	// If, so this should be a no-op and only the remote needs to be set.
	// ensure creation of distinct remote.
	repoName, err = createRepo(repo, repoName)
	if err != nil {
		return "", "", errors.Trace(err)
	}
	return repo.SetRemote(repoOwner, repoName)
}

func createRepo(repo backing.Repo, name string) (string, error) {
	if err := repo.CreateRemote(name); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			parts := strings.Split(name, "-")
			num := parts[len(parts)-1]

			// We ignore the error here because the only case in which Atoi
			// would error is if the name had not yet been appended with -n.
			// In this case, n will be set to zero which is what we require.
			n, _ := strconv.Atoi(num)
			if n != 0 {
				// Need to remove existing trailing number where present,
				// otherwise the repoName only appends rather than replaces
				// and will produce names of the pattern "myPkg-1-2-...-n-n+1".
				name = strings.TrimSuffix(name, "-"+num)
			}
			return createRepo(repo, fmt.Sprintf("%s-%d", name, n+1))
		}
		return "", err
	}
	return name, nil
}
