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

func pathFromOffsetAction(ctx *cli.Context) {
	err := pathFromOffset(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func pathFromOffset(cliCtx *cli.Context) error {
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
	paths, err := service.PathsFromOffset(ctx, &codelingo.PathsFromOffsetRequest{
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

	filterPaths(paths, cliCtx)

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(paths.Paths)
	content := buf.Bytes()
	outputBytes("", content)
	return nil
}

func filterPaths(paths *codelingo.PathsFromOffsetReply, ctx *cli.Context) {
	if ctx.Bool("all-properties") {
		return
	}

	for _, path := range paths.Paths {
		for i, fact := range path.Facts {
			if ctx.Bool("final-fact-properties") {
				if i+1 != len(path.Facts) {
					fact.Properties = make(map[string]string)
				}
			} else {
				fact.Properties = make(map[string]string)
			}
		}
	}
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
