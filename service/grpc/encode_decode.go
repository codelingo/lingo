package grpc

import (
	"golang.org/x/net/context"

	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
)

func encodeSessionRequest(ctx context.Context, request interface{}) (interface{}, error) {
	return &codelingo.SessionRequest{}, nil
}

func decodeSessionResponse(ctx context.Context, response interface{}) (interface{}, error) {
	resp := response.(*codelingo.SessionReply)
	return server.SessionResponse{
		Key: resp.Key,
	}, nil
}

func encodeQueryRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(server.QueryRequest)
	return &codelingo.QueryRequest{
		Clql: req.CLQL,
	}, nil
}

func decodeQueryResponse(ctx context.Context, response interface{}) (interface{}, error) {
	resp := response.(*codelingo.QueryReply)
	return server.QueryResponse{
		Result: resp.Result,
	}, nil
}

func encodeReviewRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(*server.ReviewRequest)
	return &codelingo.ReviewRequest{
		Key:           req.Key,
		Host:          req.Host,
		Owner:         req.Owner,
		Repo:          req.Repo,
		Sha:           req.SHA,
		FilesAndDirs:  req.FilesAndDirs,
		Recursive:     req.Recursive,
		Patches:       req.Patches,
		IsPullRequest: req.IsPullRequest,
		PullRequestID: int64(req.PullRequestID),
		Vcs:           req.Vcs,
		Dotlingo:      req.Dotlingo,
	}, nil
}

func decodeReviewResponse(ctx context.Context, response interface{}) (interface{}, error) {
	return server.ReviewResponse{}, nil
}
