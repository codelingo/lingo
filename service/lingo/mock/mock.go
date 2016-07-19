package mock

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codelingo/lingo/service/lingo/preprocessor"
	"github.com/lingo-reviews/clql/query"

	"github.com/codelingo/lexicon/lib/graph/edge"

	"github.com/juju/errors"
)

type result struct {
	Comment, Filename, Before, After, Line string
	Start, End                             int
}

// TODO(waigani) in production Results will return a chan.
// https://github.com/codelingo/demo/issues/8
func Results() ([]*result, error) {

	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Trace(err)
	}
	lingoFilepath := filepath.Join(pwd, ".lingo")

	lingoBytes, err := ioutil.ReadFile(lingoFilepath)

	// TODO(waigani) Process should return a lingo struct
	m, err := preprocessor.Process(string(lingoBytes), 0)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var lexicons []string
	for _, l := range m["lexicons"].([]interface{}) {
		lexicons = append(lexicons, l.(string))
	}

	var allResults []*result
	// TODO(waigani) lots of unsafe key lookups and type assertions here.
	// These will be avoided once preprocessor.Process returns a lingo struct.
	for _, t := range m["tenets"].([]interface{}) {
		tenet := t.(map[interface{}]interface{})

		qry := tenet["match"].(*query.Query)

		// set lexicons and run the query
		nodesOfInterest, err := qry.SetLexicons(lexicons...).Run()
		if err != nil {
			return nil, errors.Trace(err)
		}

		var tenetResults = make([]*result, len(nodesOfInterest))
		for i, n := range nodesOfInterest {

			e := n.GetEdge(edge.FileDim)
			fileProps := e.Properties().(*edge.FileProps)

			b, err := src(fileProps.Filename)
			if err != nil {
				panic(err)
			}

			start := fileProps.StartLine
			end := fileProps.EndLine

			buf := 2
			sStart := start - buf
			if sStart < 1 {
				sStart = 1
			}
			sEnd := start - 1

			r := &result{}
			// TODO(waigani) demoware. Before will not show for start lines of 0,1 or 2
			// before
			for _, line := range b[sStart:sEnd] {
				r.Before += string(line) + "\n"
			}

			// line
			if start == end {
				r.Line = string(b[start])
			} else {
				for _, line := range b[start:end] {
					r.Line += string(line) + "\n"
				}
			}

			// after
			l := len(b)

			eStart := end + 1
			if eStart > l {
				eStart = l
			}

			eEnd := eStart + buf
			if eEnd > l {
				eEnd = l
			}

			for _, line := range b[eStart:eEnd] {
				r.After += string(line) + "\n"
			}

			if com, ok := tenet["comment"]; ok {
				switch c := com.(type) {
				case string:
					r.Comment = c
				case []interface{}:
					r.Comment = c[i].(string)
				}
			}
			tenetResults[i] = r
		}
		allResults = append(allResults, tenetResults...)
	}
	return allResults, nil
}

// read src
func src(filename string) ([][]byte, error) {
	var srcBytes []byte
	var err error
	// TODO(matt) TECHDEBT use "go/scanner".Scanner instead of loading all bytes to memory.
	// https://github.com/codelingo/demo/issues/9
	srcBytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// reset index to match file lines - easier to reason about.
	lineBytes := make([][]byte, len(srcBytes))
	for i, b := range bytes.Split(srcBytes, []byte("\n")) {
		lineBytes[i+1] = b
	}

	return lineBytes, nil
}
