package grpc

import (
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"golang.org/x/net/context"
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

func encodeListLexiconsRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return &codelingo.ListLexiconsRequest{}, nil
}

func decodeListLexiconsResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	lexicons := resp.(*codelingo.ListLexiconsReply).Lexicons
	return &codelingo.ListLexiconsReply{
		Lexicons: lexicons,
	}, nil
}

func encodeListFactsRequest(ctx context.Context, req interface{}) (interface{}, error) {
	listFactsRequest := req.(codelingo.ListFactsRequest)
	return &codelingo.ListFactsRequest{
		Owner:   listFactsRequest.Owner,
		Name:    listFactsRequest.Name,
		Version: listFactsRequest.Version,
	}, nil
}

func decodeListFactsResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	facts := resp.(*codelingo.FactList).Facts
	factList := map[string][]string{}

	for k, v := range facts {
		if v.Child == nil {
			factList[k] = []string{}
		} else {
			factList[k] = v.Child
		}
	}
	return factList, nil
}

func encodeDescribeFactRequest(ctx context.Context, req interface{}) (interface{}, error) {
	request := req.(server.DescribeFactRequest)
	return &codelingo.DescribeFactRequest{
		Owner:   request.Owner,
		Name:    request.Name,
		Version: request.Version,
		Fact:    request.Fact,
	}, nil
}

func decodeDescribeFactResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	response := resp.(*codelingo.DescribeFactReply)

	properties := []server.Property{}
	for _, prop := range response.Properties {
		properties = append(properties, server.Property{
			Name:        prop.Name,
			Description: prop.Description,
		})
	}

	return &server.DescribeFactResponse{
		Examples:    response.Examples,
		Description: response.Description,
		Properties:  properties,
	}, nil
}

func encodePathsFromOffsetRequest(ctx context.Context, req interface{}) (interface{}, error) {
	pathsRequest := req.(*server.PathsFromOffsetRequest)
	return &codelingo.PathsFromOffsetRequest{
		Lang:     pathsRequest.Lang,
		Dir:      pathsRequest.Dir,
		Filename: pathsRequest.Filename,
		Src:      pathsRequest.Src,
		Start:    int64(pathsRequest.Start),
		End:      int64(pathsRequest.End),
	}, nil
}

func decodePathsFromOffsetResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	pathsResponse := resp.(*codelingo.PathsFromOffsetReply)
	paths := [][]string{}
	for _, path := range pathsResponse.Paths {
		paths = append(paths, path.Nodes)
	}
	return server.PathsFromOffsetResponse{
		Paths: paths,
	}, nil
}
