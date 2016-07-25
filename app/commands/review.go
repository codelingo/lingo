package commands

import (
	"fmt"

	"github.com/codelingo/lingo/app/commands/review"

	"github.com/codegangsta/cli"
)

func init() {
	register(&cli.Command{
		Name:  "review",
		Usage: "review code following tenets in .lingo",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "review all files under all directories from pwd down",
			}},
		Description: `

"$ lingo review" will review any unstaged changes from pwd down.
"$ lingo review [<filename>]" will review any unstaged changes in the named files.
"$ lingo review --all [<filename>]" will review all code in the named files.
"$ lingo review --all" will review all code from pwd down.

`[1:],
		Action: reviewAction,
	},
		vcsRq, dotLingoRq, homeRq, authRq, configRq,
	)
}

func reviewAction(ctx *cli.Context) {

	opts := review.Options{
		Files:      ctx.Args(),
		Diff:       ctx.Bool("diff"),
		SaveToFile: ctx.String("save"),
		KeepAll:    ctx.Bool("keep-all"),
	}
	issues, err := review.Review(opts)
	if err != nil {
		fmt.Println(err.Error())
		return
		// errors.ErrorStack(err)
	}

	fmt.Printf("Done! Found %d issues \n", len(issues))
}
