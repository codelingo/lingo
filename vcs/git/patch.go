package git

import (
	"strings"

	"github.com/juju/errors"
)

// patch -p0 < diff.patch

// This command is exactly the same as the one for generating the patch,
// save for the --numstat option which returns statistics about each
// file in the diff instead of the diff, including whether or the the
// file is binary or not. The file is binary if and only if its line in
// the statistics contains the substring "\t-\t".
func isFileBinaryGit(filename string) (bool, error) {
	repoRoot, err := repoRoot()
	if err != nil {
		return false, errors.Trace(err)
	}

	out, err := gitCMD("-C", repoRoot, "diff", "--no-index", "--numstat", "/dev/null", filename)
	if err != nil {
	    return false, errors.Trace(err)
	}
	return strings.Contains(out, "\t-\t"), nil
}

// Patch returns a diff of any uncommited changes (stagged and unstaged).
func (r *Repo) Patches() ([]string, error) {
	var patches []string

	diffPatch, err := stagedAndUnstagedPatch()
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Don't add a patch for empty diffs
	if diffPatch != "" {
		patches = append(patches, diffPatch)
	}

	files, err := newFiles()
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, file := range files {
		isBinary, err := isFileBinaryGit(file)
		if err != nil {
			return nil, errors.Trace(err)
		}

		if !isBinary {
			filePatch, err := newFilePatch(file)
			if err != nil {
				return nil, errors.Trace(err)
			}
			if err := checkPatch(filePatch); err != nil {
				return nil, errors.Trace(err)
			}
			patches = append(patches, filePatch)
		}
	}

	return patches, nil
}

// checkPatch ensures the patch can be applied cleanly.
func checkPatch(patch string) error {
	// TODO(waigani) Implement this. Currently buggy.
	return nil
	out, err := gitCMD("apply", "--index", "--check", "-", patch)
	if err != nil {
		return errors.Annotate(err, out)
	}
	if out != "" {
		return errors.New(out)
	}
	return nil
}

func newFiles() ([]string, error) {
	repoRoot, err := repoRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}

	out, err := gitCMD("-C", repoRoot, "ls-files", "--others", "--exclude-standard")
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
	repoRoot, err := repoRoot()
	if err != nil {
		return "", errors.Trace(err)
	}
	// TODO(waigani) handle errors.
	out, err := gitCMD("-C", repoRoot, "diff", "--no-index", "/dev/null", filename)
	// TODO: why does this command give the correct output with a failing exit code?
	// Possibly occurs when the diff adds a new file.
	if strings.Contains(err.Error(), "exit status 1") {
		return out, nil
	}
	return out, errors.Trace(err)
}

func repoRoot() (string, error) {
	repoRoot, err := gitCMD("rev-parse", "--show-toplevel")
	if err != nil {
		return "", errors.Trace(err)
	}
	return strings.TrimSuffix(repoRoot, "\n"), nil
}
