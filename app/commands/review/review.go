package review

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"

	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

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
		owner, repoName, err := repoOwnerAndNameFromRemote()
		if err != nil {
			return nil, errors.Annotate(err, "\nlocal vcs error")
		}

		// TODO(waigani) pass this in as opt
		repo := vcs.New(backing.Git)

		sha, err := repo.CurrentCommitId()
		if err != nil {
			return nil, errors.Trace(err)
		}

		if err := repo.Sync(); err != nil {
			return nil, err
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

	issues, err := svc.Review(reviewReq)
	if err != nil {
		if strings.Contains(err.Error(), "ambiguous argument 'HEAD'") {
			return nil, errors.New("\nThis looks like a new repository. Please make an initial commit and push to codelingo remote before reviewing. This is only required for the initial commit.")
		}

		return nil, errors.Annotate(err, "\nbad request")
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

	pCfg, err := config.Platform()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	remoteName, err := pCfg.GitRemoteName()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	cmd := exec.Command("git", "remote", "show", "-n", remoteName)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	r := regexp.MustCompile(`.*[\/:](.*)\/(.*)\.git`)
	m := r.FindStringSubmatch(string(b))
	if len(m) < 2 || m[1] == "" {
		return "", "", errors.Errorf("could not find repository owner, have you set %s as a git remote?", remoteName)
	}
	if len(m) < 3 || m[2] == "" {
		return "", "", errors.Errorf("could not find repository name, have you set %s as a git remote?", remoteName)
	}
	return m[1], m[2], nil

	// TODO(waigani) user may have added remote, but not commited code. In
	// that case, "git remote show" will give the following output:
	//
	// 	fatal: ambiguous argument 'remote': unknown revision or path not in the working tree.
	// Use '--' to separate paths from revisions, like this:
	// 'git <command> [<revision>...] -- [<file>...]'
	//
	// In this case, we need to tell the user to make an initial commit and
	// push to the remote. The steps are:
	//
	// 1. Create remote repo on codelingo git server
	// 2. Add remote as git remote
	// 3. Commit code and push to remote: `git push codelingo_dev master`
	//

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
