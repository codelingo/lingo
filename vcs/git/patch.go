package git

import (
	"strings"

	"github.com/juju/errors"
)

// patch -p0 < diff.patch

// Patch returns a diff of any uncommited changes (stagged and unstaged).
func (r *Repo) Patches() ([]string, error) {
	var patches []string

	diffPatch, err := stagedAndUnstagedPatch()
	if err != nil {
		return nil, errors.Trace(err)
	}
	patches = append(patches, diffPatch)
	files, err := newFiles()
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, file := range files {
		filePatch, err := newFilePatch(file)
		if err != nil {
			return nil, errors.Trace(err)
		}
		patches = append(patches, filePatch)
	}

	return patches, nil
}

func newFiles() ([]string, error) {
	out, err := gitCMD("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, errors.Trace(err)
	}
	var files []string
	for _, file := range strings.Split(out, "\n") {
		if file != "" {
			files = append(files, file)
		}
	}
	return files, nil
}

func stagedAndUnstagedPatch() (string, error) {
	out, err := gitCMD("diff", "HEAD")
	if err != nil {
		return "", errors.Trace(err)
	}

	return out, nil
}

func newFilePatch(filename string) (string, error) {
	// TODO(waigani) handle errors.
	out, _ := gitCMD("diff", "--no-index", "/dev/null", filename)
	out = strings.TrimSuffix(out, "\n\\ No newline at end of file")
	return out, nil
}
