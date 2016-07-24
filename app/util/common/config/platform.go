package config

type platformConfig struct {
	Addr     string `yaml:"addr"`
	GrpcPort string `yaml:"grpc_port"`
}

func Platform() (*platformConfig, error) {
	cfg := &platformConfig{}
	if err := Load(PlatformCfgFile, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

var PlatformTmpl = `
addr: codelingo.io
grpc_port: 8002
`[1:]
