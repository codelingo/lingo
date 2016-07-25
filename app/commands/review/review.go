package review

import (
	"os/exec"
	"regexp"

	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"

	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

func Review(opts Options) ([]*codelingo.Issue, error) {

	svc, err := service.New()
	if err != nil {
		return nil, errors.Trace(err)
	}

	owner, repo, err := repoOwnerAndNameFromRemote()
	if err != nil {
		return nil, errors.Annotate(err, "local vcs error")
	}

	issues, err := svc.Review(&server.ReviewRequest{
		Owner:        owner,
		Repo:         repo,
		FilesAndDirs: opts.Files,
	})
	if err != nil {
		return nil, errors.Annotate(err, "bad request")
	}

	var confirmedIssues []*codelingo.Issue
	for _, issue := range issues {
		output := opts.SaveToFile == ""
		cfm, err := NewConfirmer(output, opts.KeepAll, nil)
		if err != nil {
			panic(err.Error())
			return nil, nil
		}
		cfm.Confirm(0, issue)

		confirmedIssues = append(confirmedIssues, issue)
	}
	return confirmedIssues, nil
}

func NewRange(filename string, startLine, endLine int) *codelingo.IssueRange {
	start := &codelingo.Position{
		Filename: filename,
		Line:     int64(startLine),
	}

	end := &codelingo.Position{
		Filename: filename,
		Line:     int64(endLine),
	}

	return &codelingo.IssueRange{
		Start: start,
		End:   end,
	}
}

func repoOwnerAndNameFromRemote() (string, string, error) {
	cmd := exec.Command("git", "remote", "show", "-n", "codelingo")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	r := regexp.MustCompile(`.*[\/:](.*)\/(.*)\.git`)
	m := r.FindStringSubmatch(string(b))
	if m[1] == "" {
		err = errors.New("could not find repository owner from codelingo remote")
	}
	if m[2] == "" {
		err = errors.New("could not find repository name from codelingo remote")
	}
	return m[1], m[2], err
}

// TODO(waigani) simplify representation of Issue.
// https://github.com/codelingo/demo/issues/7
// type Issue struct {
// 	apiIssue
// 	TenetName     string `json:"tenetName,omitempty"`
// 	Discard       bool   `json:"discard,omitempty"`
// 	DiscardReason string `json:"discardReason,omitempty"`
// }

// type apiIssue struct {
// 	// The name of the issue.
// 	TenetName     string            `json:"tenetName,omitempty"`
// 	Discard       bool              `json:"discard,omitempty"`
// 	DiscardReason string            `json:"discardReason,omitempty"`
// 	Name          string            `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
// 	Position      *IssueRange       `protobuf:"bytes,2,opt,name=position" json:"position,omitempty"`
// 	Comment       string            `protobuf:"bytes,3,opt,name=comment" json:"comment,omitempty"`
// 	CtxBefore     string            `protobuf:"bytes,4,opt,name=ctxBefore" json:"ctxBefore,omitempty"`
// 	LineText      string            `protobuf:"bytes,5,opt,name=lineText" json:"lineText,omitempty"`
// 	CtxAfter      string            `protobuf:"bytes,6,opt,name=ctxAfter" json:"ctxAfter,omitempty"`
// 	Metrics       map[string]string `protobuf:"bytes,7,rep,name=metrics" json:"metrics,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
// 	Tags          []string          `protobuf:"bytes,8,rep,name=tags" json:"tags,omitempty"`
// 	Link          string            `protobuf:"bytes,9,opt,name=link" json:"link,omitempty"`
// 	NewCode       bool              `protobuf:"varint,10,opt,name=newCode" json:"newCode,omitempty"`
// 	Patch         string            `protobuf:"bytes,11,opt,name=patch" json:"patch,omitempty"`
// 	Err           string            `protobuf:"bytes,12,opt,name=err" json:"err,omitempty"`
// }

// type IssueRange struct {
// 	Start *Position `protobuf:"bytes,1,opt,name=start" json:"start,omitempty"`
// 	End   *Position `protobuf:"bytes,2,opt,name=end" json:"end,omitempty"`
// }

// type Position struct {
// 	Filename string `protobuf:"bytes,1,opt,name=filename" json:"filename,omitempty"`
// 	Offset   int64  `protobuf:"varint,2,opt,name=Offset" json:"Offset,omitempty"`
// 	Line     int64  `protobuf:"varint,3,opt,name=Line" json:"Line,omitempty"`
// 	Column   int64  `protobuf:"varint,4,opt,name=Column" json:"Column,omitempty"`
// }

type Options struct {
	Files      []string // ctx.Args()
	Diff       bool     // ctx.Bool("diff")
	SaveToFile string   // ctx.String("save")
	KeepAll    bool     // ctx.Bool("keep-all")
}
