package backing

type VCSBacking int

type Repo interface {
	Sync() error
	CurrentCommitId() (string, error)
	Patches() ([]string, error)
}

const (
	NotAuthedErr VCSError   = "not logged into CodeLingo"
	Git          VCSBacking = iota
)

type VCSError string

func (v VCSError) Error() string {
	return string(v)
}
