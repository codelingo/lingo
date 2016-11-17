package commands

import (
	"bytes"
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Name:   "list-facts",
		Usage:  "List available facts for a given lexicon",
		Action: listFactsAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Usage: "The format for the output. Can be listed (default) or \"json\" encoded.",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "A filepath to output lexicon data to. If the flag is not set, outputs to cli.",
			},
		},
	}, false)
}

func listFactsAction(ctx *cli.Context) {
	err := listFacts(ctx)
	if err != nil {
		util.OSErrf(err.Error())
		return
	}
}

func listFacts(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}

	var lexicon string
	if len(ctx.Args()) > 0 {
		lexicon = ctx.Args()[0]
	} else {
		lexicon = "codelingo/golang"
	}

	facts, err := svc.ListFacts(lexicon)
	if err != nil {
		return errors.Trace(err)
	}

	byt := getFactFormat(ctx.String("format"), facts)

	err = outputBytes(ctx.String("output"), byt)
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
		// TODO(BlakeMScurr) append more efficiently
		content = []byte(formatFacts(facts))
	}
	return content
}
