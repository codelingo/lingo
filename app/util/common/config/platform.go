package config

import (
	"fmt"

	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"github.com/codelingo/lingo/app/util"
	"path/filepath"
)

type platformConfig struct {
	*config.FileConfig
}

func Platform() (*platformConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	pCfgPath, err := fullCfgPath(PlatformCfgFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	pCfg, err := cfg.New(pCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &platformConfig{
		pCfg,
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

func (p *platformConfig) MessageQueueAddr() (string, error) {
	protocol, err := p.Get("messagequeue.address.protocol")
	if err != nil {
		return "", errors.Trace(err)
	}

	username, err := p.Get("messagequeue.address.username")
	if err != nil {
		return "", errors.Trace(err)
	}

	password, err := p.Get("messagequeue.address.password")
	if err != nil {
		return "", errors.Trace(err)
	}

	host, err := p.Get("messagequeue.address.host")
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get("messagequeue.address.port")
	if err != nil {
		return "", errors.Trace(err)
	}

	return fmt.Sprintf("%s://%s:%s@%s:%s/", protocol, username, password, host, port), nil
}

var PlatformTmpl = `
all:
  addr: codelingo.io
  port: "80"
  grpc_port: "8002"
  gitserver:
    tls: "false"
    remote:
      name: "codelingo"
      protocol: "https"
      host: "codelingo.io"
      port: "3030"
  messagequeue:
    address:
      protocol: "amqp"
      username: "codelingo"
      password: "codelingo"
      host: "codelingo.io"
      port: "5672"
dev:
  addr: localhost
  port: "3030"
  gitserver:
    tls: "true"
    remote:
      name: "codelingo_dev"
      protocol: "http"
      host: "localhost"
      port: "3000"
  messagequeue:
    address:
      protocol: "amqp"
      username: "codelingo"
      password: "codelingo"
      host: "localhost"
      port: "5672"
onprem:
  addr: 192.168.99.100
  // port: "30300"
  grpc_port: "30082"
  gitserver:
    tls: "true"
    remote:
      name: "codelingo_onprem"
      protocol: "http"
      host: "192.168.99.100"
      port: "30300"
  messagequeue:
    address:
      protocol: "amqp"
      username: "codelingo"
      password: "codelingo"
      host: "192.168.99.100"
      port: "30567"
test:
  addr: localhost
  port: "3030"
  gitserver:
    tls: "false"
    remote:
      name: "codelingo_dev"
      protocol: "http"
      host: "localhost"
      port: "3000"
  messagequeue:
    address:
      protocol: "amqp"
      username: "codelingo"
      password: "codelingo"
      host: "localhost"
      port: "5672"
`[1:]
