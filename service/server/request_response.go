package server

type ListLexiconRequest struct{}

type ListLexiconResponse struct {
	Key string `json:"lexicons"`
}

type SessionRequest struct{}

type SessionResponse struct {
	Key string `json:"key"`
}

type PathsFromOffsetRequest struct {
	Lang     string
	Dir      string
	Filename string
	Src      string
	Start    int
	End      int
}

type PathsFromOffsetResponse struct {
	Paths [][]string
}

// QueryRequest is the business domain type for a Query method request.
type QueryRequest struct {
	CLQL string `json:"clql"`
}

// QueryResponse is the business domain type for a Query method response.
type QueryResponse struct {
	Result string `json:"result"`
}

// ReviewRequest is the business domain type for a Review method request.
type ReviewRequest struct {
	// Session key for coordinating pubsub queues and multiple rpc requests
	Key string `json:"key"`
	// The repository host. Examples: "local", "github_com"
	Host string
	// name of the git user or org that owns the repo
	Owner string `json:"repoowner`
	// name of the repo
	Repo string `json:"repo`
	// sha of the commit to review
	SHA string `json:"sha"`
	// list of files and directories to limit the review to
	// TODO(waigani) support byte-offsets with files:  main.go:2343,4321;3421,3432;
	FilesAndDirs []string `json:"fileordirs"`

	// if true, all decendant files under dirs in FilesAndDirs will be reviewed.
	Recursive bool

	// a diff patch to apply to the remote branch before reviewing
	Patches []string

	IsPullRequest bool

	PullRequestID int
	// TODO(waigani) add VCS field here
	Vcs string

	Dotlingo string
}

// ReviewResponse is the business domain type for a Review method response.
type ReviewResponse struct{}
