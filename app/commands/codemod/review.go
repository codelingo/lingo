package codemod

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"time"

	"github.com/briandowns/spinner"
	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	grpcclient "github.com/codelingo/lingo/service/grpc"
	"github.com/codelingo/platform/flow/rpc/client"
	flow "github.com/codelingo/platform/flow/rpc/flowengine"

	"context"

	"github.com/juju/errors"
)

func RequestReview(ctx context.Context, req *flow.ReviewRequest) (chan *flow.Issue, chan error, error) {

	// // mock request result while testing
	// issc := make(chan *flow.Issue)
	// errc := make(chan error)
	// go func() {
	// 	defer close(issc)
	// 	defer close(errc)
	// 	issc <- &flow.Issue{
	// 		CtxBefore: "before",
	// 		CtxAfter:  "after",
	// 		LineText:  "line text",
	// 		Comment:   "this is comment",
	// 		Position: &flow.IssueRange{
	// 			Start: &flow.Position{
	// 				Filename: "main.go",
	// 				Offset:   10,
	// 				Line:     2,
	// 			},
	// 			End: &flow.Position{
	// 				Filename: "main.go",
	// 				Offset:   30,
	// 				Line:     3,
	// 			},
	// 		},
	// 	}
	// }()
	// return issc, errc, nil

	conn, err := service.GrpcConnection(service.LocalClient, service.FlowServer)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	c := client.NewFlowClient(conn)

	// Create context with metadata
	ctx, err = grpcclient.AddGcloudApiKeyToCtx(ctx)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	ctx, err = grpcclient.AddUsernameToCtx(ctx)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	issuec, errorc, err := c.Codemod(ctx, req)
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

func ConfirmIssues(cancel context.CancelFunc, issuec chan *flow.Issue, errorc chan error, keepAll bool, saveToFile string) ([]*SRCHunk, error) {
	defer util.Logger.Sync()

	var confirmedIssues []*SRCHunk
	spnr := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spnr.Start()
	defer spnr.Stop()

	output := saveToFile == ""
	cfm, err := NewConfirmer(output, keepAll, nil)
	if err != nil {
		cancel()
		return nil, errors.Trace(err)
	}

	// If user is manually confirming reviews, set a long timeout.
	timeout := time.After(time.Hour * 1)
	if keepAll {
		timeout = time.After(time.Minute * 5)
	}

l:
	for {
		select {
		case err, ok := <-errorc:
			if !ok {
				errorc = nil
				break
			}

			// Abort review
			cancel()
			util.Logger.Debugf("Review error: %s", errors.ErrorStack(err))
			return nil, errors.Trace(err)
		case iss, ok := <-issuec:
			if !keepAll {
				spnr.Stop()
			}
			if !ok {
				issuec = nil
				break
			}

			// TODO: remove errors from issues; there's a separate channel for that
			if iss.Err != "" {
				// Abort review
				cancel()
				return nil, errors.New(iss.Err)
			}

			// CONTINUE HERE

			// TODO(waigani) attach Tenet to issue
			// DEMOWARE hardcoding clql to one in root dir
			dLingoStr, err := ioutil.ReadFile(".lingo")
			if err != nil {
				return nil, errors.Trace(err)
			}
			type out struct {
				Tenets []struct {
					Bots  map[string]map[string]interface{}
					Query string
				}
			}

			var dLingo out
			err = yaml.Unmarshal(dLingoStr, &dLingo)
			if err != nil {
				return nil, errors.Trace(err)
			}

			clqlStr := "import codelingo/ast/go\n" + dLingo.Tenets[0].Bots["codelingo/codemod"]["set"].(string)

			// support one set decorator per query.
			srcs, err := ClqlToSrc(string(clqlStr))
			if err != nil {
				return nil, err
			}

			// TODO(waigani) align issue with set id. Until then, we can only support one set decorator
			src := srcs["_default"]

			// add full lines for user confirmation
			confirmSRC, err := fullLineSRC(iss, src)
			if err != nil {
				return nil, errors.Trace(err)
			}

			if cfm.Confirm(0, iss, confirmSRC) {

				// TODO(waigani) offsets need to be based off codemod decorator, not review.
				confirmedIssues = append(confirmedIssues, &SRCHunk{
					StartOffset: iss.Position.GetStart().Offset,
					EndOffset:   iss.Position.GetEnd().Offset,
					SRC:         src,
				})
			}

			if !keepAll {
				spnr.Restart()
			}
		case <-timeout:
			cancel()
			return nil, errors.New("timed out waiting for issue")
		}
		if issuec == nil && errorc == nil {
			break l
		}
	}

	// Stop spinner if it hasn't been stopped already
	if keepAll {
		spnr.Stop()
	}
	return confirmedIssues, nil
}

// returns the full lines of the SRC for the issue.
func fullLineSRC(issue *flow.Issue, newSRC string) (string, error) {

	pos := issue.GetPosition()
	startPos := pos.GetStart()
	endPos := pos.GetEnd()

	file, err := os.Open(startPos.Filename)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var byt int64
	var src []byte
	var strtLineByt int64
	var endLineByt int64
	var foundStart bool
	var foundEnd bool
	for scanner.Scan() {
		src = append(src, append(scanner.Bytes(), []byte("\n")...)...)

		startByt := byt
		endByt := startByt + int64(len(scanner.Bytes())+1) // +1 for \n char

		if startPos.Offset >= startByt && startPos.Offset <= endByt {

			// found start line
			strtLineByt = startByt
			foundStart = true
		}

		if endPos.Offset >= startByt && endPos.Offset <= endByt {

			// found end line
			endLineByt = endByt
			foundEnd = true
		}

		if foundStart && foundEnd {

			// xxx.Print(string(src))
			// fmt.Printf("\n[%d:%d]", strtLineByt, startPos.Offset)
			beforeNewSRC := string(src[strtLineByt:startPos.Offset])
			endNewSRC := string(src[endPos.Offset:endLineByt])

			return beforeNewSRC + newSRC + endNewSRC, nil

		}

		byt = endByt
	}

	return "", errors.Trace(scanner.Err())
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
