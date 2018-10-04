package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/vcs/git"
	"github.com/codelingo/lingo/vcs/p4"
	"github.com/juju/errors"
)

const (
	vcsGit string = "git"
	vcsP4  string = "perforce"
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
		return vcsGit, nil
	case P4:
		return vcsP4, nil
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

	repoName := filepath.Base(dir)

	if syncErr := repo.Sync(repoOwner, repoName); syncErr != nil {

		// Check if the error is due to a missing remote...
		errStr := syncErr.Error()
		missingLocalRemote := strings.Contains(errStr, "Could not read from remote repository")
		missingRemote, err := regexp.MatchString("fatal: repository '.*' not found.*", errStr)
		if err != nil {
			return errors.Annotate(err, syncErr.Error())
		}

		// ...if not, exit...
		if !missingRemote && !missingLocalRemote {

			if strings.Contains(errStr, "src refspec HEAD does not match any") {

				return errors.New("This looks like a new repository. lingo Requires at least one commit to exist in the repository. Please make your first commit and try again.")

			}

			return errors.Trace(syncErr)
		}

		// ...otherwise attempt to set up the remote.
		if vcsType == Git {

			// create a new remote repo
			repoName, err = CreateRepo(repo, repoName)
			if err != nil {
				return errors.Trace(err)
			}

			// [re]set the local remote with the new remote repo
			_, _, err = repo.SetRemote(repoOwner, repoName)
			if err != nil {
				return errors.Trace(err)
			}

			// attempt to sync again
			if err := repo.Sync(repoOwner, repoName); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func CreateRepo(repo Repo, name string) (string, error) {
	if err := repo.CreateRemote(name); err != nil {
		if util.IsRepoExistsError(errors.Cause(err)) {
			parts := strings.Split(name, "-")
			num := parts[len(parts)-1]

			// We ignore the error here because the only case in which Atoi
			// would error is if the name had not yet been appended with -n.
			// In this case, n will be set to zero which is what we Require.
			n, _ := strconv.Atoi(num)
			if n != 0 {
				// Need to remove existing trailing number where present,
				// otherwise the repoName only appends rather than replaces
				// and will produce names of the pattern "myPkg-1-2-...-n-n+1".
				name = strings.TrimSuffix(name, "-"+num)
			}
			return CreateRepo(repo, fmt.Sprintf("%s-%d", name, n+1))
		}
		return "", errors.Trace(err)
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
		return "", errors.New("Please run `lingo config setup`")
	}
	return repoOwner, nil
}
