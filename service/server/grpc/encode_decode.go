package grpc

import (
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"golang.org/x/net/context"
)

func DecodePathsFromOffsetRequest(ctx context.Context, req interface{}) (interface{}, error) {
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

func DecodeListFactsRequest(ctx context.Context, req interface{}) (interface{}, error) {
	listFactsRequest := req.(*codelingo.ListFactsRequest)
	return &codelingo.ListFactsRequest{
		Owner:   listFactsRequest.Owner,
		Name:    listFactsRequest.Name,
		Version: listFactsRequest.Version,
	}, nil
}

func EncodeListFactsResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	facts := resp.(codelingo.FactList).Facts
	return &codelingo.FactList{
		Facts: facts,
	}, nil
}

func DecodeDescribeFactRequest(ctx context.Context, req interface{}) (interface{}, error) {
	request := req.(*codelingo.DescribeFactRequest)
	return &server.DescribeFactRequest{
		Owner:   request.Owner,
		Name:    request.Name,
		Version: request.Version,
		Fact:    request.Fact,
	}, nil
}

func EncodeDescribeFactResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	response := resp.(server.DescribeFactResponse)

	properties := []*codelingo.Property{}
	for _, prop := range response.Properties {
		properties = append(properties, &codelingo.Property{
			Name:        prop.Name,
			Description: prop.Description,
		})
	}

	return &codelingo.DescribeFactReply{
		Examples:    response.Examples,
		Description: response.Description,
		Properties:  properties,
	}, nil
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
		Dir:           reviewRequest.Dir,
	}, nil
}

func EncodeReviewResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return codelingo.ReviewReply{}, nil
}
