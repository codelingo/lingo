package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/vcs/git"
	"github.com/codelingo/lingo/vcs/p4"
	"github.com/juju/errors"
)

func New() (Type, Repo, error) {
	b, err := DetectVCSType()
	if err != nil {
		return 999, nil, errors.Trace(err)
	}
	switch b {
	case Git:
		return Git, git.New(), nil
	case P4:
		return P4, p4.New(), nil
	}
	return 999, nil, errors.New("cannot find a matched VCS type")
}

func TypeToString(b Type) (string, error) {
	switch b {
	case Git:
		return "Git", nil
	case P4:
		return "P4", nil
	}
	return "", errors.New("unknow VCS type")
}

func DetectVCSType() (Type, error) {
	cmd := exec.Command("git", "status")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return Git, nil
	}
	gitErr := errors.Annotate(err, string(out))

	cmd = exec.Command("p4", "status")
	out, err = cmd.CombinedOutput()
	if err != nil {
		errors.Annotate(err, gitErr.Error())
		return 999, errors.Annotate(err, "cannot find a known VCS in current directory")
	}
	return P4, nil
}

// sync the local repository with the remote, creating the remote if it does
// not exist.
func SyncRepo(vcsType Type, repo Repo) error {
	repoOwner, err := getRepoOwner(vcsType)
	if err != nil {
		return errors.Trace(err)
	}
	// get the repo name, default to working directory name
	dir, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}
	if syncErr := repo.Sync(repoOwner, dir); syncErr != nil {
		if vcsType == Git {
			// This case is triggered when a local remote has been added but
			// the remote repo does not exist. In this case, we create the
			// remote and try to sync again.
			missingRemote, err := regexp.MatchString("fatal: repository '.*' not found.*", syncErr.Error())
			if err != nil {
				return errors.Trace(err)
			}
			if missingRemote {
				_, repoName, err := repo.OwnerAndNameFromRemote()
				if err != nil {
					return errors.Trace(err)
				}

				// TODO(waigani) use typed errors
				if err := repo.CreateRemote(repoName); err != nil && !strings.HasPrefix(err.Error(), "repository already exists") {
					return errors.Trace(err)
				}
				if err := repo.Sync(repoOwner, dir); err != nil {
					return errors.Trace(err)
				}
			}
		}
		return errors.Trace(syncErr)
	}
	return nil
}

func InitRepo(vcsType Type, repo Repo) error {
	repoOwner, err := getRepoOwner(vcsType)
	if err != nil {
		return errors.Trace(err)
	}
	// get the repo name, default to working directory name
	dir, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}
	repoName := filepath.Base(dir)

	// TODO(benjamin-rood) Check if repo name and contents are identical.
	// If, so this should be a no-op and only the remote needs to be set.
	// ensure creation of distinct remote.
	repoName, err = CreateRepo(repo, repoName)
	if err != nil {
		return errors.Trace(err)
	}
	if err = repo.AssertNotTracked(); err != nil {
		// TODO (benjamin-rood): Check the error type
		return errors.Trace(err)
	}
	_, _, err = repo.SetRemote(repoOwner, repoName)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func CreateRepo(repo Repo, name string) (string, error) {
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
			return CreateRepo(repo, fmt.Sprintf("%s-%d", name, n+1))
		}
		return "", err
	}
	return name, nil
}

func getRepoOwner(vcsType Type) (string, error) {
	authCfg, err := config.Auth()
	if err != nil {
		return "", errors.Trace(err)
	}
	repoOwner := ""
	switch vcsType {
	case Git:
		// TODO(waigani) Try to get owner and name from origin remote first.
		// get the repo owner name
		repoOwner, err = authCfg.GetGitUserName()
		if err != nil {
			return "", errors.Trace(err)
		}
	case P4:
		repoOwner, err = authCfg.GetP4UserName()
		if err != nil {
			return "", errors.Trace(err)
		}
	}
	if repoOwner == "" {
		return "", errors.New("Please run lingo setup")
	}
	return repoOwner, nil
}
