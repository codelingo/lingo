package codemod

import (
	"fmt"
	"io/ioutil"

	"github.com/juju/errors"
)

func Write(newSRCs []*SRCHunk) error {

	// TODO(waigani) use one open file handler per file to write all changes
	// and use a buffered writer: https://www.devdungeon.com/content/working-
	// files-go#write_buffered
	for _, newSRC := range newSRCs {
		filename := newSRC.Filename
		fileSRC, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.Trace(err)
		}

		fileSRC = append(fileSRC[0:newSRC.StartOffset], append([]byte(newSRC.SRC), fileSRC[newSRC.EndOffset:]...)...)
		if err := ioutil.WriteFile(filename, []byte(fileSRC), 0644); err != nil {
			return errors.Trace(err)
		}
		fmt.Printf("Modified file %s\n", filename)
	}

	return nil
}
