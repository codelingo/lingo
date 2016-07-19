package edge

// TODO(waigani) empty nodes with ids can be created and stored in the db when
// the query is compiled. Then the Node is populated once the query is exicuted
// and the fact deduces the node exists.
// We'll need to a track unique query id as well.
// TODO(waigani) lookup parent from db with e.parentId
type props struct {
	// as CLQL is a tree graph, each edge must have one parent. This is the parent node's id.
	ParentId int
}

type DirProps struct {
	props
	Dirname string
}

// fileProps
type FileProps struct {
	props

	// the file the node was found in.
	Filename string

	StartLine, StartColumn, EndLine, EndColumn int
}

// TODO(waigani) implement
type GitProps struct {
	props
}

// TODO(waigani) implement
type CommonASTProps struct {
	props
}

// TODO(waigani) implement
type PythonASTProps struct {
	props
}

// TODO(waigani) implement
type GoASTProps struct {
	props
}
