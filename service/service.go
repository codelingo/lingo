package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"

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

	MaxGrpcMessageSize = 4 << 30
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
		isTLS = true
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

	var tlsOpt grpc.DialOption
	if !isTLS {
		tlsOpt = grpc.WithInsecure()
	} else {
		// TODO: cache credentials locally to reduce connection overhead
		// TODO: host may be insecure and will fail here; prompt for insecure or require flag
		creds, err := credsFromHost(grpcAddr)
		if err != nil {
			return nil, errors.Trace(err)
		}
		tlsOpt = grpc.WithTransportCredentials(creds)
	}

	conn, err := grpc.Dial(grpcAddr, tlsOpt, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MaxGrpcMessageSize),
		grpc.MaxCallSendMsgSize(MaxGrpcMessageSize),
	))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return conn, nil
}

// credsFromHost retrieves the public certificate from the given host and returns the transport credentials.
func credsFromHost(host string) (credentials.TransportCredentials, error) {
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer conn.Close()
	err = conn.Handshake()
	if err != nil {
		return nil, errors.Trace(err)
	}
	cert := conn.ConnectionState().PeerCertificates[0]
	cp := x509.NewCertPool()
	cp.AddCert(cert)
	return credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp}), nil
}

func ListLexicons(ctx context.Context) ([]string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.ListLexiconsRequest{}
	reply, err := c.ListLexicons(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.Lexicons, nil
}

func ListFacts(ctx context.Context, owner, name, version string) (map[string][]string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.ListFactsRequest{
		Owner:   owner,
		Name:    name,
		Version: version,
	}
	reply, err := c.ListFacts(ctx, req)
	if err != nil {
		return nil, err
	}

	factMap := make(map[string][]string)
	facts := reply.Facts
	for parent, children := range facts {
		factMap[parent] = children.Child
	}

	return factMap, nil

}

func DescribeFact(ctx context.Context, owner, name, version, fact string) (*rpc.DescribeFactReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
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
		return nil, err
	}

	return reply, nil
}

func QueryFromOffset(ctx context.Context, req *rpc.QueryFromOffsetRequest) (*rpc.QueryFromOffsetReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	reply, err := c.QueryFromOffset(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func LatestClientVersion(ctx context.Context) (string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return "", errors.Trace(err)
	}
	c := rpc.NewCodeLingoClient(conn)

	req := &rpc.LatestClientVersionRequest{}
	reply, err := c.LatestClientVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return reply.Version, nil
}
