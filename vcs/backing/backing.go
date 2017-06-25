package backing

type VCSBacking int

type Repo interface {
	BuildQueries(host, owner, name string) ([]string, error)
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
	ReadFile(filename string) (string, error)
	Clone(path, url string) error
	ApplyPatch(diff string) error
	ClearChanges() error
	CheckoutRemote(sha string) error
}

const (
	NotAuthedErr VCSError   = "not logged into CodeLingo"
	Git          VCSBacking = iota
)

type VCSError string

func (v VCSError) Error() string {
	return string(v)
}
