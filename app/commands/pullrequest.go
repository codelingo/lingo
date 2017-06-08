package commands

import (
	"fmt"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/bot/clair"
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
	msg, err := reviewPullRequestCMD(ctx)
	if err != nil {
		// Debugging
		// print(errors.ErrorStack(err))
		util.OSErr(err)
		return
	}
	fmt.Println(msg)
}

func reviewPullRequestCMD(ctx *cli.Context) (string, error) {
	if l := len(ctx.Args()); l != 1 {
		return "", errors.Errorf("expected one arg, got %d", l)
	}

	dotlingo, err := review.ReadDotLingo(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	opts, err := review.ParsePR(ctx.Args()[0])
	if err != nil {
		return "", errors.Trace(err)
	}

	issuec, err := clair.Review(clair.Request{
		PullRequest: opts,
		DotLingo:    dotlingo,
	})
	if err != nil {
		return "", errors.Trace(err)
	}

	issues, err := review.ConfirmIssues(issuec, ctx.Bool("keep-all"), ctx.String("save"))
	if err != nil {
		return "", errors.Trace(err)
	}

	// TODO: streaming back to the client, verify issues on the client side.
	if len(issues) == 0 {
		return "Done! No issues found.\n", nil
	}

	msg, err := review.MakeReport(issues, ctx.String("format"), ctx.String("save"))
	if err != nil {
		return "", errors.Trace(err)
	}

	fmt.Println(fmt.Printf("Done! Found %d issues \n", len(issues)))
	return msg, nil
}
