// CLAIR is responsiblle for co-ordinating reviews on repositories.
package clair

import (
	"github.com/codelingo/lingo/bot/clair/query"
	"github.com/codelingo/lingo/bot/clair/resource"
	"github.com/codelingo/lingo/bot/clair/result"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/juju/errors"
)

// Request is the object that will be sent to the bot layer
type Request struct {
	PullRequest *resource.PROpts
	// TODO: do we need to push a diff across?
	DotLingo string // ctx.Bool("lingo-file")
}

// Review emulates the endpoint that will exist on the CLAIR bot, taking a single review request and
// streaming back a set of issues.
func Review(req Request) (chan *codelingo.Issue, error) {
	var queries []string
	var err error

	repo, err := resource.PrepareRepo(req.PullRequest)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if req.DotLingo == "" {
		var host string
		if req.PullRequest != nil {
			host = req.PullRequest.Repo.Host
		} else {
			host = "local"
		}

		// TODO: return a channel of queries
		queries, err = query.BuildQueries(repo, host)
		if err != nil {
			return nil, errors.Trace(err)
		}
	} else {
		queries = []string{req.DotLingo}
	}

	resultc, err := queryPlatform(queries)
	if err != nil {
		return nil, errors.Trace(err)
	}

	issuec, err := result.RouteIssues(resultc)
	return issuec, errors.Trace(err)
}

func queryPlatform(dotlingos []string) (chan *codelingo.QueryReply, error) {
	queryc := make(chan *codelingo.QueryRequest)
	go func() {
		for _, dl := range dotlingos {
			queryc <- &codelingo.QueryRequest{Dotlingo: dl}
		}
		close(queryc)
	}()

	resultc, err := service.Query(queryc)
	return resultc, errors.Trace(err)
}
