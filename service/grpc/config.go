package grpc

import (
	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	serviceConfig "github.com/codelingo/lingo/service/config"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"path/filepath"
)

func GetGcloudEndpointCtx() (context.Context, error) {
	configsHome, err := util.ConfigHome()
	if err != nil {
		return nil, errors.Trace(err)
	}

	envFilepath := filepath.Join(configsHome, commonConfig.EnvCfgFile)
	cfg := serviceConfig.New(envFilepath)

	env, err := cfg.GetEnv()
	if err != nil {
		return nil, errors.Trace(err)
	}

	ctx := context.Background()

	if env == "all" || env == "staging" {
		cfg, err := commonConfig.Platform()
		if err != nil {
			return nil, errors.Trace(err)
		}
		gcloudAPIKey, err := cfg.GetValue("gcloud.API_key")
		if err != nil {
			return nil, errors.Trace(err)
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-api-key", gcloudAPIKey))
		return ctx, nil
	}

	return ctx, nil
}
