package query

import (
	"strings"

	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"
	"github.com/juju/errors"
)

// Build queries analyzes the current repo and returns a set of queries extracted from
// the .lingo files. The query should give all the necessary information for the backend
// vcs lexicon to get resources, so it needs to refer to the host to pull the code from.
func BuildQueries(repo backing.Repo, host string) ([]string, error) {
	var dotlingos []string
	var err error

	// TODO: replace this system with nfs-like communication.
	if err = vcs.SyncRepo(repo); err != nil {
		return nil, errors.Trace(err)
	}

	dotlingos, err = repo.BuildQueries(host)
	if err != nil {
		if noCommitErr(err) {
			return nil, errors.New(noCommitErrMsg)
		}

		return nil, errors.Annotate(err, "\nbad request")
	}
	return dotlingos, nil
}

const noCommitErrMsg = "This looks like a new repository. Please make an initial commit before running `lingo review`. This is only required for the initial commit, subsequent changes to your repo will be picked up by lingo without committing."

// TODO(waigani) use typed error
func noCommitErr(err error) bool {
	return strings.Contains(err.Error(), "ambiguous argument 'HEAD'")
}
