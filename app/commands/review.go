package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/app/commands/review"

	"github.com/codelingo/lingo/app/util"

	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"
)

func init() {
	register(&cli.Command{
		Name:        "review",
		Usage:       "review code following tenets in .lingo.",
		Subcommands: cli.Commands{*pullRequestCmd},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.LingoFile.String(),
				Usage: "A .lingo file to perform the review with. If the flag is not set, .lingo files are read from the branch being reviewed",
			},
			cli.StringFlag{
				Name:  util.DiffFlg.String(),
				Usage: "Review only unstaged changes in the working tree",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "File to save found issues to",
			},
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Value: "json-pretty",
				Usage: "How to format the found issues. Possible values are: json, json-pretty",
			},
			cli.BoolFlag{
				Name:  util.InteractiveFlg.String(),
				Usage: "Be prompted to confirm each issue",
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
	if err := initRepo(ctx); err != nil {
		// TODO(waigani) use error types
		// Note: Prior repo init is a valid state.
		if !strings.Contains(err.Error(), "already exists") {
			return "", errors.Trace(err)
		}
	}

	dotlingo, err := readDotLingo(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	opts := review.Options{
		FilesAndDirs: ctx.Args(),
		Diff:         ctx.Bool("diff"),
		SaveToFile:   ctx.String("output"),
		KeepAll:      !ctx.Bool("interactive"),
		DotLingo:     dotlingo,
	}

	issues, err := review.Review(opts)
	if err != nil {
		return "", errors.Trace(err)
	}

	if len(issues) == 0 {
		return fmt.Sprintf("Done! No issues found.\n"), nil
	}

	var data []byte
	format := ctx.String("format")
	switch format {
	case "json":
		data, err = json.Marshal(issues) //json.Marshal(issues)
		if err != nil {
			return "", errors.Trace(err)
		}
	case "json-pretty":
		data, err = json.MarshalIndent(issues, " ", " ") //json.Marshal(issues)
		if err != nil {
			return "", errors.Trace(err)
		}
	default:
		return "", errors.Errorf("Unknown format %q", format)
	}

	if outputFile := opts.SaveToFile; outputFile != "" {
		err = ioutil.WriteFile(outputFile, data, 0775)
		if err != nil {
			return "", errors.Annotate(err, "Error writing issues to file")
		}
		return fmt.Sprintf("Done! %d issues written to %s \n", len(issues), outputFile), nil
	}

	return string(data), nil
}

func readDotLingo(ctx *cli.Context) (string, error) {
	var dotlingo []byte

	if filename := ctx.String(util.LingoFile.Long); filename != "" {
		var err error
		dotlingo, err = ioutil.ReadFile(filename)
		if err != nil {
			return "", errors.Trace(err)
		}
	}
	return string(dotlingo), nil
}

// TODO(waigani) start ingesting repo as soon as it's inited

func initRepo(ctx *cli.Context) error {

	repo := vcs.New(backing.Git)
	authCfg, err := util.AuthConfig()
	if err != nil {
		return errors.Trace(err)
	}

	if err = repo.AssertNotTracked(); err != nil {
		// TODO (benjamin-rood): Check the error type
		return errors.Trace(err)
	}

	// TODO(waigani) Try to get owner and name from origin remote first.

	// get the repo owner name
	repoOwner, err := authCfg.Get("gitserver.user.username")
	if err != nil {
		return errors.Trace(err)
	}

	// get the repo name, default to working directory name
	dir, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	repoName := filepath.Base(dir)

	// TODO(benjamin-rood) Check if repo name and contents are identical.
	// If, so this should be a no-op and only the remote needs to be set.
	// ensure creation of distinct remote.
	repoName, err = createRepo(repo, repoName)
	if err != nil {
		return errors.Trace(err)
	}
	_, _, err = repo.SetRemote(repoOwner, repoName)
	return err
}

func createRepo(repo backing.Repo, name string) (string, error) {
	if err := repo.CreateRemote(name); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			parts := strings.Split(name, "-")
			num := parts[len(parts)-1]

			// We ignore the error here because the only case in which Atoi
			// would error is if the name had not yet been appended with -n.
			// In this case, n will be set to zero which is what we require.
			n, _ := strconv.Atoi(num)
			if n != 0 {
				// Need to remove existing trailing number where present,
				// otherwise the repoName only appends rather than replaces
				// and will produce names of the pattern "myPkg-1-2-...-n-n+1".
				name = strings.TrimSuffix(name, "-"+num)
			}
			return createRepo(repo, fmt.Sprintf("%s-%d", name, n+1))
		}
		return "", err
	}
	return name, nil
}
