package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	rpc "github.com/codelingo/rpc/service"
	"github.com/juju/errors"
	"github.com/urfave/cli"
)

func init() {
	register(&cli.Command{
		Hidden: true,
		Name:   "describe-fact",
		Usage:  "Describe a fact belonging to a given lexicon.",
		Action: describeFactAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Usage: "The format for the output. Can be listed (default) or \"json\" encoded.",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "A filepath to output description to. If the flag is not set, outputs to cli.",
			},
			cli.StringFlag{
				Name:  util.VersionFlg.String(),
				Usage: "The version of the lexicon containing the fact. Leave empty for the latest version.",
			},
		},
	}, false, false, verify.VersionRq)
}

func describeFactAction(ctx *cli.Context) {
	err := describeFact(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func describeFact(cliCtx *cli.Context) error {
	var owner, name, lexicon, fact string
	if len(cliCtx.Args()) > 0 {
		lexicon = cliCtx.Args()[0]
	}

	if args := strings.Split(lexicon, "/"); len(args) == 3 {
		owner = args[0]
		name = args[1]
		fact = args[2]
	} else {
		return errors.New("Please specify a properly namespaced fact, ie,\nlingo describe-fact rpc/go/func_decl")
	}

	ctx, _ := util.UserCancelContext(context.Background())
	description, err := service.DescribeFact(ctx, owner, name, cliCtx.String("version"), fact)
	if err != nil {
		return errors.Trace(err)
	}

	byt := getDescriptionFormat(cliCtx.String("format"), description)

	err = outputBytes(cliCtx.String("output"), byt)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// TODO(BlakeMScurr) Refactor this and getFormat (from list_lexicons)
// and getFactFormat (from list_facts) which have very similar logic
func getDescriptionFormat(format string, output *rpc.DescribeFactReply) []byte {
	var content []byte
	switch format {
	case "json":
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(output)
		content = buf.Bytes()
	default:
		content = []byte(formatDescription(output))
	}
	return content
}

func formatDescription(description *rpc.DescribeFactReply) string {
	// TODO(BlakeMScurr) use a string builder and optimise this
	ret := "Description:\n\t"
	ret += description.Description
	ret += "\nExamples:\n\t"
	ret += description.Examples
	ret += "\nProperties:\n"

	for _, property := range description.Properties {
		ret += "\t" + property.Name + ": " + property.Description + "\n"
	}

	return ret
}
