package grpc

import (
	"io"

	// kitot "github.com/codelingo/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/codelingo/kit/endpoint"
	"github.com/codelingo/kit/log"
	"github.com/codelingo/kit/sd"
	grpctransport "github.com/codelingo/kit/transport/grpc"
	"github.com/codelingo/lingo/service/grpc/codelingo"
)

func MakeSessionEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"Session",
			encodeSessionRequest,
			decodeSessionResponse,
			codelingo.SessionReply{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

// MakeQueryEndpointFactory returns a loadbalancer.Factory that transforms GRPC
// host:port strings into Endpoints that call the Query method on a GRPC server
// at that address.
func MakeQueryEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"Query",
			encodeQueryRequest,
			decodeQueryResponse,
			codelingo.QueryReply{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

func MakeReviewEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"Review",
			encodeReviewRequest,
			decodeReviewResponse,
			codelingo.ReviewReply{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

func MakeListFactsEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"ListFacts",
			encodeListFactsRequest,
			decodeListFactsResponse,
			codelingo.FactList{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

func MakeListLexiconsEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"ListLexicons",
			encodeListLexiconsRequest,
			decodeListLexiconsResponse,
			codelingo.ListLexiconsReply{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}
