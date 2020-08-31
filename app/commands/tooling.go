package commands

import (
	"bytes"
	"encoding/json"
	"strings"

	"context"

	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
	"github.com/urfave/cli"
)

func init() {
	register(&cli.Command{
		Hidden: true,
		Name:   "tooling",
		Usage:  "Commands under tooling are intended to support the other software such as IDEs and are not intended to be called by the user directly from the CLI",
		Subcommands: []cli.Command{
			{
				Name:   "list-facts",
				Usage:  "List available facts for a given lexicon.",
				Action: listFactsAction,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "debug",
						Usage: "Display debug messages",
					},
					cli.StringFlag{
						Name:  util.FormatFlg.String(),
						Usage: "The format for the output. Can be listed (default) or \"json\" encoded.",
					},
					cli.StringFlag{
						Name:  util.OutputFlg.String(),
						Usage: "A filepath to output lexicon data to. If the flag is not set, outputs to cli.",
					},
					cli.StringFlag{
						Name:  util.VersionFlg.String(),
						Usage: "The version of the lexicon. Leave empty for the latest version.",
					},
				},
			},
			{
				Name:   "query-from-offset",
				Usage:  "Generate CLQL query to match code in a specific section of a file.",
				Action: queryFromOffsetAction,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "debug",
						Usage: "Display debug messages",
					},
					cli.BoolFlag{
						Name:  "all-properties, a",
						Usage: "List all properties of all facts in the query path",
					},
					cli.BoolFlag{
						Name:  "final-fact-properties, f",
						Usage: "List all properties of the final fact in the query path",
					},
					cli.BoolFlag{
						Name:  util.InsecureFlg.String(),
						Usage: "Allow command to run against an insecure development environment",
					},
				},
			},
		},
	}, false, false, verify.VersionRq)
}

func listFactsAction(ctx *cli.Context) {
	err := listFacts(ctx)
	if err != nil {
		util.Logger.Debug(errors.ErrorStack(err))
		util.FatalOSErr(err)
		return
	}
}

func listFacts(cliCtx *cli.Context) error {
	var owner, name, lexicon string
	if len(cliCtx.Args()) > 0 {
		lexicon = cliCtx.Args()[0]
	}

	if args := strings.Split(lexicon, "/"); len(args) == 2 {
		owner = args[0]
		name = args[1]
	} else {
		return errors.New("Please specify a properly namespaced lexicon, ie,\nlingo lexicons list-facts codelingo/go")
	}

	ctx, _ := util.UserCancelContext(context.Background())
	facts, err := service.ListFacts(ctx, owner, name, cliCtx.String("version"))
	if err != nil {
		return errors.Trace(err)
	}

	byt := getFactFormat(cliCtx.String("format"), facts)

	err = outputBytes(cliCtx.String("output"), byt)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func formatFacts(facts map[string][]string) string {
	// TODO(BlakeMScurr) use a string builder and optimise this
	ret := ""
	for key, fact := range facts {
		ret += key
		ret += "\n"
		for child := range fact {
			ret += "\t"
			ret += fact[child]
			ret += "\n"
		}
	}

	return ret
}

// TODO(BlakeMScurr) Refactor this and getFormat (from list_lexicons)
// which very similar logic
func getFactFormat(format string, facts map[string][]string) []byte {
	var content []byte
	switch format {
	case "json":
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(facts)
		content = buf.Bytes()
	default:
		content = []byte(formatFacts(facts))
	}
	return content
}
