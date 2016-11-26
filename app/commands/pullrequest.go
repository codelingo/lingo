package commands

import (
	"fmt"

	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"

	"github.com/codelingo/lingo/app/commands/review"

	"github.com/codegangsta/cli"
)

var pullRequestCmd = &cli.Command{
	Name:      "pull-request",
	ShortName: "pr",
	Usage:     "review a remote pull-request",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  util.LingoFile.String(),
			Usage: "A list of .lingo files to preform the review with. If the flag is not set, .lingo files are read from the branch being reviewed.",
		},
		// TODO(waigani) as this is a review sub-command, it should be able to use the
		// lingo-file flag from review.
		// cli.BoolFlag{
		// 	Name:  "all",
		// 	Usage: "review all files under all directories from pwd down",
		// },
	},
	Description: `
"$ lingo review pull-request https://github.com/codelingo/lingo/pull/1" will review all code in the diff between the pull request and it's base repository.
"$ lingo review pr https://github.com/codelingo/lingo/pull/1" will review all code in the diff between the pull request and it's base repository.
`[1:],
	// "$ lingo review" will review any unstaged changes from pwd down.
	// "$ lingo review [<filename>]" will review any unstaged changes in the named files.
	// "$ lingo review --all [<filename>]" will review all code in the named files.
	Action: reviewPullRequestAction,
}

func init() {
	register(pullRequestCmd,
		true,
		homeRq, authRq, configRq, versionRq,
	)
}

func reviewPullRequestAction(ctx *cli.Context) {
	if l := len(ctx.Args()); l != 1 {
		fmt.Printf("expected one arg, got %d", l)
		return
	}

	dotlingo, err := readDotLingo(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	opts := review.Options{
		PullRequest: ctx.Args()[0],
		SaveToFile:  ctx.String("save"),
		KeepAll:     ctx.Bool("keep-all"),
		DotLingo:    dotlingo,
	}

	issueCount, err := review.Review(opts)
	if err != nil {
		fmt.Println(errors.ErrorStack(err))
		return
		// errors.ErrorStack(err)
	}

	fmt.Printf("Done! Found %d issues \n", issueCount)
}
