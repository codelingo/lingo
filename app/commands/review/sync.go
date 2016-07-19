package review

import (
	"bytes"
	"os/exec"

	"github.com/juju/errors"
)

// TODO(waigani) provide sync for github repos

// Sync HEAD of the current branch with remote.
func sync() error {

	// TODO(waigani) disabled for demo
	return nil

	cmd := exec.Command("git", "push", "codelingo", "HEAD")
	b := &bytes.Buffer{}
	cmd.Stderr = b
	if err := cmd.Run(); err != nil {
		// TODO(waigani) better error handling
		return errors.Annotate(err, string(b.Bytes()))
	}
	return cmd.Run()

}
