package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
)

const (
	websiteHTTPAddr  = "website"
	platformGRPCAddr = "platform"
	flowGRPCAddr     = "flow"

	gitServerRemote = "gitserver.remote"
	gitServerAddr   = "gitserver.addr"

	p4RemoteName      = "p4server.remote.name"
	p4RemoteDepotName = "p4server.remote.depot.name"
	p4ServerHost      = "p4server.remote.host"
	p4ServerPort      = "p4server.remote.port"
	p4ServerProtocol  = "p4server.remote.protocol"
)

// defaultConfig is the config that is written when an existing config can't be found.
const defaultConfig = `paas:
  website: https://www.codelingo.io
  platform: grpc-platform.codelingo.io:443
  flow: grpc-flow.codelingo.io:443
  gitserver:
    addr: https://git.codelingo.io:443
    remote: codelingo
`

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
		err := ioutil.WriteFile(pCfgFilePath, []byte(defaultConfig), 0644)
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
		websiteHTTPAddr,
		platformGRPCAddr,
		flowGRPCAddr,
		gitServerAddr,
		gitServerRemote,
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
	return p.GetValue(gitServerRemote)
}

func (p *platformConfig) GitServerAddr() (string, error) {
	addr, err := p.GetValue(gitServerAddr)
	if err != nil {
		return "", errors.Trace(err)
	}
	return addr, nil
}

func (p *platformConfig) WebSiteAddress() (string, error) {
	addr, err := p.GetValue(websiteHTTPAddr)
	if err != nil {
		return "", errors.Trace(err)
	}
	return addr, nil
}

func (p *platformConfig) PlatformAddress() (string, error) {
	addr, err := p.GetValue(platformGRPCAddr)
	if err != nil {
		return "", errors.Trace(err)
	}
	return addr, nil
}

func (p *platformConfig) FlowAddress() (string, error) {
	addr, err := p.GetValue(flowGRPCAddr)
	if err != nil {
		return "", errors.Trace(err)
	}
	return addr, nil
}

func (p *platformConfig) P4ServerAddr() (string, error) {
	addr, err := p.GetValue(p4ServerHost)
	if err != nil {
		return "", errors.Trace(err)
	}
	port, err := p.GetValue(p4ServerPort)
	if err != nil {
		return "", errors.Trace(err)
	}
	return addr + ":" + port, nil
}

func (p *platformConfig) P4RemoteName() (string, error) {
	remoteName, err := p.GetValue(p4RemoteName)
	if err != nil {
		return "", errors.Trace(err)
	}
	return remoteName, nil
}

func (p *platformConfig) P4RemoteDepotName() (string, error) {
	remoteName, err := p.GetValue(p4RemoteDepotName)
	if err != nil {
		return "", errors.Trace(err)
	}
	return remoteName, nil
}
