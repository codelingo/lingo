package server

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
	// name of the git user or org that owns the repo
	Owner string `json:"repoowner`
	// name of the repo
	Repo string `json:"repo`
	// sha of the commit to review
	SHA string `json:"sha"`
	// list of files and directories to limit the review to
	FilesAndDirs []string `json:"fileordirs"`

	// if true, all decendant files under dirs in FilesAndDirs will be reviewed.
	Recursive bool

	// a diff patch to apply to the remote branch before reviewing
	Patch string

	// TODO(waigani) add VCS field here
}

// ReviewResponse is the business domain type for a Review method response.
// type ReviewResponse struct {
// 	Issues []*codelingo.Issue `json:"issues"`
// }
