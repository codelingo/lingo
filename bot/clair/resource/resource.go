// The resource package is responsible for making sure that the platform will have the right resources
// available to run the query. In particular, it makes sure that the base lexicon is able to pull
// the VCS resource that will be specified in the query stub.
package resource

import (
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/codelingo/lib/gh"
	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"
	lexvcs "github.com/codelingo/platform/backend/lexicon/vcs"
	"github.com/codelingo/platform/util"
	"github.com/juju/errors"
)

// RepoOpts is a host specific git repository identifier
type RepoOpts struct {
	Host  string
	Owner string
	Name  string
	// HostName is just a way to identify the host that can be used as a file path,
	// ie, with "." replaced with "_".
	HostName string
}

// PROpts is a global cross-host git pull request identifier
type PROpts struct {
	Repo RepoOpts
	PRID int
}

// PullRepo pulls the repo specified in the review request locally so that tenets can be extracted.
// TODO: generalize backing type - remove references to git
func PrepareRepo(req *PROpts) (backing.Repo, error) {
	// TODO: repo type should depend on request
	repo := vcs.New(backing.Git)

	if req != nil {
		err := GetPullRequest(repo, req)
		if err != nil {
			return repo, errors.Trace(err)
		}
	}

	if err := vcs.InitRepo(backing.Git); err != nil {
		// TODO(waigani) use error types
		// Note: Prior repo init is a valid state.
		if !strings.Contains(err.Error(), "already exists") {
			return repo, errors.Trace(err)
		}
	}
	return repo, nil
}

// GetPullRequest is responsible for taking a pull request definition and pulling it down into a local
// repo.
func GetPullRequest(backingRepo backing.Repo, req *PROpts) error {
	switch lexvcs.Host(req.Repo.HostName) {
	case lexvcs.LOCALHOST:
		return errors.New("pull requests are not supported on local repositories")
	case lexvcs.GITHUBHOST:
		// TODO(blakemscurr) ensure that there is a github api token available on the client/bot layer.
		// get store config options
		cfg, err := util.Config()
		if err != nil {
			return errors.Trace(err)
		}

		// Set up the API client and get the pull request object.
		token, err := cfg.GetValue("service.github_api.token")
		if err != nil {
			return errors.Trace(err)
		}

		// TODO(blakemscurr) move pull logic into vcs/host. Then implement for github, mock and bitbucket.
		c := gh.NewClient(token)
		repo := c.Repo(req.Repo.Owner, req.Repo.Name)
		pr, err := repo.Pull(req.PRID)
		if err != nil {
			return errors.Trace(err)
		}

		// Then, build the review scope from it's diff.
		diff, err := c.Diff(pr.DiffURL)
		if err != nil {
			return errors.Trace(err)
		}

		path, err := localRepoPath()
		if err != nil {
			return errors.Trace(err)
		}

		// TODO(waigani) we need to add Auth to this for private repos - ssh key.
		if err := os.MkdirAll(path, 0755); err != nil {
			return errors.Trace(err)
		}

		return errors.Trace(clone(backingRepo, req, diff.Raw))
	}

	return errors.Errorf("unknown host %q", req.Repo.Host)
}

// Clone clones from a host and applies unstaged changes.
// TODO: combine git libraries for client, bot endpoint layer, and lexicon. It is unnecessarily
// difficult to move logic between inconsistent packages that are supposed to do the same thing.
// This logic exists in the lexicon layer, but had to be altered to fit the structure in the bot/client.
func clone(repo backing.Repo, req *PROpts, diff string) error {
	path, err := localRepoPath()
	if err != nil {
		return errors.Trace(err)
	}

	ownerPath := filepath.Join(path, req.Repo.HostName, req.Repo.Owner)
	repoPath := filepath.Join(ownerPath, req.Repo.Name)

	// checkout the base repo once
	_, err = os.Stat(repoPath)

	if os.IsNotExist(err) {

		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return errors.Trace(err)
		}

		err := os.Chdir(repoPath)
		if err != nil {
			return errors.Trace(err)
		}

		url, err := repoURL(req)
		if err != nil {
			return errors.Trace(err)
		}

		// TODO: Build and use a repo struct to apply from
		err = repo.Clone(ownerPath, url)
		if err != nil {
			return errors.Trace(err)
		}
	} else if err == nil {
		err := os.Chdir(repoPath)
		if err != nil {
			return errors.Trace(err)
		}

		err = repo.ClearChanges()
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		return errors.Trace(err)
	}
	return errors.Trace(repo.ApplyPatch(diff))
}

func localRepoPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Trace(err)
	}
	return usr.HomeDir + "/.codelingo/repos", nil
}

// repoUrl produces a url from which to pull code from the repo specified in the opts.
func repoURL(req *PROpts) (string, error) {
	// TODO: allow multiple protocols and ports
	path, err := urlJoin("https://"+req.Repo.Host, req.Repo.Owner, req.Repo.Name)
	return path, errors.Trace(err)
}

func urlJoin(base string, paths ...string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", errors.Trace(err)
	}
	pathsStr := path.Join(paths...)
	u.Path = path.Join(u.Path, pathsStr)
	return u.String(), nil
}
