package review

import (
	"fmt"

	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/codelingo/lingo/bot/clair"
	"github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/codelingo/lingo/service/server"
	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/backing"

	"github.com/briandowns/spinner"
	"github.com/codelingo/lingo/service"
	"github.com/juju/errors"
)

const noCommitErrMsg = "This looks like a new repository. Please make an initial commit before running `lingo review`. This is only required for the initial commit, subsequent changes to your repo will be picked up by lingo without committing."

// TODO(waigani) this function is used by other services, such as CLAIR.
// Refactor to have a services package, which both the lingo tool and services
// such as CLAIR use.
func Review(opts Options) ([]*codelingo.Issue, error) {
	// build the review request either from a pull request URL or the current repository
	var dotlingos []string

	// TODO(waigani) pass this in as opt
	var err error
	repo := vcs.New(backing.Git)

	if err = syncRepo(repo); err != nil {
		return nil, errors.Trace(err)
	}

	// TODO: move the query building logic to CLAIR. CLAIR encapsulates all the logic relating
	// specifically to reviewing repositories.
	// You will be able to install CLAIR to use with the CLI with a command like `lingo clair review`,
	// and CLAIR will also be able to listen for GitHub pull requests.
	dotlingos, err = repo.BuildQueries()
	if err != nil {
		if noCommitErr(err) {
			return nil, errors.New(noCommitErrMsg)
		}

		return nil, errors.Annotate(err, "\nbad request")
	}

	// TODO: don't bother with query generation if there is a Dotlingo argument
	if opts.DotLingo != "" {
		dotlingos = []string{opts.DotLingo}
	}

	queryc := make(chan *codelingo.QueryRequest)
	go func() {
		for _, dl := range dotlingos {
			queryc <- &codelingo.QueryRequest{Dotlingo: dl}
		}
		close(queryc)
	}()

	resultc, err := service.Query(queryc)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var confirmedIssues []*codelingo.Issue
	spnr := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spnr.Start()

	output := opts.SaveToFile == ""
	cfm, err := NewConfirmer(output, opts.KeepAll, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// If user is manually confirming reviews, set a long timeout.
	timeout := time.After(time.Hour * 1)
	if opts.KeepAll {
		timeout = time.After(time.Second * 30)
	}

l:
	for {
		select {
		case result, ok := <-resultc:
			if !opts.KeepAll {
				spnr.Stop()
			}

			if !ok {
				break l
			}

			if result.Error != "" {
				return nil, errors.New(result.Error)
			}

			iss, err := clair.BuildIssue(result)
			if err != nil {
				return nil, errors.Trace(err)
			}

			if cfm.Confirm(0, iss) {
				confirmedIssues = append(confirmedIssues, iss)
			}

			if !opts.KeepAll {
				spnr.Restart()
			}
		case <-timeout:
			return nil, errors.New("timed out waiting for issue")
		}
	}

	// Stop spinner if it hasn't been stopped already
	if opts.KeepAll {
		spnr.Stop()
	}
	return confirmedIssues, nil
}

// sync the local repository with the remote, creating the remote if it does
// not exist.
func syncRepo(repo backing.Repo) error {

	if syncErr := repo.Sync(); syncErr != nil {

		// This case is triggered when a local remote has been added but
		// the remote repo does not exist. In this case, we create the
		// remote and try to sync again.
		missingRemote, err := regexp.MatchString("fatal: repository '.*' not found.*", syncErr.Error())
		if err != nil {
			return errors.Trace(err)
		}
		if missingRemote {
			_, repoName, err := repo.OwnerAndNameFromRemote()
			if err != nil {
				return errors.Trace(err)
			}

			// TODO(waigani) use typed errors
			if err := repo.CreateRemote(repoName); err != nil && !strings.HasPrefix(err.Error(), "repository already exists") {
				return errors.Trace(err)
			}
			if err := repo.Sync(); err != nil {
				return errors.Trace(err)
			}
		}

		return errors.Trace(syncErr)
	}
	return nil
}

func showIngestProgress(progressc server.Ingestc, messagec server.Messagec) error {
	timeout := time.After(time.Second * 30)
	var ingestTotal int
	var ingestComplete bool
	isIngesting := false

	// ingestSteps is how far along the ingest process we are
	var ingestSteps int
	var err error

	select {
	case message := <-messagec:
		msgStr := string(message)

		// TODO(junyu) Currently, we are receiving queue messages from
		// message channel to inform user about the queue and let them
		// wait in line. Ideally, we should allow parallel tree ingestion
		// so that the process can start immediately.
		if msgStr == "Queue not empty" {
			return errors.New("Queued for Ingest, try again later.")
		}

		// TODO(waigani) Currently, messagec is a mix of info and errors.
		// Either create a separate errors channel or use log level constants.
		if strings.HasPrefix(msgStr, "error") {
			return errors.New(msgStr)
		}

		if msgStr != "" {
			fmt.Println(msgStr)
		}
	case progress, ok := <-progressc:
		if !ok {
			ingestComplete = true
			break
		}

		if !isIngesting {
			fmt.Println("Ingesting...")
			isIngesting = true
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
	l:
		for {
			timeout := time.After(time.Second * 600)
			select {
			case _, ok := <-progressc:
				if ingestProgress.Increment() == total || !ok {
					ingestProgress.Finish()
					break l
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
