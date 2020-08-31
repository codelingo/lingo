package commands

import (
	"strings"

	"github.com/codelingo/lingo/app/commands/verify"
	"github.com/codelingo/lingo/app/util"
	"github.com/urfave/cli"

	"os"

	"github.com/juju/errors"
)

// "github.com/codelingo/lingo/app/util"
// "go.uber.org/zap"

type lingoCMD struct {
	isDevOnly bool
	isSubCMD  bool
	cmd       *cli.Command
}

var cmds = map[verify.Require][]*lingoCMD{}

// returns a map of cmd name to list of Requirements
func cmdRequirements(c map[verify.Require][]*lingoCMD) map[string][]verify.Require {
	req := map[string][]verify.Require{}

	for r, cmdList := range c {
		for _, lCMD := range cmdList {
			req[lCMD.cmd.Name] = append(req[lCMD.cmd.Name], r)
		}
	}
	return req
}

func register(cmd *cli.Command, isSubcommand, isDevOnly bool, req ...verify.Require) {
	lCMD := &lingoCMD{
		isDevOnly: isDevOnly,
		isSubCMD:  isSubcommand,
		cmd:       cmd,
	}
	cmds[verify.BaseRq] = append(cmds[verify.BaseRq], lCMD)
	for _, r := range req {
		cmds[r] = append(cmds[r], lCMD)
	}
}

func All() []cli.Command {
	isDev := os.Getenv("LINGO_DEV_CLI") == "true"

	var all []cli.Command
	for _, lCMD := range cmds[verify.BaseRq] {
		if !lCMD.isSubCMD && (!lCMD.isDevOnly || (lCMD.isDevOnly && isDev)) {
			all = append(all, *lCMD.cmd)
		}
	}
	return all
}

func Before(c *cli.Context) error {
	cmdReq := cmdRequirements(cmds)

	var currentCMDName string
	var flags []string
	args := c.Args()
	if args.Present() {
		currentCMDName = args.First()
		flags = args.Tail()

		for _, flag := range flags {
			if flag == "--debug" {
				err := util.SetDebugLogger()
				if err != nil {
					return errors.Trace(err)
				}
			}
		}

		// TODO(waigani) This is a manual hack to validate subcommands. Why is
		// c.Command.Name empty?
		subCMDs := map[string][]string{
			"review": []string{"pull-request", "pr"},
		}
		if subs, ok := subCMDs[args.First()]; ok {
			for _, subCMD := range subs {
				if c.Args().Get(1) == subCMD {
					currentCMDName = subCMD
					break
				}
			}
		}
	}
	// No Requirements should be needed to show help
	if currentCMDName == "help" || isHelpAlias(flags) {
		return nil
	}

	if reqs, ok := cmdReq[currentCMDName]; ok {
		for _, req := range reqs {
			if err := req.Verify(); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

// isHelpAlias returns true when a command's arguments are equivalent to the
// help command. For example, `lingo review --help` == `lingo help review`.
func isHelpAlias(flags []string) bool {
	helpFlags := strings.Split(cli.HelpFlag.GetName(), ", ")
	return len(flags) == 1 && (flags[0] == "--"+helpFlags[0] || flags[0] == "-"+helpFlags[1])
}

// // A list of cmds that need a .lingo file
// var cmdNeedsDotLingo = []string{
// 	"add",
// 	"remove",
// 	"rm",
// 	"review",
// 	"pull",
// 	"list",
// 	"ls",
// 	"write-docs",
// 	"docs",
// 	"edit",
// }

// var cmdNeedsLingoHome = []string{
// 	"build",
// 	"init",
// 	"add",
// 	"remove",
// 	"rm",
// 	"review",
// 	"pull",
// 	"list",
// 	"ls",
// 	"write-docs",
// 	"docs",
// 	"edit",
// 	"setup-auto-complete",
// 	"update",
// 	"config",
// }
