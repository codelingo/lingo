package app

import (
	"fmt"
	"strings"

	"github.com/codelingo/lingo/app/commands"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/juju/errors"
	"github.com/urfave/cli"
)

// TODO(waigani) have a global state that tenets can share. An issue may be
// found based on the combination of tenets.
// https://github.com/codelingo/demo/issues/1

func New() *cli.App {
	// setCommandHelpTemplate()
	app := cli.NewApp()
	app.Name = "lingo"
	app.Usage = "Code Quality That Scales."
	app.Before = commands.Before
	app.Commands = commands.All()
	app.Version = common.ClientVersion
	// TODO(waigani) once messaging is implemented, add -q flag to suppress them here.
	// app.Flags = common.GlobalOptions
	app.CommandNotFound = func(c *cli.Context, command string) {
		util.FatalOSErr(errors.Errorf("'%s' is not a lingo command. See 'lingo --help'.", command))
	}
	app.EnableBashCompletion = true

	return app
}

func setCommandHelpTemplate(args ...string) {
	var argStr string
	if len(args) == 0 {
		argStr = "[arguments...]"
	} else {
		argStr = "<" + strings.Join(args, "> <") + ">"
	}
	cli.CommandHelpTemplate = fmt.Sprintf(`
NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   lingo {{.Name}}{{if .Flags}} [options]{{end}} %s{{if .Description}}

EXAMPLES:
   {{.Description}}{{end}}{{if .Flags}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}
`[1:], argStr)
}
