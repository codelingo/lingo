package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	grpcclient "github.com/codelingo/lingo/service/grpc"
	"github.com/codelingo/rpc/flow"
	"github.com/codelingo/rpc/flow/client"
	"github.com/juju/errors"
)

var searchCommand = cli.Command{
	Name:        "search",
	Usage:       "Search code following queries in rpc.yaml.",
	Subcommands: cli.Commands{*pullRequestCmd},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  util.OutputFlg.String(),
			Usage: "File to save found results to.",
		},
		cli.StringFlag{
			Name:  util.FormatFlg.String(),
			Value: "json-pretty",
			Usage: "How to format the found results. Possible values are: json, json-pretty.",
		},
	},
	Description: `
""$ lingo search <filename>" .
`[1:],
	Action: searchAction,
}

func searchAction(ctx *cli.Context) {
	err := searchRequire()
	if err != nil {
		util.FatalOSErr(err)
		return
	}

	msg, err := searchCMD(ctx)
	if err != nil {

		// Debugging
		util.Logger.Debugw("searchAction", "err_stack", errors.ErrorStack(err))

		util.FatalOSErr(err)
		return
	}

	fmt.Println(msg)
}

func searchRequire() error {
	reqs := []require{homeRq, authRq, configRq, versionRq}
	for _, req := range reqs {
		err := req.Verify()
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func searchCMD(cliCtx *cli.Context) (string, error) {
	defer util.Logger.Sync()
	util.Logger.Debugw("searchCMD called")

	args := cliCtx.Args()
	if len(args) == 0 {
		return "", errors.New("Please specify the filepath to a rpc.yaml file.")
	}

	dotlingo, err := ioutil.ReadFile(args[0])
	if err != nil {
		return "", errors.Trace(err)
	}

	conn, err := service.GrpcConnection(service.LocalClient, service.FlowServer)
	if err != nil {
		return "", errors.Trace(err)
	}

	c := client.NewFlowClient(conn)
	ctx, cancel := util.UserCancelContext(context.Background())

	// Create context with metadata
	ctx, err = grpcclient.AddGcloudApiKeyToCtx(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}
	ctx, err = grpcclient.AddUsernameToCtx(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	fmt.Println("Running search flow...")
	resultc, errorc, err := c.Search(ctx, &flow.SearchRequest{
		Dotlingo: string(dotlingo),
	})
	if err != nil {
		return "", errors.Trace(err)
	}

	results := []*flow.SearchReply{}

l:
	for {
		select {
		case err, ok := <-errorc:
			if !ok {
				errorc = nil
				break
			}

			// Abort search
			cancel()
			util.Logger.Debugf("Search error: %s", errors.ErrorStack(err))
			return "", errors.Trace(err)
		case result, ok := <-resultc:
			if !ok {
				resultc = nil
				break
			}

			if result.Error != "" {
				cancel()
				return "", errors.New(result.Error)
			}

			results = append(results, result)
		}
		if resultc == nil && errorc == nil {
			break l
		}
	}

	msg, err := OutputResults(results, cliCtx.String("format"), cliCtx.String("output"))
	return msg, errors.Trace(err)
}

func OutputResults(results []*flow.SearchReply, format, outputFile string) (string, error) {
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = json.Marshal(results)
		if err != nil {
			return "", errors.Trace(err)
		}
	case "json-pretty":
		data, err = json.MarshalIndent(results, " ", " ")
		if err != nil {
			return "", errors.Trace(err)
		}
	default:
		return "", errors.Errorf("Unknown format %q", format)
	}

	if outputFile != "" {
		err = ioutil.WriteFile(outputFile, data, 0775)
		if err != nil {
			return "", errors.Annotate(err, "Error writing results to file")
		}
		return fmt.Sprintf("Done! %d results written to %s \n", len(results), outputFile), nil
	}

	return string(data), nil
}
