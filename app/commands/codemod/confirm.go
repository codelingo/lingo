package codemod

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/rpc/flow"
	"github.com/fatih/color"
	"github.com/waigani/diffparser"
)

type IssueConfirmer struct {
	keepAll bool
	output  bool
}

func NewConfirmer(output, keepAll bool, d *diffparser.Diff) (*IssueConfirmer, error) {
	cfm := IssueConfirmer{
		keepAll: keepAll,
		output:  output,
	}

	return &cfm, nil
}

func GetDiffRootPath(filename string) string {
	// Get filename relative to git root folder
	// TODO: Handle error in case of git not being installed
	// https://github.com/codelingo/demo/issues/2
	out, err := exec.Command("git", "ls-tree", "--full-name", "--name-only", "HEAD", filename).Output()
	if err == nil && len(out) != 0 {
		if len(out) != 0 {
			filename = strings.TrimSuffix(string(out), "\n")
		}
	}
	return filename
}

var editor string

// confirm returns true if the issue should be kept or false if it should be
// dropped.
func (c IssueConfirmer) Confirm(attempt int, issue *flow.Issue, newSRC string) bool {
	if c.keepAll {
		return true
	}
	if attempt == 0 {
		fmt.Println(c.FormatPlainText(issue, newSRC))
	}
	attempt++
	var options string
	fmt.Print("\n[o]pen")
	if c.output {
		fmt.Print(" [k]eep [R]eplace")
	}
	fmt.Print(": ")

	fmt.Scanln(&options)

	switch options {
	case "o":
		var app string
		defaultEditor := "vi" // TODO(waigani) use EDITOR or VISUAL env vars
		// https://github.com/codelingo/demo/issues/3
		if editor != "" {
			defaultEditor = editor
		}
		fmt.Printf("application (%s):", defaultEditor)
		fmt.Scanln(&app)
		filename := issue.Position.Start.Filename
		if app == "" {
			app = defaultEditor
		}
		// c := issue.Position.Start.Column // TODO(waigani) use column
		// https://github.com/codelingo/demo/issues/4
		l := issue.Position.Start.Line
		cmd, err := util.OpenFileCmd(app, filename, l)
		if err != nil {
			fmt.Println(err)
			return c.Confirm(attempt, issue, newSRC)
		}

		if err = cmd.Start(); err != nil {
			log.Println(err)
		}
		if err = cmd.Wait(); err != nil {
			log.Println(err)
		}

		editor = app

		c.Confirm(attempt, issue, newSRC)
	case "k":
		issue.Discard = true

		// TODO(waigani) only prompt for reason if we're sending to a service.
		// https://github.com/codelingo/demo/issues/5
		fmt.Print("reason: ")
		in := bufio.NewReader(os.Stdin)
		issue.DiscardReason, _ = in.ReadString('\n')

		// TODO(waigani) we are now always returning true. Need to decide
		// how caller will deal with removing isseus, ie. KeptIssues vs AllIssues,
		// then being returning false here
		// https://github.com/codelingo/demo/issues/6
		return true
	case "", "r", "R", "\n":
		return true
	default:
		fmt.Printf("invalid input: %s\n", options)
		fmt.Println(options)
		c.Confirm(attempt, issue, newSRC)
	}

	// TODO(waigani) build up issues here.

	return true
}

func (c *IssueConfirmer) FormatPlainText(issue *flow.Issue, newSRC string) string {
	m := color.New(color.FgWhite, color.Faint).SprintfFunc()
	y := color.New(color.FgRed).SprintfFunc()
	yf := color.New(color.FgWhite, color.Faint).SprintfFunc()
	g := color.New(color.FgGreen).SprintfFunc()
	filename := issue.Position.Start.Filename

	addrFmtStr := fmt.Sprintf("%s:%d", filename, issue.Position.End.Line)
	if col := issue.Position.End.Column; col != 0 {
		addrFmtStr += fmt.Sprintf(":%d", col)
	}
	address := m(addrFmtStr)

	ctxBefore := indent(yf("\n...\n%s", issue.CtxBefore), false, false)
	oldLines := indent(y("\n%s", issue.LineText), false, true)

	newLines := indent(g("\n%s", newSRC), true, false)
	ctxAfter := indent(yf("\n%s\n...", issue.CtxAfter), false, false)
	src := ctxBefore + oldLines + newLines + ctxAfter

	return fmt.Sprintf("%s\n%s", address, src)
}

func indent(str string, add, remove bool) string {
	replace := "\n    "
	if add {
		replace = "\n  + "
	}
	if remove {
		replace = "\n  - "
	}
	return strings.Replace(str, "\n", replace, -1)
}
