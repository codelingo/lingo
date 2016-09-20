package review

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

type RepoOpts struct {
	Host      string
	RepoOwner string
	RepoName  string
}

type PROpts struct {
	RepoOpts
	PRID int
}

// TODO(waigani) add bitbucket PR
func parseGithubPR(urlStr string) (*PROpts, error) {
	result, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Trace(err)
	}
	parts := strings.Split(strings.Trim(result.Path, "/"), "/")
	if l := len(parts); l != 4 {
		return nil, errors.Errorf("pull request URL path should have four parts, found %d", l)
	}

	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &PROpts{
		RepoOpts: RepoOpts{
			Host:      result.Host,
			RepoOwner: parts[0],
			RepoName:  parts[1],
		},
		PRID: n,
	}, nil
}
