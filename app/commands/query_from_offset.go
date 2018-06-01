package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/server"
	"github.com/juju/errors"
)

// TODO(BlakeMScurr) usage info should be generated from cli
const usage = `
lingo query-from-facts - Generate CLQL to match segment of code within a given file
 
USAGE:
 	lingo query-from-facts <filename> <start> <end>`

func pathFromOffsetAction(ctx *cli.Context) {
	err := pathFromOffset(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func pathFromOffset(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}
	badArgsErr := errors.New(usage)
	if len(ctx.Args()) != 3 {
		return badArgsErr
	}

	file, err := validateFilePath(ctx.Args()[0])
	if err != nil {
		return errors.Annotate(badArgsErr, err.Error())
	}

	start, err := strconv.Atoi(ctx.Args()[1])
	if err != nil {
		return errors.Annotate(badArgsErr, "start must be an integer")
	}

	end, err := strconv.Atoi(ctx.Args()[2])
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

	src := string(contents[:])
	paths, err := svc.PathsFromOffset(&server.PathsFromOffsetRequest{
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

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(paths.Paths)
	content := buf.Bytes()
	outputBytes("", content)
	return nil
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
