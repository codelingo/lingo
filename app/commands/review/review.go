package review

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/codelingo/flow/backend/service/client"
	"github.com/codelingo/flow/backend/service/flow"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/server"

	"github.com/juju/errors"
)

func RequestReview(req *flow.ReviewRequest) (chan *flow.Issue, chan error, error) {
	conn, err := service.GrpcConnection(service.LocalClient, service.FlowServer)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	c := client.NewFlowClient(conn)
	issuec, errorc, err := c.Review(context.Background(), req)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	return issuec, errorc, nil
}

func MakeReport(issues []*flow.Issue, format, outputFile string) (string, error) {
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = json.Marshal(issues)
		if err != nil {
			return "", errors.Trace(err)
		}
	case "json-pretty":
		data, err = json.MarshalIndent(issues, " ", " ")
		if err != nil {
			return "", errors.Trace(err)
		}
	default:
		return "", errors.Errorf("Unknown format %q", format)
	}

	if outputFile != "" {
		err = ioutil.WriteFile(outputFile, data, 0775)
		if err != nil {
			return "", errors.Annotate(err, "Error writing issues to file")
		}
		return fmt.Sprintf("Done! %d issues written to %s \n", len(issues), outputFile), nil
	}

	return string(data), nil
}

// Read a .lingo file from a filepath argument
func ReadDotLingo(ctx *cli.Context) (string, error) {
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

func ConfirmIssues(issuec chan *flow.Issue, errorc chan error, keepAll bool, saveToFile string) ([]*flow.Issue, error) {
	var confirmedIssues []*flow.Issue
	spnr := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spnr.Start()

	output := saveToFile == ""
	cfm, err := NewConfirmer(output, keepAll, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// If user is manually confirming reviews, set a long timeout.
	timeout := time.After(time.Hour * 1)
	if keepAll {
		timeout = time.After(time.Second * 30)
	}

l:
	for {
		select {
		case err, ok := <-errorc:
			if !ok {
				break
			}

			defer util.Logger.Sync()
			util.Logger.Debug("Review error: ", err.Error())
		case iss, ok := <-issuec:
			if !keepAll {
				spnr.Stop()
			}

			if !ok {
				break l
			}

			if iss.Err != "" {
				return nil, errors.New(iss.Err)
			}

			if cfm.Confirm(0, iss) {
				confirmedIssues = append(confirmedIssues, iss)
			}

			if !keepAll {
				spnr.Restart()
			}
		case <-timeout:
			return nil, errors.New("timed out waiting for issue")
		}
	}

	// Stop spinner if it hasn't been stopped already
	if keepAll {
		spnr.Stop()
	}
	return confirmedIssues, nil
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

func NewRange(filename string, startLine, endLine int) *flow.IssueRange {
	start := &flow.Position{
		Filename: filename,
		Line:     int64(startLine),
	}

	end := &flow.Position{
		Filename: filename,
		Line:     int64(endLine),
	}

	return &flow.IssueRange{
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
