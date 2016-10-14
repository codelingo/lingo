package config

import (
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
)

type platformConfig struct {
	*config.Config
}

func Platform() (*platformConfig, error) {
	cfgPath, err := fullCfgPath(PlatformCfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg, err := config.New(cfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &platformConfig{
		Config: cfg,
	}, nil
}

func (p *platformConfig) GitRemoteName() (string, error) {
	return p.Get("gitserver.remote.name")
}

func (p *platformConfig) GitServerHost() (string, error) {
	return p.Get("gitserver.remote.host")
}

func (p *platformConfig) GitServerPort() (string, error) {
	return p.Get("gitserver.remote.port")
}

func (p *platformConfig) GitServerProtocol() (string, error) {
	return p.Get("gitserver.remote.protocol")
}

func (p *platformConfig) GitServerAddr() (string, error) {

	protocol, err := p.Get("gitserver.remote.protocol")
	if err != nil {
		return "", errors.Trace(err)
	}

	host, err := p.Get("gitserver.remote.host")
	if err != nil {
		return "", errors.Trace(err)
	}

	addr := protocol + "://" + host
	port, err := p.Get("gitserver.remote.port")
	if err != nil || port == "" {
		return addr, nil
	}
	return addr + ":" + port, nil
}

func (p *platformConfig) Address() (string, error) {
	addr, err := p.Get("addr")
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get("port")
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) GrpcAddress() (string, error) {

	addr, err := p.Get("addr")
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get("grpc_port")
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

var PlatformTmpl = `
all:
  addr: codelingo.io
  port: "80"
  grpc_port: "8002"
  gitserver:
    remote:
      name: "codelingo"
      protocol: "http"
      host: "codelingo.io"
      port: "3030"
dev:
  addr: localhost
  port: "3030"
  gitserver:
    remote:
      name: "codelingo_dev"
      protocol: "http"
      host: "localhost"
      port: "3000"
test:
  addr: localhost
  port: "3030"
  gitserver:
    remote:
      name: "codelingo_dev"
      protocol: "http"
      host: "localhost"
      port: "3000"
`[1:]
