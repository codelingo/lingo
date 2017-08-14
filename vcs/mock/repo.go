package mock

import (
	"errors"
	"github.com/codelingo/lingo/vcs"
)

// Repo mocking for unit testing.
// Intended to mock behaviour of the git.Repo implementation.
type Repo struct {
	vcs.Repo
}

// Minimal methods which implement backing.Repo interface.
// All methods reurn the default "zero" values except where fleshed out.

func (mockrepo *Repo) Sync(repoOwner string, workingDir string) error {
	return nil
}

func (mockrepo *Repo) Clone(path, url string) error {
	return nil
}

func (mockrepo *Repo) ApplyPatch(diff string) error {
	return nil
}

func (mockrepo *Repo) ClearChanges() error {
	return nil
}

func (mockrepo *Repo) CheckoutRemote(name string) error {
	return nil
}

func (mockrepo *Repo) ReadFile(filename string) (string, error) {
	return "", nil
}

func (mockrepo *Repo) CurrentCommitId() (string, error) {
	return "", nil
}

func (mockrepo *Repo) Patches() ([]string, error) {
	return nil, nil
}

func (mockrepo *Repo) SetRemote(owner, name string) (string, string, error) {
	return "", "", nil
}
func (mockrepo *Repo) CreateRemote(name string) error {
	switch name {
	case "existingPkg":
		return errors.New("already exists")
	case "existingPkg-1105":
		return errors.New("already exists")
	case "existing-Pkg":
		return errors.New("already exists")
	case "existing-Pkg-0":
		return errors.New("already exists")
	}

	return nil
}

func (mockrepo *Repo) Exists(name string) (bool, error) {
	return false, nil
}

func (mockrepo *Repo) OwnerAndNameFromRemote() (string, string, error) {
	return "", "", nil
}

func (mockrepo *Repo) AssertNotTracked() error {
	return nil
}

func (mockrepo *Repo) WorkingDir() (string, error) {
	return "", nil
}
