// The review package contains helper methods to create review requests to be sent to the bot
// endpoint layer, especially CLAIR.
package review

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

// RepoOpts is a host specific git repository identifier
type RepoOpts struct {
	Host      string
	RepoOwner string
	RepoName  string
}

// PROpts is a global cross-host git pull request identifier
type PROpts struct {
	RepoOpts
	PRID int
	Host string
}

// ParsePR produces a set of PROpts to be sent to CLAIR in the bot endpoint layer
func ParsePR(urlStr string) (*PROpts, error) {
	// TODO: Try to parse urlStr as a request for each VCS host
	opts, err := parseGithubPR(urlStr)
	return opts, errors.Trace(err)
}

func parseGithubPR(urlStr string) (*PROpts, error) {
	result, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Trace(err)
	}

	parts := strings.Split(strings.Trim(result.Path, "/"), "/")
	if l := len(parts); l != 4 {
		return nil, errors.Errorf("pull request URL needs to be in the following format: https://github.com/<username>/<repo_name>/pull/<pull_number>")
	}

	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, errors.Trace(err)
	}

	host := strings.Replace(result.Host, ".", "_", -1)
	return &PROpts{
		RepoOpts: RepoOpts{
			Host:      host,
			RepoOwner: parts[0],
			RepoName:  parts[1],
		},
		PRID: n,
		Host: "Github",
	}, nil
}

// TODO: implement
func parseBitBucketPR(urlStr string) (*PROpts, error) {
	return nil, errors.New("BitBucket not supported.")
}

// TODO: implement
func parseGitlabPR(urlStr string) (*PROpts, error) {
	return nil, errors.New("Gitlab not supported.")
}
