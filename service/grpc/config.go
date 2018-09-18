package grpc

import (
	"strings"

	"github.com/codelingo/lingo/app/util"
	commonConfig "github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func AddUsernameToCtx(ctx context.Context) (context.Context, error) {
	authCfg, err := commonConfig.Auth()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// TODO: have a single CodeLingo username instead of using repo usernames
	for i := 0; i < 3; i++ {
		var username string
		var err error

		switch i {
		case 0:
			username, err = authCfg.GetGitUserName()
		case 1:
			username, err = authCfg.GetP4UserName()
		default:
			util.Logger.Warn("Using `demo` account - please run `lingo config setup` to access your private repos.")
			username, err = "demo", nil
		}

		if err != nil {
			if strings.Contains(err.Error(), "Could not find value") {
				continue
			} else {
				return nil, errors.Trace(err)
			}
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(make(map[string]string))
		}
		md = md.Copy()
		md["username"] = append(md["username"], username)

		return metadata.NewOutgoingContext(ctx, md), nil
	}

	return nil, errors.New("failed to add username to context")
}
