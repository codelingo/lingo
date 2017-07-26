package p4

// TODO (Junyu) remove any unnecessary functions.
import (
	"strings"

	"github.com/juju/errors"
)

// patch -p0 < diff.patch

// Patch returns a diff of any uncommited changes (stagged and unstaged).
func (r *Repo) Patches() ([]string, error) {
	var patches []string
	return patches, nil
}

// checkPatch ensures the patch can be applied cleanly.
func checkPatch(patch string) error {
	// TODO(waigani) Implement this. Currently buggy.
	return nil
	out, err := p4CMD("apply", "--index", "--check", "-", patch)
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

	out, err := p4CMD("-C", repoRoot, "ls-files", "--others", "--exclude-standard")
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
	out, err := p4CMD("diff", "HEAD")
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
	out, err := p4CMD("-C", repoRoot, "diff", "--no-index", "/dev/null", filename)
	// TODO: why does this command give the correct output with a failing exit code?
	// Possibly occurs when the diff adds a new file.
	if strings.Contains(err.Error(), "exit status 1") {
		return out, nil
	}
	return out, errors.Trace(err)
}

func repoRoot() (string, error) {
	repoRoot, err := p4CMD("rev-parse", "--show-toplevel")
	if err != nil {
		return "", errors.Trace(err)
	}
	return strings.TrimSuffix(repoRoot, "\n"), nil
}
