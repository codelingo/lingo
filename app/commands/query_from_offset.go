package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"context"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/juju/errors"
)

// TODO(BlakeMScurr) usage info should be generated from cli
const usage = `
lingo query-from-offset - Generate CLQL to match segment of code within a given file
 
USAGE:
	lingo query-from-offset <filename> <start> <end>`

func queryFromOffsetAction(ctx *cli.Context) {
	err := queryFromOffset(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

type genFact struct {
	FactName   string                 `json:"fact_name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Children   []*genFact             `json:"children,omitempty"`
}

func queryFromOffset(cliCtx *cli.Context) error {
	badArgsErr := errors.New(usage)
	if len(cliCtx.Args()) != 3 {
		return badArgsErr
	}

	file, err := validateFilePath(cliCtx.Args()[0])
	if err != nil {
		return errors.Annotate(badArgsErr, err.Error())
	}

	start, err := strconv.ParseInt(cliCtx.Args()[1], 10, 64)
	if err != nil {
		return errors.Annotate(badArgsErr, "start must be an integer")
	}

	end, err := strconv.ParseInt(cliCtx.Args()[2], 10, 64)
	if err != nil {
		return errors.Annotate(badArgsErr, "end must be an integer")
	}

	if start > end {
		return errors.Annotate(badArgsErr, "start must be smaller than end")
	}

	lang := filepath.Ext(file)[1:]
	dir := filepath.Base(filepath.Dir(file))
	filename := filepath.Base(file)
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Trace(err)
	}

	ctx, _ := util.UserCancelContext(context.Background())
	src := string(contents[:])
	reply, err := service.QueryFromOffset(ctx, &codelingo.QueryFromOffsetRequest{
		Lang:     lang,
		Dir:      dir,
		Filename: filename,
		Src:      src,
		Start:    start,
		End:      end,
	})
	if err != nil {
		return errors.Annotate(badArgsErr, err.Error())
	}

	facts := make([]*genFact, len(reply.Facts))
	for i, fact := range reply.Facts {
		gFact, err := buildGenFact(cliCtx, fact)
		if err != nil {
			return errors.Trace(err)
		}
		facts[i] = gFact
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(facts)
	content := buf.Bytes()
	outputBytes("", content)
	return nil
}

// buildGenFact builds a new genFact from the codelingo.GenFact.
// Properties are kept according to the cli options.
func buildGenFact(cliCtx *cli.Context, fact *codelingo.GenFact) (*genFact, error) {
	gFact := &genFact{
		FactName: fact.FactName,
		Children: make([]*genFact, len(fact.Children)),
	}

	if cliCtx.Bool("all-properties") || (cliCtx.Bool("final-fact-properties") && len(fact.Children) == 0) {
		gFact.Properties = make(map[string]interface{})
		for name, prop := range fact.Properties {
			iProp, err := genPropToInterface(prop)
			if err != nil {
				return nil, errors.Trace(err)
			}
			gFact.Properties[name] = iProp
		}
	}

	for i, cChild := range fact.Children {
		gChild, err := buildGenFact(cliCtx, cChild)
		if err != nil {
			return nil, errors.Trace(err)
		}
		gFact.Children[i] = gChild
	}

	return gFact, nil
}

func genPropToInterface(property *codelingo.GenProperty) (interface{}, error) {
	switch val := property.Value.(type) {
	case *codelingo.GenProperty_Int:
		return val.Int, nil
	case *codelingo.GenProperty_Float:
		return val.Float, nil
	case *codelingo.GenProperty_Bool:
		return val.Bool, nil
	case *codelingo.GenProperty_String_:
		return val.String_, nil
	}
	return nil, errors.Errorf("unknown property type %T", property.Value)
}

func validateFilePath(path string) (string, error) {
	dirPath := filepath.Dir(path)
	fileName := filepath.Base(path)

	// Check that it exists and is a directory
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("file %q not found", path))
	}

	abs, err := filepath.Abs(filepath.Join(dirPath, fileName))
	if err != nil {
		return "", errors.Trace(err)
	}
	return abs, nil
}
