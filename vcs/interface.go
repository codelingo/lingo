package vcs

type Type int

type Repo interface {
	Sync(repoOwner string, workingDir string) error
	CurrentCommitId() (string, error)
	Patches() ([]string, error)
	// TODO(waigani) owner + name should be part of Repo struct.
	SetRemote(owner, name string) (string, string, error)
	CreateRemote(name string) error
	Exists(name string) (bool, error)
	OwnerAndNameFromRemote() (string, string, error)
	AssertNotTracked() error
	WorkingDir() (string, error)
	ReadFile(filename string) (string, error)
	Clone(path, url string) error
	ApplyPatch(diff string) error
	ClearChanges() error
	CheckoutRemote(sha string) error
}

const (
	NotAuthedErr Error   = "not logged into CodeLingo"
	Git          Type = iota
	P4
)

type Error string

func (v Error) Error() string {
	return string(v)
}

