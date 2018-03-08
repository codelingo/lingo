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
	// TODO In .lingo we have a *statement* of *facts*.
	// With the store we have a *query* *path* that finds a *node*.
	// We need to keep the naming consistent and specific to the domain.
	// Using path here potentially bleeds the layers.
	// Key terms to use consciously in the right places:
	// fact, property, branch, leaf, kind, statement, query, path, node ...
	// We need to define Tenets to guide this.

	Paths []*Path
}

type Path struct {
	Facts []*GenFact
}

type GenFact struct {
	FactName   string
	Properties map[string]string
}

// QueryRequest is the business domain type for a Query method request.
type QueryRequest struct {
	Dotlingo string `json:"clql"`
}

// QueryResponse is the business domain type for a Query method response.
type QueryResponse struct {
	// Id of the result node found by the query
	ID string `json:"result"`
	// The list of kinds that
	Kind  []string
	Data  map[string][]string
	Error string
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

	Dir string
}

// ReviewResponse is the business domain type for a Review method response.
type ReviewResponse struct{}

type DescribeFactRequest struct {
	Owner   string
	Name    string
	Version string
	Fact    string
}

type DescribeFactResponse struct {
	Examples    string
	Description string
	Properties  []Property
}

type Property struct {
	Name        string
	Description string
}

type LatestClientVersionRequest struct{}

type LatestClientVersionResponse struct {
	Key string `json:"version"`
}
