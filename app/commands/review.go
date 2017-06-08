package commands

import (
	"fmt"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/bot/clair"

	"github.com/codelingo/lingo/app/commands/review"
	"github.com/codelingo/lingo/app/util"

	"os"

	"github.com/codegangsta/cli"
)

func init() {
	register(&cli.Command{
		Name:        "review",
		Usage:       "Review code following tenets in .lingo.",
		Subcommands: cli.Commands{*pullRequestCmd},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.LingoFile.String(),
				Usage: "A .lingo file to perform the review with. If the flag is not set, .lingo files are read from the branch being reviewed.",
			},
			cli.StringFlag{
				Name:  util.DiffFlg.String(),
				Usage: "Review only unstaged changes in the working tree.",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "File to save found issues to.",
			},
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Value: "json-pretty",
				Usage: "How to format the found issues. Possible values are: json, json-pretty.",
			},
			cli.BoolFlag{
				Name:  util.InteractiveFlg.String(),
				Usage: "Be prompted to confirm each issue.",
			},
			cli.StringFlag{
				Name:  util.DirectoryFlg.String(),
				Usage: "Review a given directory.",
			},

			// cli.BoolFlag{
			// 	Name:  "all",
			// 	Usage: "review all files under all directories from pwd down",
			// },
		},
		Description: `
"$ lingo review" will review all code from pwd down.
"$ lingo review <filename>" will only review named file.
`[1:],
		// "$ lingo review" will review any unstaged changes from pwd down.
		// "$ lingo review [<filename>]" will review any unstaged changes in the named files.
		// "$ lingo review --all [<filename>]" will review all code in the named files.
		Action: reviewAction,
	},
		false,
		vcsRq, homeRq, authRq, configRq, versionRq,
	)
}

func reviewAction(ctx *cli.Context) {
	msg, err := reviewCMD(ctx)
	if err != nil {

		// Debugging
		// print(errors.ErrorStack(err))
		util.OSErr(err)
		return
	}

	fmt.Println(msg)
}

func reviewCMD(ctx *cli.Context) (string, error) {
	dir := ctx.String("directory")
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return "", errors.Trace(err)
		}
	}

	dotlingo, err := review.ReadDotLingo(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	// Send review request to the bot layer.
	// When CLAIR is taken out of the client we will need to Init and Sync the repo from here,
	// as well as having the Init logic in the bot layer to recieve external resources.
	issuec, err := clair.Review(clair.Request{
		DotLingo: dotlingo,
	})

	if err != nil {
		return "", errors.Trace(err)
	}

	issues, err := review.ConfirmIssues(issuec, !ctx.Bool("interactive"), ctx.String("output"))
	if err != nil {
		return "", errors.Trace(err)
	}

	if len(issues) == 0 {
		return fmt.Sprintf("Done! No issues found.\n"), nil
	}

	msg, err := review.MakeReport(issues, ctx.String("format"), ctx.String("output"))
	return msg, errors.Trace(err)
}
