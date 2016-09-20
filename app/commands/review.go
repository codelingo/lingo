package commands

import (
	"fmt"

	"github.com/codelingo/lingo/app/commands/review"

	"github.com/codelingo/lingo/app/util"

	"github.com/codegangsta/cli"
)

func init() {
	register(&cli.Command{
		Name:        "review",
		Usage:       "review code following tenets in .lingo",
		Subcommands: cli.Commands{*pullRequestCmd},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.LingoFile.String(),
				Usage: "A list of .lingo files to preform the review with. If the flag is not set, .lingo files are read from the branch being reviewed.",
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
		vcsRq, dotLingoRq, homeRq, authRq, configRq,
	)
}

func reviewAction(ctx *cli.Context) {

	opts := review.Options{
		FilesAndDirs: ctx.Args(),
		Diff:         ctx.Bool("diff"),
		SaveToFile:   ctx.String("save"),
		KeepAll:      ctx.Bool("keep-all"),
	}

	issues, err := review.Review(opts)
	if err != nil {
		fmt.Println(err.Error())
		return
		// errors.ErrorStack(err)
	}

	fmt.Printf("Done! Found %d issues \n", len(issues))
}
