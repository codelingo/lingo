package result

import (
	"fmt"

	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/juju/errors"
)

// RouteIssues listens on a result channel, turns them into issues, and fans the out onto channels
// to be sent to any destination - comment on a PR, back to the user's CLI, into a report maker etc.
func RouteIssues(resultc chan *codelingo.QueryReply) (chan *codelingo.Issue, error) {
	// TODO: allow multiple destinations
	issuec := make(chan *codelingo.Issue)
	go func() {
		defer close(issuec)
		for {
			select {
			case result, ok := <-resultc:
				if !ok {
					return
				}

				issue, err := buildIssue(result)
				if err != nil {
					fmt.Println(errors.Trace(err))
				}

				issuec <- issue
			}
		}
	}()
	return issuec, nil
}
