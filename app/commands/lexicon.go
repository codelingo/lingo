package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	register(&cli.Command{
		Name:   "lexicon",
		Usage:  "Explain what lexicons are and how to use them.",
		Action: lexiconAction,
		Subcommands: []cli.Command{
			{
				Name:   "list",
				Usage:  "List available lexicons.",
				Action: lexiconListAction,
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
			},
			{
				Name:   "list-facts",
				Usage:  "List available facts for a given lexicon.",
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
					cli.StringFlag{
						Name:  util.VersionFlg.String(),
						Usage: "The version of the lexicon. Leave empty for the latest version.",
					},
				},
			},
		},
	}, false, versionRq)
}

func lexiconAction(ctx *cli.Context) {
	err := lexicon(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func lexicon(ctx *cli.Context) error {
	fmt.Println("A lexicon is a collection of facts used to describe a domain of knowledge.")
	return nil
}

func lexiconListAction(ctx *cli.Context) {
	err := listLexicons(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func listLexicons(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}

	// TODO: discovery layer
	lexicons, err := svc.ListLexicons()
	if err != nil {
		return errors.Trace(err)
	}

	err = outputBytes(ctx.String("output"), getFormat(ctx.String("format"), lexicons))
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func getFormat(format string, lexicons []string) []byte {
	var content []byte
	switch format {
	case "json":
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(lexicons)
		content = buf.Bytes()
	default:
		// TODO(BlakeMScurr) append more efficiently
		str := strings.Join(lexicons, "\n")
		str += "\n"
		content = []byte(str)
	}
	return content
}

func outputBytes(output string, content []byte) error {
	if output == "" {
		fmt.Print(string(content))
		return nil
	}

	outputPath, err := getFilePath(output)
	if err != nil {
		return errors.Trace(err)
	}

	if _, err := os.Stat(outputPath); err == nil {
		return errors.Trace(err)
	}

	return errors.Trace(ioutil.WriteFile(outputPath, content, 0644))
}

func getFilePath(path string) (string, error) {
	dirPath := filepath.Dir(path)
	fileName := filepath.Base(path)

	// Check that it exists and is a directory
	if pathInfo, err := os.Stat(dirPath); os.IsNotExist(err) {
		return "", errors.Annotatef(err, "directory %q not found", dirPath)
	} else if !pathInfo.IsDir() {
		return "", errors.Errorf("%q is not a directory", dirPath)
	}

	abs, err := filepath.Abs(filepath.Join(dirPath, fileName))
	if err != nil {
		return "", errors.Trace(err)
	}
	return abs, nil
}

func listFactsAction(ctx *cli.Context) {
	err := listFacts(ctx)
	if err != nil {
		util.OSErr(err)
		return
	}
}

func listFacts(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}

	var owner, name, lexicon string
	if len(ctx.Args()) > 0 {
		lexicon = ctx.Args()[0]
	}

	if args := strings.Split(lexicon, "/"); len(args) == 2 {
		owner = args[0]
		name = args[1]
	} else {
		return errors.New("Please specify a properly namespaced lexicon, ie,\nlingo lexicon list-facts codelingo/go")
	}

	facts, err := svc.ListFacts(owner, name, ctx.String("version"))
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
		content = []byte(formatFacts(facts))
	}
	return content
}
