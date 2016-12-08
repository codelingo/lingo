package commands

import (
	"bytes"
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/server"
	"github.com/juju/errors"
	"io/ioutil"
	"path/filepath"
)

func init() {
	register(&cli.Command{
		Name:   "query-from-offset",
		Usage:  "Generate CLQL query to match code in a specific section of a file",
		Action: pathFromOffsetAction,
		Flags: []cli.Flag{
			cli.Uint64Flag{
				Name:  util.StartFlg.String(),
				Usage: "The start of the range to find queries in.",
			},
			cli.Uint64Flag{
				Name:  util.EndFlg.String(),
				Usage: "The end of the range to find queries in.",
			},
			cli.StringFlag{
				Name:  util.InputFlg.String(),
				Usage: "The filepath of the file for which you want to generate queries.",
			},
		},
	}, false, versionRq)
}

func pathFromOffsetAction(ctx *cli.Context) {
	err := pathFromOffset(ctx)
	if err != nil {
		util.OSErrf(err.Error())
		return
	}
}

func pathFromOffset(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}
	start := ctx.Int("start")
	end := ctx.Int("end")
	file, err := getFilePath(ctx.String("input"))
	if err != nil {
		return errors.Trace(err)
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
		return errors.Trace(err)
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(paths.Paths)
	content := buf.Bytes()
	outputBytes("", content)
	return nil
}
