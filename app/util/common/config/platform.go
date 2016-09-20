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

func (p *platformConfig) GitServerAddr() (string, error) {

	addr, err := p.Get("gitserver.remote.addr")
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get("gitserver.remote.port")
	if err != nil || port == "" {
		return addr, nil
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
  grpc_port: "8002"
  gitserver:
    remote:
      name: "codelingo"
      addr: "http://codelingo.io"
      port: "3030"
dev:
  addr: localhost
  gitserver:
    remote:
      name: "codelingo_dev"
      addr: "http://localhost"
      port: "3000"
test:
  addr: localhost
  gitserver:
    remote:
      name: "codelingo_dev"
      addr: "http://localhost"
      port: "3000"
`[1:]
