package clair

import (
	"bytes"
	"strconv"

	"github.com/codelingo/lingo/service/server"
	"github.com/codelingo/platform/controller"

	"github.com/codelingo/lingo/service/grpc/codelingo"

	"github.com/juju/errors"
)

// TODO(waigani) get scope working again
// scope = controller.ReviewScope{}
// TODO(waigani) modify query to set it's scope
// contextdir := lingoDir
// if strings.Compare(reviewScope.Dir, lingoDir) > 0 {
// 	contextdir = reviewScope.Dir
// }
// suffix := contextdir

// // Then target the working directory.
// // TODO this assumes dir nodes with names ".", "a" "a/b" etc, which should be a flexible function of VCS, AST etc lexicons.
// for suffix != "" {
// 	suffix = strings.SplitAfterN(suffix, "/", 2)[1]
// 	nodeName := strings.TrimSuffix(contextdir, "/"+suffix)
// 	startPath = startPath.Out(tree.ChildEdge).Has(quad.String("name"), quad.String(strings.TrimPrefix(nodeName, "./")))
// }

// bots modify queries going in and build results coming out

// TODO(waigani) move the bot to the client side - as it builds the query and the result on the client.

func BuildIssue(n *server.QueryResponse) (*codelingo.Issue, error) {

	comment := "mock-comment"
	name := "mock-tenet-name"

	startOffset, err := strconv.Atoi(n.Data["start_offset"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	startLine, err := strconv.Atoi(n.Data["start_line"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	startColumn, err := strconv.Atoi(n.Data["start_column"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	endOffset, err := strconv.Atoi(n.Data["end_offset"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	endLine, err := strconv.Atoi(n.Data["end_line"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	endColumn, err := strconv.Atoi(n.Data["end_column"][0])
	if err != nil {
		return nil, errors.Trace(err)
	}

	// we let this panic if filename is not set as it's a developer error.
	filename := n.Data["filename"][0]
	iRange := &codelingo.IssueRange{
		Start: &codelingo.Position{
			Filename: filename,
			Offset:   int64(startOffset),
			Line:     int64(startLine),
			Column:   int64(startColumn),
		},
		End: &codelingo.Position{
			Filename: filename,
			Offset:   int64(endOffset),
			Line:     int64(endLine),
			Column:   int64(endColumn),
		},
	}

	iss := &codelingo.Issue{
		Name:     name,
		Position: iRange,
		// TODO(waigani) make comment a slice
		Comment: comment,
		// Metrics
		Tags:    []string{"one", "two"},
		Link:    "some link",
		NewCode: true,
		Patch:   "some patch",
		Err:     "this is err",
	}

	// TODO(waigani) validate all pos info.
	if iss.Position.Start.Filename == "" {
		return nil, errors.New("no filename found for issue")
	}

	iss.CtxAfter = "line after"
	iss.CtxBefore += "line before"
	iss.LineText = "actual line"

	// TODO(waigani) set lines from graph

	// TODO(waigani) check what client is calling, if CLAIR, no need to set SRC
	// fileSRC, err := repo.ReadFile(iss.Position.Start.Filename, commitID)
	// if err != nil {
	// 	return nil, errors.Trace(err)
	// }
	// setIssueLines(fileSRC, iss)

	return iss, nil
}

// Note: this is not needed when setting comments on github.
func setIssueLines(fileSRC string, issue *codelingo.Issue) {
	srcBytes := []byte(fileSRC)

	// reset index to match file lines - easier to reason about.
	lineBytes := make([][]byte, len(srcBytes))
	for i, b := range bytes.Split(srcBytes, []byte("\n")) {
		lineBytes[i+1] = b
	}
	start := int(issue.Position.Start.Line)
	end := int(issue.Position.End.Line)

	buf := 2
	sStart := start - buf
	if sStart < 1 {
		sStart = 1
	}
	sEnd := start - 1
	if sEnd < sStart {
		sEnd = sStart
	}

	// TODO(waigani) demoware. Before will not show for start lines of 0,1 or 2
	// before
	for _, line := range lineBytes[sStart:sEnd] {
		issue.CtxBefore += string(line) + "\n"
	}

	// line
	if start == end {
		issue.LineText = string(lineBytes[start])
		// TODO(waigani) DEMOWARE. end+1 < start should never be hit. Find out
		// where it's coming from.
	} else if end+1 > start {
		for _, line := range lineBytes[start : end+1] {
			issue.LineText += string(line) + "\n"
		}
		// Trim final newline
		issue.LineText = issue.LineText[:len(issue.LineText)-1]
	}

	// after
	l := len(lineBytes)

	eStart := end + 1
	if eStart > l {
		eStart = l
	}

	eEnd := eStart + buf
	if eEnd > l {
		eEnd = l
	}

	for _, line := range lineBytes[eStart:eEnd] {
		issue.CtxAfter += string(line) + "\n"
	}
}

// TODO(waigani) get diff review working again.
func isInScope(scope controller.ReviewScope, issue *codelingo.Issue) bool {

	// no diffs, review is not scoped
	if len(scope.DiffFiles) == 0 {
		return true
	}
	for _, file := range scope.DiffFiles {
		start := issue.Position.Start
		end := issue.Position.End
		if file.NewName == start.Filename {
			for _, hunk := range file.Hunks {
				hunkStart := hunk.NewRange.Start
				if int(start.Line) >= hunkStart && int(end.Line) <= hunk.NewRange.Length+hunkStart {
					return true
				}
			}
		}
	}
	return false
}
