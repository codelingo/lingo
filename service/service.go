package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"

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
	var err error
	var isTLS bool
	var cert *x509.Certificate

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

	if isTLS {
		// TODO: host may be insecure and will fail here; prompt for insecure or require flag

		util.Logger.Debug("getting tls cert from cache...")
		cert, err = getCertFromCache(grpcAddr)
		if err != nil {
			// TODO(waigani) check error
			// return nil, errors.Trace(err)

			// if cert hasn't been cached, get a new one which caches it under the hood
			util.Logger.Debug("no cert found, creating new one...")
			if cert, err = newCert(grpcAddr); err != nil && !insecureAllowed {
				return nil, errors.Trace(err)
			}
		}
	}

	conn, err := dial(grpcAddr, cert, insecureAllowed)
	if err != nil {
		if cert == nil {
			return nil, errors.Trace(err)
		}

		// TODO(waigani) check error

		// if cert is stale, get a new one
		util.Logger.Debug("dial up failed with given cert, creating new cert...")
		if cert, err = newCert(grpcAddr); err != nil {
			return nil, errors.Trace(err)
		}

		if conn, err = dial(grpcAddr, cert, insecureAllowed); err != nil {
			return nil, errors.Trace(err)
		}

	}
	util.Logger.Debug("...got answer from grpc server.")

	return conn, nil
}

func dial(target string, cert *x509.Certificate, insecureAllowed bool) (*grpc.ClientConn, error) {
	tlsOpt := grpc.WithInsecure()
	if cert != nil {
		creds, err := credsFromCert(cert)
		if err != nil {
			if insecureAllowed {
				util.Logger.Warn("failed secure, trying insecure")
				tlsOpt = grpc.WithInsecure()
			} else {
				return nil, errors.Trace(err)
			}
		} else {
			tlsOpt = grpc.WithTransportCredentials(creds)
		}
	}

	util.Logger.Debug("dialing grpc server...")
	return grpc.Dial(target, tlsOpt, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MaxGrpcMessageSize),
		grpc.MaxCallSendMsgSize(MaxGrpcMessageSize),
	))
}

func newCert(host string) (*x509.Certificate, error) {
	cert, err := certFromHost(host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := cacheRawCert(host, cert.Raw); err != nil {
		return nil, errors.Trace(err)
	}

	return cert, nil
}

func credsFromCert(cert *x509.Certificate) (credentials.TransportCredentials, error) {
	cp := x509.NewCertPool()
	cp.AddCert(cert)
	return credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp}), nil
}

func getCertFromCache(host string) (*x509.Certificate, error) {

	certP, err := certPath(host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rawCert, err := ioutil.ReadFile(certP)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return x509.ParseCertificate(rawCert)

}

func certPath(host string) (string, error) {
	homePath, err := util.LingoHome()
	if err != nil {
		return "", errors.Trace(err)
	}

	env, err := util.GetEnv()
	if err != nil {
		return "", errors.Trace(err)
	}

	return path.Join(homePath, fmt.Sprintf("certs/%s/%s.cert", env, host)), nil

}

func cacheRawCert(host string, rawCert []byte) error {
	certP, err := certPath(host)
	if err != nil {
		return errors.Trace(err)
	}

	if err := os.MkdirAll(filepath.Dir(certP), 0755); err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(ioutil.WriteFile(certP, rawCert, 0755))
}

// credsFromHost retrieves the public certificate from the given host and returns the transport credentials.
func certFromHost(host string) (*x509.Certificate, error) {
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer conn.Close()
	err = conn.Handshake()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return conn.ConnectionState().PeerCertificates[0], nil
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
		return nil, err
	}
	return reply.Lexicons, nil
}

func ListFacts(ctx context.Context, owner, name, version string) (map[string][]string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
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
		return nil, err
	}

	return reply, nil
}

func QueryFromOffset(ctx context.Context, req *rpc.QueryFromOffsetRequest) (*rpc.QueryFromOffsetReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
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
	conn, err := GrpcConnection(LocalClient, PlatformServer, false)
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
