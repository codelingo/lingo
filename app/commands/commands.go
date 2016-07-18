package commands

import (
	"fmt"
	"strings"

	"github.com/juju/errors"

	"github.com/codegangsta/cli"
)

var cmds = map[require][]*cli.Command{}

// returns a map of cmd name to list of requirements
func cmdRequirements(c map[require][]*cli.Command) map[string][]require {
	req := map[string][]require{}

	for r, cmdList := range c {
		for _, cmd := range cmdList {
			req[cmd.Name] = append(req[cmd.Name], r)
		}
	}
	return req
}

func register(cmd *cli.Command, req ...require) {
	cmds[baseRq] = append(cmds[baseRq], cmd)
	for _, r := range req {
		cmds[r] = append(cmds[r], cmd)
	}
}

func All() []cli.Command {
	var all []cli.Command
	for _, cmd := range cmds[baseRq] {
		all = append(all, *cmd)
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
	}

	// No requirements should be needed to show help
	if isHelpAlias(flags) {
		return nil
	}

	if reqs, ok := cmdReq[currentCMDName]; ok {
		for _, req := range reqs {
			if err := req.Verify(); err != nil {
				fmt.Printf(err.Error())
				return nil
			}
		}
	} else {
		return errors.Errorf("command %q not found", currentCMDName)
	}
	return nil
}

// isHelpAlias returns true when a command's arguments are equivalent to the
// help command. For example, `lingo review --help` == `lingo help review`.
func isHelpAlias(flags []string) bool {
	helpFlags := strings.Split(cli.HelpFlag.Name, ", ")
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
