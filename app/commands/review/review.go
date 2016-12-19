package review

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"

	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

const noCommitErrMsg = "This looks like a new repository. Please make an initial commit before running `lingo review`. This is only required for the initial commit, subsequent changes to your repo will be picked up by lingo without committing."

// TODO(waigani) this function is used by other services, such as CLAIR.
// Refactor to have a services package, which both the lingo tool and services
// such as CLAIR use.
func Review(opts Options) ([]*codelingo.Issue, error) {
	// build the review request either from a pull request URL or the current repository
	var reviewReq *server.ReviewRequest

	// Build request from pull-request url
	if opts.PullRequest != "" {
		// TODO(waigani) support other hosts, e.g. bitbucket
		prOpts, err := parseGithubPR(opts.PullRequest)
		if err != nil {
			return nil, errors.Trace(err)
		}

		reviewReq = &server.ReviewRequest{
			Host:          prOpts.Host,
			Owner:         prOpts.RepoOwner,
			Repo:          prOpts.RepoName,
			IsPullRequest: true,
			PullRequestID: prOpts.PRID,
			// TODO(waigani) make this a CLI flag
			Recursive: false,
		}
		// Otherwise, build review request from current repository
	} else {

		// TODO(waigani) pass this in as opt
		repo := vcs.New(backing.Git)
		owner, repoName, err := repo.OwnerAndNameFromRemote()
		if err != nil {
			return nil, errors.Annotate(err, "\nlocal vcs error")
		}

		sha, err := repo.CurrentCommitId()
		if err != nil {
			if noCommitErr(err) {
				return nil, errors.New(noCommitErrMsg)
			}
			return nil, errors.Trace(err)
		}

		if err := repo.Sync(); err != nil {
			return nil, errors.Trace(err)
		}

		patches, err := repo.Patches()
		if err != nil {
			return nil, errors.Trace(err)
		}

		reviewReq = &server.ReviewRequest{
			Host:         "local",
			Owner:        owner,
			Repo:         repoName,
			FilesAndDirs: opts.FilesAndDirs,
			SHA:          sha,
			Patches:      patches,
			// TODO(waigani) make this a CLI flag
			Recursive: true,
		}
	}

	reviewReq.Dotlingo = opts.DotLingo

	svc, err := service.New()
	if err != nil {
		return nil, errors.Trace(err)
	}
	issuesc, messagec, progressc, err := svc.Review(nil, reviewReq)
	if err != nil {
		if noCommitErr(err) {
			return nil, errors.New(noCommitErrMsg)
		}

		return nil, errors.Annotate(err, "\nbad request")
	}
	go func() {
		for message := range messagec {
			//  Currently, the message chan just prints while we're waiting
			//  for issues. So we don't worry about closing it or timeouts
			//  etc.
			if message != "" {
				fmt.Println(string(message))
			}
		}
	}()

	showIngestProgress(progressc)

	// TODO(waigani) these should both be chans - as per first MVP.
	var confirmedIssues []*codelingo.Issue
	output := opts.SaveToFile == ""
	cfm, err := NewConfirmer(output, opts.KeepAll, nil)
	if err != nil {
		panic(err.Error())
		return nil, nil
	}

	// If user is manually confirming reviews, set a long timeout.
	timeout := time.After(time.Hour * 1)
	if opts.KeepAll {
		timeout = time.After(time.Second * 30)
	}

l:
	for {
		select {
		case issue, ok := <-issuesc:
			if !ok {
				break l
			}
			if cfm.Confirm(0, issue) {
				confirmedIssues = append(confirmedIssues, issue)
			}
		case <-timeout:
			return nil, errors.New("timed out waiting for issue")
		}
	}
	return confirmedIssues, nil
}

func showIngestProgress(progressc server.Ingestc) error {
	timeout := time.After(time.Second * 30)
	var ingestTotal int
	var ingestComplete bool

	// ingestSteps is how far along the ingest process we are
	var ingestSteps int
	var err error
	select {
	case progress, ok := <-progressc:
		if !ok {
			ingestComplete = true
			break
		}

		parts := strings.Split(progress, "/")
		if len(parts) != 2 {
			return errors.Errorf("ingest progress has wrong format. Expected n/n got %q", progress)
		}
		ingestSteps, err = strconv.Atoi(parts[0])
		if err != nil {
			return errors.Trace(err)
		}
		ingestTotal, err = strconv.Atoi(parts[1])
		if err != nil {
			return errors.Trace(err)
		}
	case <-timeout:
		return errors.New("timed out waiting for ingest to start")
	}

	if !ingestComplete {
		// ingest is not complete
		if ingestSteps < ingestTotal {
			return ingestBar(ingestSteps, ingestTotal, progressc)
		}
	}
	return nil
}

func ingestBar(current, total int, progressc server.Ingestc) error {
	ingestProgress := pb.StartNew(total)
	var finished bool

	// fast forward to where the ingest is up to.
	for i := 0; i < current; i++ {
		if ingestProgress.Increment() == total {
			ingestProgress.Finish()
			finished = true
			break
		}
	}

	if !finished {
		for {
			timeout := time.After(time.Second * 600)
			select {
			case _, ok := <-progressc:
				if ingestProgress.Increment() == total || !ok {
					ingestProgress.Finish()
					break
				}
			case <-timeout:
				return errors.New("timed out waiting for progress")
			}
		}
	}
	return nil
}

// TODO(waigani) use typed error
func noCommitErr(err error) bool {
	return strings.Contains(err.Error(), "ambiguous argument 'HEAD'")
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
	// TODO(waigani) validate PullRequest
	PullRequest  string
	FilesAndDirs []string
	Diff         bool   // ctx.Bool("diff") TODO(waigani) this should be a sub-command which proxies to git diff
	SaveToFile   string // ctx.String("save")
	KeepAll      bool   // ctx.Bool("keep-all")
	DotLingo     string // ctx.Bool("lingo-file")
	// TODO(waigani) add KeepAllWithTag. Use this for CLAIR autoreviews
	// TODO(waigani) add streaming json output
}
