package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	endpointCtx "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/codelingo/kit/sd"
	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/service/serviceLogger"
	"github.com/juju/errors"

	"github.com/opentracing/opentracing-go"
	// zipkin "github.com/openzipkin/zipkin-go-opentracing"
	// appdashot "github.com/sourcegraph/appdash/opentracing"

	"time"

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

func Review(req *codelingo.ReviewRequest) (chan *codelingo.Issue, error) {
	cc, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}

	client := codelingo.NewCodeLingoClient(cc)

	newCtx, err := grpcclient.GetGcloudEndpointCtx()
	if err != nil {
		return nil, errors.Trace(err)
	}

	ctx, cancel := context.WithCancel(newCtx)
	// Cancel the context passed to the platform on exit.
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		cancel()
		os.Exit(1)
	}()

	stream, err := client.Review(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v.Query(_) = _, %v", client, err))
	}

	if err := stream.Send(req); err != nil {
		errors.Trace(err)
	}

	err = stream.CloseSend()
	if err != nil {
		errors.Trace(err)
	}

	issuec := make(chan *codelingo.Issue)
	go func() {
		for {
			in, err := stream.Recv()

			if err != nil {
				if err != io.EOF {
					issuec <- &codelingo.Issue{Err: err.Error()}
				}
				close(issuec)
				return
			}
			issuec <- in
		}
	}()

	return issuec, nil
}

func (c client) Query(ctx context.Context, queryc chan *codelingo.QueryRequest) (chan *codelingo.QueryReply, chan error) {
	return nil, nil
}

func Query(cc *grpc.ClientConn, queryc chan *codelingo.QueryRequest) (chan *codelingo.QueryReply, error) {
	client := codelingo.NewCodeLingoClient(cc)
	resultc, err := runQuery(client, queryc)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resultc, nil
}

const (
	FlowServer = "flowserver"
	// TODO: remove public access to the platform server
	PlatformServer = "platformserver"
	FlowClient     = "flowclient"
	LocalClient    = "localclient"
)

// GrpcConnection creates a connection between a given server and client type.
// TODO(BlakeMScurr): this should be moved into its own service repo, so that flow and platform don't have
// to depend on the client. The code pertaining specifically to the client side and flow side
// configs should be kept in the client/flow repos, and addresses and tls values should be
// passed as arguments.
func GrpcConnection(client, server string) (*grpc.ClientConn, error) {

	var grpcAddr string
	var isTLS bool
	switch client {
	case LocalClient:
		pCfg, err := config.Platform()
		if err != nil {
			return nil, errors.Trace(err)
		}
		strTLS, err := pCfg.GetValue("gitserver.tls")
		if err != nil {
			return nil, errors.Trace(err)
		}

		isTLS, err = strconv.ParseBool(strTLS)
		if err != nil {
			return nil, errors.Trace(err)
		}

		switch server {
		case FlowServer:
			grpcAddr, err = pCfg.FlowAddress()
			if err != nil {
				return nil, errors.Trace(err)
			}
		case PlatformServer:
			grpcAddr, err = pCfg.PlatformAddress()
			if err != nil {
				return nil, errors.Trace(err)
			}
		default:
			return nil, errors.Errorf("Unknown Server %s:", server)
		}
	case FlowClient:
		isTLS = false
		// TODO: this is hardcoded to platform address:port
		// create a flow config and read from that
		grpcAddr = "localhost:8002"
	}

	grpclog.SetLogger(serviceLogger.New())
	var tlsOpt grpc.DialOption

	if !isTLS {
		tlsOpt = grpc.WithInsecure()
	} else {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(cert)) {
			return nil, errors.New("credentials: failed to append certificates")
		}
		creds := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp})
		tlsOpt = grpc.WithTransportCredentials(creds)
	}
	// There may be multiple instances
	cc, err := grpc.Dial(grpcAddr, tlsOpt)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return cc, nil
}

func runQuery(client codelingo.CodeLingoClient, queryc chan *codelingo.QueryRequest) (chan *codelingo.QueryReply, error) {
	newCtx, err := grpcclient.GetGcloudEndpointCtx()
	if err != nil {
		return nil, errors.Trace(err)
	}

	ctx, cancel := context.WithCancel(newCtx)

	// Cancel the context passed to the platform on exit.
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		cancel()
		os.Exit(1)
	}()

	stream, err := client.Query(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v.Query(_) = _, %v", client, err))
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	resultc := make(chan *codelingo.QueryReply)
	go func() {
		wg.Wait()
		close(resultc)
	}()

	go func() {
		defer wg.Done()
		for {
			in, err := stream.Recv()

			if err != nil {
				if err != io.EOF {
					resultc <- &codelingo.QueryReply{Error: err.Error()}
				}
				return
			}
			resultc <- in
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case query, ok := <-queryc:
				if !ok {
					err = stream.CloseSend()
					if err != nil {
						resultc <- &codelingo.QueryReply{Error: err.Error()}
					}
					return
				}

				if err := stream.Send(query); err != nil {
					resultc <- &codelingo.QueryReply{Error: err.Error()}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultc, nil
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

func (c client) ListFacts(owner, name, version string) (map[string][]string, error) {
	request := codelingo.ListFactsRequest{
		Owner:   owner,
		Name:    name,
		Version: version,
	}
	reply, err := c.endpoints["listfacts"](c.Context, request)
	if err != nil {
		return nil, err
	}

	r := reply.(map[string][]string)

	return r, nil

}

func (c client) DescribeFact(owner, name, version, fact string) (*server.DescribeFactResponse, error) {
	request := server.DescribeFactRequest{
		Owner:   owner,
		Name:    name,
		Version: version,
		Fact:    fact,
	}

	reply, err := c.endpoints["describefact"](c.Context, request)
	if err != nil {
		return nil, err
	}
	response := reply.(*server.DescribeFactResponse)
	return response, nil
}

func (c client) PathsFromOffset(request *server.PathsFromOffsetRequest) (*server.PathsFromOffsetResponse, error) {
	reply, err := c.endpoints["pathsfromoffset"](c.Context, request)
	if err != nil {
		return nil, err
	}
	response := reply.(server.PathsFromOffsetResponse)
	return &response, nil
}

func cancelReview(sessionKey string) error {

	platConfig, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}
	mqAddress, err := platConfig.MessageQueueAddr()
	if err != nil {
		return errors.Trace(err)
	}
	cPub, err := rabbitmq.NewPublisher(mqAddress, sessionKey+"-cancel")
	if err != nil {
		return errors.Trace(err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.WriteRune('\x00')
	if err != nil {
		return err
	}

	err = cPub.Publish("", buf)
	if err != nil {
		return err
	}
	return nil
}

// TODO(waigani) context.Context is here to conform to CodeLingoService
// interface. The interface takes a context because the platform
// implementation needs it. Refactor so this client side func sig does not
// require a context.
func (c client) Review(_ context.Context, req *server.ReviewRequest) (server.Issuec, server.Messagec, server.Ingestc, error) {
	// set defaults

	// Initialise review session and receive channel prefix
	prefix, err := c.Session(&server.SessionRequest{})
	req.Key = prefix

	// Send a cancel signal to the platform on exit.
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		cancelReview(prefix)
		os.Exit(1)
	}()

	platConfig, err := config.Platform()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	mqAddress, err := platConfig.MessageQueueAddr()
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	// TODO: Make prefix+"-issues" and equivs constants as they are shared
	// between codelingo/platform and codelingo/lingo
	issueSubscriber, err := rabbitmq.NewSubscriber(mqAddress, prefix+"-issues", "")
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}
	messageSubscriber, err := rabbitmq.NewSubscriber(mqAddress, prefix+"-messages", "")
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	ingestSubscriber, err := rabbitmq.NewSubscriber(mqAddress, prefix+"-ingest-progress", "")
	if err != nil {
		return nil, nil, nil, errors.Trace(err)
	}

	issueSubc := issueSubscriber.Start()
	messageSubc := messageSubscriber.Start()
	ingestSubc := ingestSubscriber.Start()
	issuec := make(server.Issuec)
	messagec := make(server.Messagec)
	ingestc := make(server.Ingestc)

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

	// read from subscriber chans onto message, issue, and ingest chans,
	// transforming pubsub.Message into codelingo.Issue or server.Message
	// (or a straight string in the case of the working version of ingest chan)

	// TODO(waigani) !ok is never used and isEnd(byt) is a workarond.
	// A proper pubsub should close the chan upstream.
	go func() {
		defer close(messagec)
		defer messageSubscriber.Stop()
		defer close(issuec)
		defer issueSubscriber.Stop()

	l:
		for {
			select {
			case ingestProgress, ok := <-ingestSubc:
				byt1, err := ioutil.ReadAll(ingestProgress)
				if sendErrIfErr(err) ||
					isEnd(byt1) ||
					sendErrIfErr(ingestc.Send(string(byt1))) ||
					!ok {

					// no more ingestion updates.
					close(ingestc)
					ingestSubscriber.Stop()
					break l
				}

			case msg, ok := <-messageSubc:
				byt, err := ioutil.ReadAll(msg)
				if sendErrIfErr(err) ||
					!ok || isEnd(byt) {
					// no more messages.
					continue
				}

				if err := messagec.Send(string(byt)); err != nil {
					// yes panic, this is a developer error
					panic(err.Error())
				}
			}
		}

		finished := 0
		for {
			if finished >= 2 {
				break
			}
			select {
			case issueMsg, ok := <-issueSubc:
				byt, err := ioutil.ReadAll(issueMsg)
				if sendErrIfErr(err) || !ok || isEnd(byt) {
					// no more issues.
					finished++
					continue
				}

				issue := &codelingo.Issue{}
				if sendErrIfErr(json.Unmarshal(byt, issue)) ||
					sendErrIfErr(issuec.Send(issue)) ||
					sendErrIfErr(issueMsg.Done()) {
					return
				}

			case msg, ok := <-messageSubc:
				byt, err := ioutil.ReadAll(msg)
				if sendErrIfErr(err) ||
					!ok ||
					isEnd(byt) {
					// no more messages.
					finished++
					continue
				}

				if err := messagec.Send(string(byt)); err != nil {
					// yes panic, this is a developer error
					panic(err.Error())
				}

				// TODO(waigani) This needs refactoring. We don't we know this
				// is an error at this point.
				// err = userFacingErrs(errors.New(string(byt)))
				// sendErrIfErr(err)
				// sendErrIfErr(msg.Done())

				// TODO(waigani) DEMOWARE setting to 600
			case <-time.After(time.Second * 600):
				sendErrIfErr(errors.New("timed out waiting for issues"))
				return
			}
		}
	}()

	_, err = c.endpoints["review"](c.Context, req)
	if err != nil {
		return nil, nil, nil, err
	}

	return issuec, messagec, ingestc, nil
}

func (c client) LatestClientVersion() (string, error) {
	request := codelingo.LatestClientVersionRequest{}
	reply, err := c.endpoints["latestclientversion"](c.Context, request)
	if err != nil {
		return "", err
	}
	r := reply.(*codelingo.LatestClientVersionReply)
	return r.Version, nil
}

// TODO(waigani) construct logger separately and pass into New.
// TODO(waigani) swap os.Exit(1) for return err

// NewClient returns a QueryService that's backed by the provided Endpoints
func New() (server.CodeLingoService, error) {
	pCfg, err := config.Platform()
	if err != nil {
		return nil, errors.Trace(err)
	}

	grpcAddr, err := pCfg.PlatformAddress()
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

	// Setup a logger to catch the grpc disconnecting logs and replace with user friendly ones
	grpclog.SetLogger(serviceLogger.New())
	var tlsOpt grpc.DialOption
	platCfg, err := config.Platform()
	if err != nil {
		return nil, errors.Trace(err)
	}
	isTLS, err := platCfg.GetValue("gitserver.tls")
	if err != nil {
		return nil, errors.Trace(err)
	}
	if isTLS == "false" {
		tlsOpt = grpc.WithInsecure()
	} else {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(cert)) {
			return nil, errors.New("credentials: failed to append certificates")
		}
		creds := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp})
		tlsOpt = grpc.WithTransportCredentials(creds)
	}

	instances := strings.Split(grpcAddrs, ",")
	sessionFactory := grpcclient.MakeSessionEndpointFactory(tracer, tracingLogger, tlsOpt)
	queryFactory := grpcclient.MakeQueryEndpointFactory(tracer, tracingLogger, tlsOpt)
	reviewFactory := grpcclient.MakeReviewEndpointFactory(tracer, tracingLogger, tlsOpt)
	listFactsFactory := grpcclient.MakeListFactsEndpointFactory(tracer, tracingLogger, tlsOpt)
	listLexiconsFactory := grpcclient.MakeListLexiconsEndpointFactory(tracer, tracingLogger, tlsOpt)
	pathsFromOffsetFactory := grpcclient.MakePathsFromOffsetEndpointFactory(tracer, tracingLogger, tlsOpt)
	describeFactFactory := grpcclient.MakeDescribeFactEndpointFactory(tracer, tracingLogger, tlsOpt)
	latestClientVersionFactory := grpcclient.MakeLatestClientVersionFactory(tracer, tracingLogger, tlsOpt)

	newCtx, err := grpcclient.GetGcloudEndpointCtx()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return client{
		Context: newCtx,
		Logger:  logger,
		endpoints: map[string]endpoint.Endpoint{
			// TODO(waigani) this could be refactored further, a lot of dups
			"session":             buildEndpoint(tracer, "session", instances, sessionFactory, randomSeed, logger),
			"query":               buildEndpoint(tracer, "query", instances, queryFactory, randomSeed, logger),
			"review":              buildEndpoint(tracer, "review", instances, reviewFactory, randomSeed, logger),
			"listfacts":           buildEndpoint(tracer, "listfacts", instances, listFactsFactory, randomSeed, logger),
			"listlexicons":        buildEndpoint(tracer, "listlexicons", instances, listLexiconsFactory, randomSeed, logger),
			"pathsfromoffset":     buildEndpoint(tracer, "pathsfromoffset", instances, pathsFromOffsetFactory, randomSeed, logger),
			"describefact":        buildEndpoint(tracer, "describefact", instances, describeFactFactory, randomSeed, logger),
			"latestclientversion": buildEndpoint(tracer, "latestclientversion", instances, latestClientVersionFactory, randomSeed, logger),
		},
	}, nil
}

type grpcOptions struct {
	tls bool
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

	// TODO Refactor to have command specific endpoints, such that retry etc
	// can be adjusted on a per command basis
	// wrap result from loadbalancer.Retry() in function that handles nice printing
	// rather than using retry() which is largely copied
	endpoint, _ := random.Endpoint()
	endpoint = retry(1, 10*time.Second, endpoint)

	return endpoint
	// return kitot.TraceClient(tracer, operationName)(endpoint)
}

func retry(max int, timeout time.Duration, endpoint endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx endpointCtx.Context, request interface{}) (response interface{}, err error) {
		var (
			newctx, cancel = context.WithTimeout(ctx, timeout)
			responses      = make(chan interface{}, 1)
			errs           = make(chan error, 1)
			a              = []string{}
		)
		defer cancel()
		for i := 1; i <= max; i++ {
			go func() {
				response, err := endpoint(newctx, request)
				if err != nil {
					// TODO (junyu) this should use error type instead of string comparison
					if strings.Contains(err.Error(), "grpc: the connection is unavailable") {
						return
					} else {
						errs <- err
						return
					}
				}
				responses <- response
			}()

			select {
			case <-newctx.Done():
				return nil, newctx.Err()
			case response := <-responses:
				return response, nil
			case err := <-errs:
				a = append(a, err.Error())
				continue
			}
		}
		return nil, errors.New(strings.Join(a, "\n"))
	}
}

func userFacingErrs(err error) error {
	// TODO type matching rather than string matching
	// make err struct that can be reformed
	message := err.Error()
	switch {
	case strings.Contains(message, "error: There is no language called:"):
		lang := strings.Split(message, ":")[4]
		lang = lang[1:]
		return errors.Errorf("error: Lingo doesn't support \"%s\" yet", lang)
	// TODO this should be more specific parse error on platform:
	//Error in S25: $(1,), Pos(offset=38, line=7, column=2), expected one of: < ! var indent id
	case strings.Contains(message, "error: expected one of: < ! var indent id"):
		return errors.New("error: Queries must not be terminated by colons.")
	case strings.Contains(message, "error: missing yield"):
		return errors.New("error: You must yield a result, put '<' before any fact or property.")
	default:
		return errors.Trace(err)
	}
}

// TODO Add as a file to ~/.codelingo/config during lingo setup and read in.
// TODO(BlakeMScurr) Update certificate automatically as LetsEncrypt
// renews it.
const cert = `-----BEGIN CERTIFICATE-----
MIIE/DCCA+SgAwIBAgISAzr+eXeUSkyibqSmL0NLlo8RMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xNzAxMDMwMDU2MDBaFw0x
NzA0MDMwMDU2MDBaMBcxFTATBgNVBAMTDGNvZGVsaW5nby5pbzCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBANAUD38GI1qMcvWxtTSknY6gaIt30ssK/iFu
dVmHKaBPlecv6YLmRJC4TjNjIk2VLpeerD0bWPZNSzx3CLs8nCLLfsNGIdLIKvTz
C0YYWr0W/aPubh3k3S3X7CwbDg5/kNzkmuG2DU/KGfxStPzC1JHx+ODaIkDlyZar
xhRiWhvgDJp7+h/Sd71RU4RlBsxOUsuRhnAzlvOoXcnhenn3ffB6Vms1mH7UfbTf
0QYQpi5ErOzPk8ZWo2/fGIfnWyd1H94YaRRDGhNJWfOxtjumuJhe467/NCp6LvF7
G0R4tj4e42Z0fHUv3YSebxYiPmg+iGMhLAVO0WWGZeF9V9u9hCMCAwEAAaOCAg0w
ggIJMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH
AwIwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUCYsrFgKOukNWERko70kGOdYKbWUw
HwYDVR0jBBgwFoAUqEpqYwR93brm0Tm3pkVl7/Oo7KEwcAYIKwYBBQUHAQEEZDBi
MC8GCCsGAQUFBzABhiNodHRwOi8vb2NzcC5pbnQteDMubGV0c2VuY3J5cHQub3Jn
LzAvBggrBgEFBQcwAoYjaHR0cDovL2NlcnQuaW50LXgzLmxldHNlbmNyeXB0Lm9y
Zy8wFwYDVR0RBBAwDoIMY29kZWxpbmdvLmlvMIH+BgNVHSAEgfYwgfMwCAYGZ4EM
AQIBMIHmBgsrBgEEAYLfEwEBATCB1jAmBggrBgEFBQcCARYaaHR0cDovL2Nwcy5s
ZXRzZW5jcnlwdC5vcmcwgasGCCsGAQUFBwICMIGeDIGbVGhpcyBDZXJ0aWZpY2F0
ZSBtYXkgb25seSBiZSByZWxpZWQgdXBvbiBieSBSZWx5aW5nIFBhcnRpZXMgYW5k
IG9ubHkgaW4gYWNjb3JkYW5jZSB3aXRoIHRoZSBDZXJ0aWZpY2F0ZSBQb2xpY3kg
Zm91bmQgYXQgaHR0cHM6Ly9sZXRzZW5jcnlwdC5vcmcvcmVwb3NpdG9yeS8wDQYJ
KoZIhvcNAQELBQADggEBAIerNcWkOm7aN+EMYZnWylCS4JZQItfZvrmRyuxkwoKc
ixbEidK8hrvDf7edJKgsb9nsdOUcmlSUp82lT5iIBtvN1melMBngLjqH58SY1nJW
6Sa/GEmvaKoC2RLqLOlvY9QYcaA3bkqpNiSr+Xz+AuP8KKlvZ4iV9VKuSxKESeNt
PAZQV9HY0FpOvDfDcIyhC/io7n0PQu661u/0uc/gpFGcX4AtOKN7F4fSVL6NPDhi
g8UL5PLPnv2Umidknk4xpsHt34NRv7kxhQShD7L9lm4LMDzZSW62ySXeAqWTQpBw
zp8/kZht7EA7u1ayv6KE4kfYdWz71gjmnOjLwzVb1oU=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEkjCCA3qgAwIBAgIQCgFBQgAAAVOFc2oLheynCDANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTE2MDMxNzE2NDA0NloXDTIxMDMxNzE2NDA0Nlow
SjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxIzAhBgNVBAMT
GkxldCdzIEVuY3J5cHQgQXV0aG9yaXR5IFgzMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAnNMM8FrlLke3cl03g7NoYzDq1zUmGSXhvb418XCSL7e4S0EF
q6meNQhY7LEqxGiHC6PjdeTm86dicbp5gWAf15Gan/PQeGdxyGkOlZHP/uaZ6WA8
SMx+yk13EiSdRxta67nsHjcAHJyse6cF6s5K671B5TaYucv9bTyWaN8jKkKQDIZ0
Z8h/pZq4UmEUEz9l6YKHy9v6Dlb2honzhT+Xhq+w3Brvaw2VFn3EK6BlspkENnWA
a6xK8xuQSXgvopZPKiAlKQTGdMDQMc2PMTiVFrqoM7hD8bEfwzB/onkxEz0tNvjj
/PIzark5McWvxI0NHWQWM6r6hCm21AvA2H3DkwIDAQABo4IBfTCCAXkwEgYDVR0T
AQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwfwYIKwYBBQUHAQEEczBxMDIG
CCsGAQUFBzABhiZodHRwOi8vaXNyZy50cnVzdGlkLm9jc3AuaWRlbnRydXN0LmNv
bTA7BggrBgEFBQcwAoYvaHR0cDovL2FwcHMuaWRlbnRydXN0LmNvbS9yb290cy9k
c3Ryb290Y2F4My5wN2MwHwYDVR0jBBgwFoAUxKexpHsscfrb4UuQdf/EFWCFiRAw
VAYDVR0gBE0wSzAIBgZngQwBAgEwPwYLKwYBBAGC3xMBAQEwMDAuBggrBgEFBQcC
ARYiaHR0cDovL2Nwcy5yb290LXgxLmxldHNlbmNyeXB0Lm9yZzA8BgNVHR8ENTAz
MDGgL6AthitodHRwOi8vY3JsLmlkZW50cnVzdC5jb20vRFNUUk9PVENBWDNDUkwu
Y3JsMB0GA1UdDgQWBBSoSmpjBH3duubRObemRWXv86jsoTANBgkqhkiG9w0BAQsF
AAOCAQEA3TPXEfNjWDjdGBX7CVW+dla5cEilaUcne8IkCJLxWh9KEik3JHRRHGJo
uM2VcGfl96S8TihRzZvoroed6ti6WqEBmtzw3Wodatg+VyOeph4EYpr/1wXKtx8/
wApIvJSwtmVi4MFU5aMqrSDE6ea73Mj2tcMyo5jMd6jmeWUHK8so/joWUoHOUgwu
X4Po1QYz+3dszkDqMp4fklxBwXRsW10KXzPMTZ+sOPAveyxindmjkW8lGy+QsRlG
PfZ+G6Z6h7mjem0Y+iWlkYcV4PIWL1iwBi8saCbGS5jN2p8M+X+Q7UNKEkROb3N6
KOqkqm57TH2H3eDJAkSnh6/DNFu0Qg==
-----END CERTIFICATE-----
`
