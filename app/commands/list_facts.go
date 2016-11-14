package commands

import (
	"bytes"
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/juju/errors"
	"strings"
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

func formatFacts(factList *codelingo.FactList) []string {
	// TODO(BlakeMScurr) check and optimise append and concat efficiency
	ret := []string{}

	for _, branchFact := range factList.Facts {
		ret = append(ret, indentFacts(branchFact, "")...)
	}
	return ret
}

func indentFacts(facts *codelingo.Fact, tabs string) []string {
	// TODO(BlakeMScurr) check and optimise append and concat efficiency

	ret := []string{tabs + facts.Kind}
	for _, prop := range facts.Properties {
		newStr := tabs + "\t" + prop
		ret = append(ret, newStr)
	}
	for _, branchFact := range facts.Facts {
		ret = append(ret, indentFacts(branchFact, tabs+"\t")...)
	}
	return ret

}

// TODO(BlakeMScurr) Refactor this and getFormat (from list_lexicons)
// which very similar logic
func getFactFormat(format string, facts *codelingo.FactList) []byte {
	var content []byte
	switch format {
	case "json":
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(facts)
		content = buf.Bytes()
	default:
		// TODO(BlakeMScurr) append more efficiently
		content = []byte(strings.Join(formatFacts(facts), "\n") + "\n")
	}
	return content
}
