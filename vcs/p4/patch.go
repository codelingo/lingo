package p4

import (
	"strings"

	"github.com/juju/errors"
)

// Todo(Junyu) add patch function for added and removed files

// Patch returns a diff of any uncommited changes (stagged and unstaged).
func (r *Repo) Patches() ([]string, error) {
	var patches []string
	diffPatch, err := stagedAndUnstagedPatch()
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Don't add a patch for empty diffs
	if !strings.Contains(diffPatch, "File(s) not opened") {
		patches = append(patches, diffPatch)
	}

	return patches, nil
}

func stagedAndUnstagedPatch() (string, error) {
	out, err := p4CMD("diff", "-du")
	if err != nil {
		return "", errors.Trace(err)
	}
	return out, nil
}
