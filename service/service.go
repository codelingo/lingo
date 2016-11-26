package service

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/codelingo/kit/sd"
	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"

	"github.com/opentracing/opentracing-go"
	// zipkin "github.com/openzipkin/zipkin-go-opentracing"
	// appdashot "github.com/sourcegraph/appdash/opentracing"
	"golang.org/x/net/context"

	"github.com/codelingo/kit/endpoint"
	"github.com/codelingo/kit/log"
	"github.com/codelingo/kit/pubsub/rabbitmq"
	loadbalancer "github.com/codelingo/kit/sd/lb"
	grpcclient "github.com/codelingo/lingo/service/grpc"

	"github.com/codelingo/lingo/service/grpc/codelingo"

	// kitot "github.com/codelingo/kit/tracing/opentracing"

	"github.com/codelingo/lingo/service/server"
)

type client struct {
	context.Context
	log.Logger
	endpoints map[string]endpoint.Endpoint
}

// isEnd returns true if a buffer contains only a single null byte,
// indicating that the queue will have no further messages
func isEnd(b []byte) bool {
	return len(b) == 1 && b[0] == '\x00'
}

// TODO(pb): If your service interface methods don't return an error, we have
// no way to signal problems with a service client. If they don't take a
// context, we have to provide a global context for any transport that
// requires one, effectively making your service a black box to any context-
// specific information. So, we should make some recommendations:
//
// - To get started, a simple service interface is probably fine.
//
// - To properly deal with transport errors, every method on your service
//   should return an error. This is probably important.
//
// - To properly deal with context information, every method on your service
//   can take a context as its first argument. This may or may not be
//   important.

func (c client) Session(req *server.SessionRequest) (string, error) {
	reply, err := c.endpoints["session"](c.Context, req)
	if err != nil {
		return "", err
	}

	r := reply.(server.SessionResponse)
	return r.Key, nil
}

func (c client) Query(clql string) (string, error) {
	request := server.QueryRequest{
		CLQL: clql,
	}
	reply, err := c.endpoints["query"](c.Context, request)
	if err != nil {
		// c.Logger.Log("err", err)
		return "", err
	}

	r := reply.(server.QueryResponse)
	return r.Result, nil
}

func (c client) ListFacts(lexicon string) (map[string][]string, error) {
	request := codelingo.ListFactsRequest{
		Lexicon: lexicon,
	}
	reply, err := c.endpoints["listfacts"](c.Context, request)
	if err != nil {
		return nil, err
	}

	r := reply.(map[string][]string)

	return r, nil

}

func (c client) ListLexicons() ([]string, error) {
	request := codelingo.ListLexiconsRequest{}
	reply, err := c.endpoints["listlexicons"](c.Context, request)
	if err != nil {
		return nil, err
	}
	r := reply.(*codelingo.ListLexiconsReply)
	return r.Lexicons, nil
}

func (c client) Review(req *server.ReviewRequest) (server.Issuec, server.Messagec, error) {
	// set defaults
	if req.Host == "" {
		return nil, nil, errors.New("repository host is not set")
	}
	if req.Owner == "" {
		return nil, nil, errors.New("repository owner is not set")
	}
	if req.Repo == "" {
		return nil, nil, errors.New("repository name is not set")
	}

	// Initialise review session and receive channel prefix
	prefix, err := c.Session(&server.SessionRequest{})
	req.Key = prefix

	platConfig, err := config.Platform()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	mqAddress, err := platConfig.MessageQueueAddr()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	issueSubscriber, err := rabbitmq.NewSubscriber(mqAddress, prefix+"-issues", "")
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	messageSubscriber, err := rabbitmq.NewSubscriber(mqAddress, prefix+"-messages", "")
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	issueSubc := issueSubscriber.Start()
	messageSubc := messageSubscriber.Start()
	issuec := make(server.Issuec)
	messagec := make(server.Messagec)

	// helper func to send errors to the message chan
	sendErrIfErr := func(err error) bool {
		if err != nil {
			if err2 := messagec.Send(err.Error()); err2 != nil {
				// yes panic, this is a developer error
				panic(errors.Annotate(err, err2.Error()))
			}
			return true
		}
		return false
	}

	// read from subscriber chans onto issue and message chans, transforming
	// pubsub.Message into codelingo.Issue or server.Message
	go func() {
		defer close(messagec)
		defer close(issuec)
		defer issueSubscriber.Stop()
		defer messageSubscriber.Stop()
	l:
		for {
			select {
			case issueMsg, ok := <-issueSubc:
				// TODO(waigani) !ok is never used and isEnd is a workarond. A
				// proper pubsub should close the chan upstream.
				byt, err := ioutil.ReadAll(issueMsg)
				if sendErrIfErr(err) {
					break l
				}
				if !ok || isEnd(byt) {
					// no more issues.
					break l
				}

				issue := &codelingo.Issue{}
				if sendErrIfErr(json.Unmarshal(byt, issue)) ||
					sendErrIfErr(issuec.Send(issue)) ||
					sendErrIfErr(issueMsg.Done()) {
					break l
				}

			case msg, ok := <-messageSubc:
				byt, err := ioutil.ReadAll(msg)
				if sendErrIfErr(err) ||
					!ok ||
					isEnd(byt) {
					// no more messages.
					break l
				}

				// TODO: Process messages
				sendErrIfErr(messagec.Send(string(byt) + "\n"))
				sendErrIfErr(msg.Done())
				// TODO(waigani) DEMOWARE setting to 600
			case <-time.After(time.Second * 600):
				sendErrIfErr(errors.New("timed out waiting for issues x"))
				break l
			}
		}
	}()

	_, err = c.endpoints["review"](c.Context, req)
	if err != nil {
		return nil, nil, err
	}

	return issuec, messagec, nil
}

// TODO(waigani) construct logger separately and pass into New.
// TODO(waigani) swap os.Exit(1) for return err

// NewClient returns a QueryService that's backed by the provided Endpoints
func New() (server.CodeLingoService, error) {
	pCfg, err := config.Platform()
	if err != nil {
		return nil, errors.Trace(err)
	}

	grpcAddr, err := pCfg.GrpcAddress()
	if err != nil {
		return nil, errors.Trace(err)
	}

	var (
		grpcAddrs = grpcAddr //flag.String("grpc.addrs", grpcAddr, "Comma-separated list of addresses for gRPC servers")

		// Three OpenTracing backends (to demonstrate how they can be interchanged):
		//	zipkinAddr           = flag.String("zipkin.kafka.addr", "", "Enable Zipkin tracing via a Kafka Collector host:port")
		// appdashAddr          = flag.String("appdash.addr", "", "Enable Appdash tracing via an Appdash server host:port")
		// lightstepAccessToken = flag.String("lightstep.token", "", "Enable LightStep tracing via a LightStep access token")
	)
	// flag.Parse()
	// if len(os.Args) < 2 {
	// 	fmt.Fprintf(os.Stderr, "\n%s [flags] method\n\n", filepath.Base(os.Args[0]))
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	randomSeed := time.Now().UnixNano()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	logger = log.NewContext(logger).With("transport", "grpc")
	tracingLogger := log.NewContext(logger).With("component", "tracing")

	// Set up OpenTracing
	var tracer opentracing.Tracer
	// {
	// 	switch {
	// 	case *appdashAddr != "" && *lightstepAccessToken == "" && *zipkinAddr == "":
	// 		tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
	// 	case *appdashAddr == "" && *lightstepAccessToken != "" && *zipkinAddr == "":
	// 		tracer = lightstep.NewTracer(lightstep.Options{
	// 			AccessToken: *lightstepAccessToken,
	// 		})
	// 		defer lightstep.FlushLightStepTracer(tracer)
	// 	case *appdashAddr == "" && *lightstepAccessToken == "" && *zipkinAddr != "":
	// 		collector, err := zipkin.NewKafkaCollector(
	// 			strings.Split(*zipkinAddr, ","),
	// 			zipkin.KafkaLogger(tracingLogger),
	// 		)
	// 		if err != nil {
	// 			tracingLogger.Log("err", "unable to create kafka collector", "fatal", err)
	// 			os.Exit(1)
	// 		}
	// 		tracer, err = zipkin.NewTracer(
	// 			zipkin.NewRecorder(collector, false, "localhost:8000", "querysvc-client"),
	// 		)
	// 		if err != nil {
	// 			tracingLogger.Log("err", "unable to create zipkin tracer", "fatal", err)
	// 			os.Exit(1)
	// 		}
	// 	case *appdashAddr == "" && *lightstepAccessToken == "" && *zipkinAddr == "":
	// 		tracer = opentracing.GlobalTracer() // no-op
	// 	default:
	// 		tracingLogger.Log("fatal", "specify a single -appdash.addr, -lightstep.access.token or -zipkin.kafka.addr")
	// 		os.Exit(1)
	// 	}
	// }

	instances := strings.Split(grpcAddrs, ",")
	sessionFactory := grpcclient.MakeSessionEndpointFactory(tracer, tracingLogger)
	queryFactory := grpcclient.MakeQueryEndpointFactory(tracer, tracingLogger)
	reviewFactory := grpcclient.MakeReviewEndpointFactory(tracer, tracingLogger)
	listFactsFactory := grpcclient.MakeListFactsEndpointFactory(tracer, tracingLogger)
	listLexiconsFactory := grpcclient.MakeListLexiconsEndpointFactory(tracer, tracingLogger)

	return client{
		Context: context.Background(),
		Logger:  logger,
		endpoints: map[string]endpoint.Endpoint{
			// TODO(waigani) this could be refactored further, a lot of dups
			"session":      buildEndpoint(tracer, "session", instances, sessionFactory, randomSeed, logger),
			"query":        buildEndpoint(tracer, "query", instances, queryFactory, randomSeed, logger),
			"review":       buildEndpoint(tracer, "review", instances, reviewFactory, randomSeed, logger),
			"listfacts":    buildEndpoint(tracer, "listfacts", instances, listFactsFactory, randomSeed, logger),
			"listlexicons": buildEndpoint(tracer, "listlexicons", instances, listLexiconsFactory, randomSeed, logger),
		},
	}, nil
}

// TODO(waigani) this needs to be refactored. It was using the original go-kit
// pattern - which then changed, a lot. Below is a quick fix to get the client
// working, but we need to rewrite grpc client/server to use the new go-kit
// patterns.
func buildEndpoint(tracer opentracing.Tracer, operationName string, instances []string, factory sd.Factory, seed int64, logger log.Logger) endpoint.Endpoint {
	// publisher := static.NewPublisher(instances, factory, logger)

	var endpoints []endpoint.Endpoint
	for _, inst := range instances {
		ep, closer, err := factory(inst)
		if err != nil {
			logger.Log(err)
		}
		// TODO(waigani) when do we close?
		_ = closer
		endpoints = append(endpoints, ep)
	}

	subscriber := sd.FixedSubscriber(endpoints)
	random := loadbalancer.NewRandom(subscriber, seed)
	endpoint := loadbalancer.Retry(10, 10*time.Second, random)
	return endpoint
	// return kitot.TraceClient(tracer, operationName)(endpoint)
}
