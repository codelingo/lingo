package grpc

import (
	"io"

	// kitot "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/grpc/query"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// MakeQueryEndpointFactory returns a loadbalancer.Factory that transforms GRPC
// host:port strings into Endpoints that call the Query method on a GRPC server
// at that address.
func MakeQueryEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"codelingo.CodeLingo",
			"Query",
			encodeQueryRequest,
			decodeQueryResponse,
			query.QueryReply{},
			// grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

func MakeReviewEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
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
