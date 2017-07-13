package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
)

const (
	gitRemoteName          = "gitserver.remote.name"
	gitServerHost          = "gitserver.remote.host"
	gitServerPort          = "gitserver.remote.port"
	gitServerProtocol      = "gitserver.remote.protocol"
	webSiteAddr            = "website.addr"
	webSitePort            = "website.port"
	platformServerAddr     = "platform.addr"
	platformServerPort     = "platform.port"
	platformServerGrpcPort = "grpc_port"
	flowServerGrpcPort     = "flow.port"
	flowServerGrpcAddress  = "flow.address"
	mqAddrProtocol         = "messagequeue.address.protocol"
	mqAddrUsername         = "messagequeue.address.username"
	mqAddrPassword         = "messagequeue.address.password"
	mqAddrHost             = "messagequeue.address.host"
	mqAddrPort             = "messagequeue.address.port"
)

type platformConfig struct {
	*config.FileConfig
}

func PlatformInDir(dir string) (*platformConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	envFile := filepath.Join(configHome, EnvCfgFile)
	cfg := config.New(envFile)

	pCfgPath := filepath.Join(dir, PlatformCfgFile)
	pCfg, err := cfg.New(pCfgPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &platformConfig{
		pCfg,
	}, nil
}

func Platform() (*platformConfig, error) {
	configHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return PlatformInDir(configHome)
}

func CreatePlatformFileInDir(dir string, overwrite bool) error {
	pCfgFilePath := filepath.Join(dir, PlatformCfgFile)
	if _, err := os.Stat(pCfgFilePath); os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(pCfgFilePath, []byte(PlatformTmpl), 0644)
		if err != nil {
			return errors.Annotate(err, "verifyConfig: Could not create platform config")
		}
	}
	return nil
}

func CreatePlatformFile() error {
	configHome, err := util.ConfigHome()
	if err != nil {
		return errors.Trace(err)
	}
	return CreatePlatformFileInDir(configHome, false)
}

func (p *platformConfig) Dump() (map[string]interface{}, error) {
	keyMap := make(map[string]interface{})

	var platDumpConsts = []string{
		gitRemoteName,
		gitServerHost,
		gitServerPort,
		gitServerProtocol,
		webSiteAddr,
		webSitePort,
		platformServerPort,
		platformServerAddr,
		platformServerPort,
		platformServerGrpcPort,
		mqAddrProtocol,
		mqAddrUsername,
		mqAddrPassword,
		mqAddrHost,
		mqAddrPort,
	}

	for _, pCon := range platDumpConsts {
		newMap, err := p.GetAll(pCon)
		if err != nil {
			return nil, errors.Trace(err)
		}
		for k, v := range newMap {
			keyMap[k] = v
		}
	}

	return keyMap, nil
}

func (p *platformConfig) GitRemoteName() (string, error) {
	return p.GetValue(gitRemoteName)
}

func (p *platformConfig) GitServerHost() (string, error) {
	return p.GetValue(gitServerHost)
}

func (p *platformConfig) GitServerPort() (string, error) {
	return p.GetValue(gitServerPort)
}

func (p *platformConfig) GitServerProtocol() (string, error) {
	return p.GetValue(gitServerProtocol)
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
	addr, err := p.GetValue(platformServerAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.GetValue(platformServerPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) WebSiteAddress() (string, error) {
	addr, err := p.GetValue(webSiteAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.GetValue(webSitePort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) PlatformAddress() (string, error) {
	addr, err := p.GetValue(platformServerAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.GetValue(platformServerGrpcPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) FlowAddress() (string, error) {

	addr, err := p.GetValue(flowServerGrpcAddress)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.GetValue(flowServerGrpcPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return addr + ":" + port, nil
}

func (p *platformConfig) MessageQueueAddr() (string, error) {
	protocol, err := p.GetValue(mqAddrProtocol)
	if err != nil {
		return "", errors.Trace(err)
	}

	username, err := p.GetValue(mqAddrUsername)
	if err != nil {
		return "", errors.Trace(err)
	}

	password, err := p.GetValue(mqAddrPassword)
	if err != nil {
		return "", errors.Trace(err)
	}

	host, err := p.GetValue(mqAddrHost)
	if err != nil {
		return "", errors.Trace(err)
	}

	port, err := p.GetValue(mqAddrPort)
	if err != nil {
		return "", errors.Trace(err)
	}

	return fmt.Sprintf("%s://%s:%s@%s:%s/", protocol, username, password, host, port), nil
}

var PlatformTmpl = `
all:
  website:
    addr: codelingo.io
    port: "80"
  platform:
    addr: codelingo.io
    port: "80"
  grpc_port: "8002"
  gitserver:
    tls: "true"
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
  website:
    addr: 10.0.17.233
    port: "30303"
  platform:
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
onprem:
  website:
    addr: 10.0.17.233
    port: "30303"
  platform:
    addr: 10.0.17.233
    port: "30303"
  grpc_port: "30082"
  gitserver:
    tls: "false"
    remote:
      name: "codelingo_onprem"
      protocol: "http"
      host: "10.0.17.233"
      port: "30300"
  messagequeue:
    address:
      protocol: "amqp"
      username: "codelingo"
      password: "codelingo"
      host: "10.0.17.233"
      port: "30567"
test:
  website:
    addr: 10.0.17.233
    port: "30303"
  platform:
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
`[1:]
