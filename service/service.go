package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Review(ctx context.Context, req *codelingo.ReviewRequest) (chan *codelingo.Issue, error) {
	cc, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}

	client := codelingo.NewCodeLingoClient(cc)
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

func Query(ctx context.Context, cc *grpc.ClientConn, queryc chan *codelingo.QueryRequest) (chan *codelingo.QueryReply, error) {
	client := codelingo.NewCodeLingoClient(cc)
	resultc, err := runQuery(ctx, client, queryc)
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

func runQuery(ctx context.Context, client codelingo.CodeLingoClient, queryc chan *codelingo.QueryRequest) (chan *codelingo.QueryReply, error) {
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

func ListLexicons(ctx context.Context) ([]string, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := codelingo.NewCodeLingoClient(conn)

	req := &codelingo.ListLexiconsRequest{}
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
	c := codelingo.NewCodeLingoClient(conn)

	req := &codelingo.ListFactsRequest{
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

func DescribeFact(ctx context.Context, owner, name, version, fact string) (*codelingo.DescribeFactReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := codelingo.NewCodeLingoClient(conn)

	req := &codelingo.DescribeFactRequest{
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

func PathsFromOffset(ctx context.Context, req *codelingo.PathsFromOffsetRequest) (*codelingo.PathsFromOffsetReply, error) {
	conn, err := GrpcConnection(LocalClient, PlatformServer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := codelingo.NewCodeLingoClient(conn)

	reply, err := c.PathsFromOffset(ctx, req)
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
	c := codelingo.NewCodeLingoClient(conn)

	req := &codelingo.LatestClientVersionRequest{}
	reply, err := c.LatestClientVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return reply.Version, nil
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
