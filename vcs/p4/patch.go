package p4

import (
	"strings"

	"github.com/juju/errors"
	"regexp"
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

	delFiles, err := deletedFiles()
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, file := range delFiles {
		patches = append(patches, file)
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

// deletedFile creates a slice of custom patch strings about the deleted files
// ie. delete //stream/main/hello.cpp
func deletedFiles() ([]string, error) {
	out, err := p4CMD("-Ztag", "-F", "%action% %depotFile%", "status")
	if err != nil {
		return nil, errors.Trace(err)
	}
	reg := regexp.MustCompile("(?m)^delete .+")
	matches := reg.FindAllString(out, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}
