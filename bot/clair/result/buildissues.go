// The result package is responsible for taking query results, transforming them into issuues,
// and routing them to their destination, whether that's the user, or Github etc.

package result

import (
	"bytes"
	"strconv"

	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/waigani/diffparser"

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

func getSingleValue(reply *codelingo.QueryReply, name string, index int) (string, error) {
	list, ok := reply.Data[name]
	if !ok {
		return "", errors.Errorf("Reply field %s does not exist.", name)
	}

	if len(list.Data) <= index {
		return "", errors.Errorf("Reply field %s has no value at index %d.", name, index)
	}

	return list.Data[0], nil
}

func buildIssue(n *codelingo.QueryReply) (*codelingo.Issue, error) {
	if n.Error != "" {
		return &codelingo.Issue{
			Err: n.Error,
		}, nil
	}

	intNames := []string{"start_offset", "start_line", "start_column", "end_offset", "end_line", "end_column"}
	values := []int{}
	for _, name := range intNames {
		val, err := getSingleValue(n, name, 0)
		if err != nil {
			return nil, errors.Trace(err)
		}

		intVal, err := strconv.Atoi(val)
		if err != nil {
			return nil, errors.Trace(err)
		}
		values = append(values, intVal)
	}

	// we let this panic if filename is not set as it's a developer error.
	filename, err := getSingleValue(n, "filename", 0)
	if err != nil {
		return nil, errors.Trace(err)
	}

	name, err := getSingleValue(n, "name", 0)
	if err != nil {
		return nil, errors.Trace(err)
	}

	comment, err := getSingleValue(n, "comment", 0)
	if err != nil {
		return nil, errors.Trace(err)
	}

	iRange := &codelingo.IssueRange{
		Start: &codelingo.Position{
			Filename: filename,
			Offset:   int64(values[0]),
			Line:     int64(values[1]),
			Column:   int64(values[2]),
		},
		End: &codelingo.Position{
			Filename: filename,
			Offset:   int64(values[3]),
			Line:     int64(values[4]),
			Column:   int64(values[5]),
		},
	}

	iss := &codelingo.Issue{
		Name:     name,
		Position: iRange,
		// TODO(waigani) make comment a slice
		Comment: comment,
		// Metrics
		NewCode: true,
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

// Checks if the start and end of a given issue fall within a new code block.
// TODO(waigani) get diff review working again.
func isInScope(diffFiles []*diffparser.DiffFile, issue *codelingo.Issue) bool {

	// no diffs, review is not scoped
	if len(diffFiles) == 0 {
		return true
	}
	for _, file := range diffFiles {
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
