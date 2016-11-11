package commands

import (
	//"bytes"
	//"encoding/json"
	//"fmt"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	//"io/ioutil"
	//"strings"

	"github.com/juju/errors"
	//"path/filepath"
	//"os"
)

func init() {
	register(&cli.Command{
		Name:   "list-facts",
		Usage:  "List available facts for a given lexicon",
		Action: listFactsAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Usage: "The format for the output. Can be plain golang structs (default) or JSON",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "A filepath to output lexicon data to. If the flag is not set, outputs to cli",
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
		lexicon = "codelingo/php"
	}

	facts, err := svc.ListFacts(lexicon)
	if err != nil {
		return errors.Trace(err)
	}
	err = outputBytes(ctx.String("output"), getFormat(ctx.String("format"), facts))
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// TODO(BlakeMScurr) this logic be needed (in a reworked form) to decode actual
// facts struct but as we are currently using []string, we can reuse
// logic from list_lexicons

// func getFormat(format string, lexicons []string) []byte {
// 	var content []byte
// 	switch format {
// 	case "json":
// 		var buf bytes.Buffer
// 		json.NewEncoder(&buf).Encode(lexicons)
// 		content = buf.Bytes()
// 	default:
// 		// TODO(BlakeMScurr) append more efficiently
// 		str := strings.Join(lexicons, ", ")
// 		str += "\n"
// 		content = []byte(str)
// 	}
// 	return content
// }

// func outputBytes(output string, content []byte) error {
// 	if output == "" {
// 		fmt.Print(string(content))
// 		return nil
// 	}

// 	outputPath, err := getFilePath(output)
// 	if err != nil {
// 		return errors.Trace(err)
// 	}

// 	if _, err := os.Stat(outputPath); err == nil {
// 		return errors.Trace(err)
// 	}

// 	return errors.Trace(ioutil.WriteFile(outputPath, content, 0644))
// }

// func getFilePath(path string) (string, error) {
// 	dirPath := filepath.Dir(path)
// 	fileName := filepath.Base(path)

// 	// Check that it exists and is a directory
// 	if pathInfo, err := os.Stat(dirPath); os.IsNotExist(err) {
// 		return "", errors.Annotatef(err, "directory %q not found", dirPath)
// 	} else if !pathInfo.IsDir() {
// 		return "", errors.Errorf("%q is not a directory", dirPath)
// 	}

// 	return filepath.Join(dirPath, fileName), nil
// }
