package service

import (
	"context"
	"crypto/tls"
	"math"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common/config"
	rpc "github.com/codelingo/rpc/service"
	"github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	FlowServer = "flowserver"
	// TODO: remove public access to the platform server
	PlatformServer = "platformserver"
	FlowClient     = "flowclient"
	LocalClient    = "localclient"

	// Can be increased to 4 << 30 (4GB) on 64 bit systems.
	MaxGrpcMessageSize = math.MaxInt32
)

// GrpcConnection creates a connection between a given server and client type.
// TODO(BlakeMScurr): this should be moved into its own service repo, so that flow and platform don't have
// to depend on the client. The code pertaining specifically to the client side and flow side
// configs should be kept in the client/flow repos, and addresses and tls values should be
// passed as arguments.
func GrpcConnection(client, server string, insecureAllowed bool) (*grpc.ClientConn, error) {
	var grpcAddr string
	isTLS := !insecureAllowed

	switch client {
	case LocalClient:
		pCfg, err := config.Platform()
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

	conn, err := dial(grpcAddr, isTLS)
	if err != nil {
		return nil, errors.Trace(err)
	}

	util.Logger.Debug("...got answer from grpc server.")

	return conn, nil
}

func dial(target string, isTLS bool) (*grpc.ClientConn, error) {

	var tlsOpt grpc.DialOption
	if !isTLS {
		tlsOpt = grpc.WithInsecure()
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		tlsOpt = grpc.WithTransportCredentials(creds)
	}

	util.Logger.Debug("dialing grpc server...")
	return grpc.Dial(target, tlsOpt, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MaxGrpcMessageSize),
		grpc.MaxCallSendMsgSize(MaxGrpcMessageSize),
	))
}

func ListLexicons(ctx context.Context) ([]string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.ListLexiconsRequest{}
	reply, err := c.ListLexicons(ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return reply.Lexicons, nil
}

func ListFacts(ctx context.Context, owner, name, version string) (map[string][]string, error) {
	req := &rpc.ListFactsRequest{
		Owner:   owner,
		Name:    name,
		Version: version,
	}

	reply, err := listFacts(ctx, req, 1)
	if err != nil {
		return nil, errors.Trace(err)
	}

	factMap := make(map[string][]string)
	facts := reply.Facts
	for parent, children := range facts {
		factMap[parent] = children.Child
	}

	return factMap, errors.Trace(err)
}

func listFacts(ctx context.Context, req *rpc.ListFactsRequest, attempt int) (*rpc.FactList, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	reply, err := c.ListFacts(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), certErrorString) {
			path, err := platformCertPath()
			if err != nil {
				return nil, errors.Trace(err)
			}

			err = os.Remove(path)
			if err != nil {
				return nil, errors.Trace(err)
			}

			// retry 5 times
			// TODO: exponential backoff and annotate with connection error
			if attempt >= maxAttempts {
				return nil, errors.Errorf("attempted to connect %d times", attempt)
			}

			reply, err = listFacts(ctx, req, attempt+1)
			return reply, errors.Trace(err)
		} else {
			return nil, errors.Trace(err)
		}
	}

	return reply, nil
}

func DescribeFact(ctx context.Context, owner, name, version, fact string) (*rpc.DescribeFactReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.DescribeFactRequest{
		Owner:   owner,
		Name:    name,
		Version: version,
		Fact:    fact,
	}
	reply, err := c.DescribeFact(ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return reply, nil
}

const certErrorString string = "transport: authentication handshake failed: x509: certificate signed by unknown authority"

func platformCertPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Trace(err)
	}
	return path.Join(usr.HomeDir, ".codelingo/certs/paas/grpc-platform.codelingo.io:443.cert"), nil
}

const maxAttempts int = 5

func QueryFromOffset(ctx context.Context, req *rpc.QueryFromOffsetRequest, insecureAllowed bool) (*rpc.QueryFromOffsetReply, error) {
	reply, err := queryFromOffset(ctx, req, 1, insecureAllowed)
	return reply, errors.Trace(err)
}
func queryFromOffset(ctx context.Context, req *rpc.QueryFromOffsetRequest, attempt int, insecureAllowed bool) (*rpc.QueryFromOffsetReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, insecureAllowed)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	reply, err := c.QueryFromOffset(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), certErrorString) {
			path, err := platformCertPath()
			if err != nil {
				return nil, errors.Trace(err)
			}

			err = os.Remove(path)
			if err != nil {
				return nil, errors.Trace(err)
			}

			// retry 5 times
			// TODO: exponential backoff and annotate with connection error
			if attempt >= maxAttempts {
				return nil, errors.Errorf("attempted to connect %d times", attempt)
			}

			reply, err = queryFromOffset(ctx, req, attempt+1, insecureAllowed)
			return reply, errors.Trace(err)
		} else {
			return nil, errors.Trace(err)
		}
	}

	return reply, nil
}

func LatestClientVersion(ctx context.Context) (string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
	if err != nil {
		return "", errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.LatestClientVersionRequest{}
	reply, err := c.LatestClientVersion(ctx, req)
	if err != nil {
		return "", errors.Trace(err)
	}

	return reply.Version, nil
}
