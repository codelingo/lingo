package grpc

import (
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"golang.org/x/net/context"
)

func DecodeQueryRequest(ctx context.Context, req interface{}) (interface{}, error) {
	queryRequest := req.(*codelingo.QueryRequest)

	return &server.QueryRequest{
		CLQL: queryRequest.Clql,
	}, nil
}

func EncodeQueryResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	domainResponse := resp.(server.QueryResponse)

	return &codelingo.QueryReply{
		Result: domainResponse.Result,
	}, nil
}

func DecodeReviewRequest(ctx context.Context, req interface{}) (interface{}, error) {
	reviewRequest := req.(*codelingo.ReviewRequest)
	return &server.ReviewRequest{
		Host:          reviewRequest.Host,
		Owner:         reviewRequest.Owner,
		Repo:          reviewRequest.Repo,
		SHA:           reviewRequest.Sha,
		FilesAndDirs:  reviewRequest.FilesAndDirs,
		Recursive:     reviewRequest.Recursive,
		Patches:       reviewRequest.Patches,
		IsPullRequest: reviewRequest.IsPullRequest,
		Vcs:           reviewRequest.Vcs,
		PullRequestID: int(reviewRequest.PullRequestID),
	}, nil
}

func EncodeReviewResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	issues := resp.([]*codelingo.Issue)

	return codelingo.ReviewReply{
		Issues: issues,
	}, nil
}
