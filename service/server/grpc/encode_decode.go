package grpc

import (
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"golang.org/x/net/context"
)

func DecodePathsFromOffsetRequest(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO(BlakeMScurr) is it sensible to decode this into a src.Fact here?
	pathsRequest := req.(*codelingo.PathsFromOffsetRequest)
	return &server.PathsFromOffsetRequest{
		Lang:     pathsRequest.Lang,
		Dir:      pathsRequest.Dir,
		Filename: pathsRequest.Filename,
		Src:      pathsRequest.Src,
		Start:    int(pathsRequest.Start),
		End:      int(pathsRequest.End),
	}, nil
}

func EncodePathsFromOffsetResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	// may not populate properly, because other encode decode methods don't do it this way
	pathsResponse := resp.(*server.PathsFromOffsetResponse)
	response := codelingo.PathsFromOffsetReply{
		Paths: []*codelingo.Path{},
	}
	for _, path := range pathsResponse.Paths {
		response.Paths = append(response.Paths, &codelingo.Path{
			Nodes: path,
		})
	}

	return response, nil
}

func DecodeListFactsRequest(ctx context.Context, req interface{}) (interface{}, error) {
	lexicon := req.(*codelingo.ListFactsRequest).Lexicon
	return &codelingo.ListFactsRequest{
		Lexicon: lexicon,
	}, nil
}

func EncodeListFactsResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	facts := resp.(codelingo.FactList).Facts
	return &codelingo.FactList{
		Facts: facts,
	}, nil
}

func DecodeListLexiconsRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return &codelingo.ListLexiconsRequest{}, nil
}

func EncodeListLexiconsResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	lexicons := resp.(codelingo.ListLexiconsReply).Lexicons
	return &codelingo.ListLexiconsReply{
		Lexicons: lexicons,
	}, nil
	return &codelingo.ListLexiconsReply{}, nil
}

func DecodeSessionRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return &server.SessionRequest{}, nil
}

func EncodeSessionResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	key := resp.(string)

	return &codelingo.SessionReply{
		Key: key,
	}, nil
}

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
		Key:           reviewRequest.Key,
		Host:          reviewRequest.Host,
		Owner:         reviewRequest.Owner,
		Repo:          reviewRequest.Repo,
		SHA:           reviewRequest.Sha,
		FilesAndDirs:  reviewRequest.FilesAndDirs,
		Recursive:     reviewRequest.Recursive,
		Patches:       reviewRequest.Patches,
		IsPullRequest: reviewRequest.IsPullRequest,
		Vcs:           reviewRequest.Vcs,
		Dotlingo:      reviewRequest.Dotlingo,
		PullRequestID: int(reviewRequest.PullRequestID),
	}, nil
}

func EncodeReviewResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return codelingo.ReviewReply{}, nil
}
