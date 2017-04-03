package config

import (
	"fmt"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"github.com/codelingo/lingo/app/util"
	"path/filepath"
	"io/ioutil"
	"os"
)

const (
	gitRemoteName = "gitserver.remote.name"
	gitServerHost = "gitserver.remote.host"
	gitServerPort = "gitserver.remote.port"
	gitServerProtocol = "gitserver.remote.protocol"
	platformServerAddr = "addr"
	platformServerPort = "port"
	platformServerGrpcPort = "grpc_port"
	mqAddrProtocol = "messagequeue.address.protocol"
	mqAddrUsername = "messagequeue.address.username"
	mqAddrPassword = "messagequeue.address.password"
	mqAddrHost = "messagequeue.address.host"
	mqAddrPort = "messagequeue.address.port"
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

func createPlatformFile(basepath string, overwrite bool) error {
	pCfgFilePath := filepath.Join(basepath, PlatformCfgFile)
	if _, err := os.Stat(pCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(pCfgFilePath, []byte(PlatformTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create platform config")
		}
	}
	return nil
}

func CreatePlatformFile(overwrite bool) error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return createPlatformFile(configHome, overwrite)
}

func CreatePlatformDefaultFile() error {
	configDefaults, err := util.ConfigDefaults()
	if err != nil {
		return errors.Trace(err)
	}
	return createPlatformFile(configDefaults, true)
}


func (p *platformConfig) GitRemoteName() (string, error) {
	return p.Get(gitRemoteName)
}

func (p *platformConfig) GitServerHost() (string, error) {
	return p.Get(gitServerHost)
}

func (p *platformConfig) GitServerPort() (string, error) {
	return p.Get(gitServerPort)
}

func (p *platformConfig) GitServerProtocol() (string, error) {
	return p.Get(gitServerProtocol)
}

func (p *platformConfig) GitServerAddr() (string, error) {

	protocol, err := p.GitServerProtocol()
	if err != nil {
		return "", errors.Trace(err)
	}

	host, err := p.GitServerHost()
	if err != nil {
		return "", errors.Trace(err)
	}

	addr := protocol + "://" + host
	port, err := p.GitServerPort()
	if err != nil || port == "" {
		return addr, nil
	}
	return addr + ":" + port, nil
}

func (p *platformConfig) Address() (string, error) {
	addr, err := p.Get(platformServerAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get(platformServerPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) GrpcAddress() (string, error) {

	addr, err := p.Get(platformServerAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get(platformServerGrpcPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) MessageQueueAddr() (string, error) {
	protocol, err := p.Get(mqAddrProtocol)
	if err != nil {
		return "", errors.Trace(err)
	}

	username, err := p.Get(mqAddrUsername)
	if err != nil {
		return "", errors.Trace(err)
	}

	password, err := p.Get(mqAddrPassword)
	if err != nil {
		return "", errors.Trace(err)
	}

	host, err := p.Get(mqAddrHost)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.Get(mqAddrPort)
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
  port: "30300"
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
