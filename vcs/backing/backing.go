package backing

type VCSBacking int

type Repo interface {
	Sync() error
	CurrentCommitId() (string, error)
	Patches() ([]string, error)
	// TODO(waigani) owner + name should be part of Repo struct.
	SetRemote(owner, name string) (string, string, error)
	CreateRemote(name string) error
	Exists(name string) (bool, error)
	OwnerAndNameFromRemote() (string, string, error)
	AssertNotTracked() error
	WorkingDir() (string, error)
}

const (
	NotAuthedErr VCSError   = "not logged into CodeLingo"
	Git          VCSBacking = iota
)

type VCSError string

func (v VCSError) Error() string {
	return string(v)
}
