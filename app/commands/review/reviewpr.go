// The review package contains helper methods to create review requests to be sent to the bot
// endpoint layer, especially CLAIR.
package review

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/codelingo/lingo/bot/clair/resource"
	"github.com/juju/errors"
)

// ParsePR produces a set of PROpts to be sent to CLAIR in the bot endpoint layer
func ParsePR(urlStr string) (*resource.PROpts, error) {
	// TODO: Try to parse urlStr as a request for each VCS host
	opts, err := parseGithubPR(urlStr)
	return opts, errors.Trace(err)
}

func parseGithubPR(urlStr string) (*resource.PROpts, error) {
	result, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var parts []string
	parts = strings.Split(strings.Trim(result.Path, "/"), "/")
	if l := len(parts); l != 4 {
		return nil, errors.Errorf("pull request URL needs to be in the following format: https://github.com/<username>/<repo_name>/pull/<pull_number>")
	}

	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, errors.Trace(err)
	}

	host := strings.Replace(result.Host, ".", "_", -1)
	if host != "github_com" {
		return nil, errors.New("Github PRs must be made to github.com.")
	}

	return &resource.PROpts{
		Repo: resource.RepoOpts{
			Host:     result.Host,
			Owner:    parts[0],
			Name:     parts[1],
			HostName: host,
		},
		PRID: n,
	}, nil
}

// TODO: implement
func parseBitBucketPR(urlStr string) (*resource.PROpts, error) {
	return nil, errors.New("BitBucket not supported.")
}

// TODO: implement
func parseGitlabPR(urlStr string) (*resource.PROpts, error) {
	return nil, errors.New("Gitlab not supported.")
}
