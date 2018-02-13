package grpc

import (
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"github.com/juju/errors"
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

// Query is bidirectionally streamed making this irrelevant
func encodeQueryRequest(ctx context.Context, request interface{}) (interface{}, error) {
	return nil, nil
}

func decodeQueryResponse(ctx context.Context, response interface{}) (interface{}, error) {
	resp := response.(*codelingo.QueryReply)
	return resp, nil
}

func encodeReviewRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(*server.ReviewRequest)
	return &codelingo.ReviewRequest{
		Host:          req.Host,
		Owner:         req.Owner,
		Repo:          req.Repo,
		Sha:           req.SHA,
		Patches:       req.Patches,
		IsPullRequest: req.IsPullRequest,
		PullRequestID: int64(req.PullRequestID),
		Vcs:           req.Vcs,
		Dotlingo:      req.Dotlingo,
		Dir:           req.Dir,
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

func encodeLatestClientVersionRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return &codelingo.LatestClientVersionRequest{}, nil
}

func decodeLatestClientVersionResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	version := resp.(*codelingo.LatestClientVersionReply).Version
	return &codelingo.LatestClientVersionReply{
		Version: version,
	}, nil
}

// Not related to kit
func ToPropData(prop interface{}) (*codelingo.DataField, error) {
	res := &codelingo.DataField{}
	switch p := prop.(type) {
	case string:
		res.Prop = &codelingo.DataField_StringProp{
			StringProp: p,
		}
	case bool:
		res.Prop = &codelingo.DataField_BoolProp{
			BoolProp: p,
		}
	case int:
		res.Prop = &codelingo.DataField_Int64Prop{
			Int64Prop: int64(p),
		}
	case float32:
		res.Prop = &codelingo.DataField_FloatProp{
			FloatProp: p,
		}
	case float64:
		res.Prop = &codelingo.DataField_FloatProp{
			FloatProp: float32(p),
		}
	default:
		return nil, errors.Errorf("Invalid property type: %v", p)
	}
	return res, nil
}

func FromPropData(field *codelingo.DataField) (interface{}, error) {
	switch p := field.Prop.(type) {
	case *codelingo.DataField_StringProp:
		return p.StringProp, nil
	case *codelingo.DataField_BoolProp:
		return p.BoolProp, nil
	case *codelingo.DataField_Int64Prop:
		return int(p.Int64Prop), nil
	case *codelingo.DataField_FloatProp:
		return p.FloatProp, nil
	default:
		return nil, errors.Errorf("Invalid property type: %v", p)
	}
}
