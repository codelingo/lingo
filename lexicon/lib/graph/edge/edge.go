package edge

import (
	"github.com/juju/errors"
)

type Dimension int

// TODO(waigani) Dimensions should be dynamically generated from lexicons.
const (
	DirDim Dimension = iota
	FileDim
	GitDim
	CommonASTDim
	GoASTDim
	PythonASTDim
)

func (d Dimension) String() string {
	switch d {
	case DirDim:
		return "directory-edge"
	case FileDim:
		return "file-edge"
	case GitDim:
		return "git-edge"
	case CommonASTDim:
		return "pingo-ast-edge"
	case PythonASTDim:
		return "python-ast-edge"
	case GoASTDim:
		return "go-ast-edge"
	}
	return "unknown"
}

// CLQL is a tree multigraph, with one tree graph per edge type.
// Edge represents a set of relation data to compatiable edges of other nodes.
type Edge interface {

	// the dimension the edge belongs to.
	Dimension() Dimension

	// Get the properties of the edge.
	Properties() interface{}

	// Set the properties of the edge
	SetProperties(interface{}) Edge
}

// The final output context is determined by the OutputTarget. For example, if
// we're going to attach a comment to a file, the OutputTarget will ask for
// the FileEdge and extract filename and line number.

type edge struct {
	parentId   int
	dimension  Dimension
	properties interface{}
}

func New(d Dimension) Edge {
	return &edge{dimension: d}
}

func (e *edge) Dimension() Dimension {
	return e.dimension
}

func (e *edge) Properties() interface{} {
	return e.properties
}

func (e *edge) SetProperties(props interface{}) Edge {
	if err := e.validatePropsType(props); err != nil {
		// yes panic, this is a developer error.
		panic(err)
	}

	e.properties = props
	return e
}

// TODO(waigani) candidate for a tenet here. when new edge is added, a few
// things need to be hooked up.
func (e *edge) validatePropsType(props interface{}) error {

	var matchType bool
	propDim := "unknown"
	switch props.(type) {
	case *DirProps:
		matchType = e.dimension == DirDim
		propDim = DirDim.String()
	case *FileProps:
		matchType = e.dimension == FileDim
		propDim = FileDim.String()
	case *GitProps:
		matchType = e.dimension == GitDim
		propDim = GitDim.String()
	case *CommonASTProps:
		matchType = e.dimension == CommonASTDim
		propDim = CommonASTDim.String()
	case *PythonASTProps:
		matchType = e.dimension == PythonASTDim
		propDim = PythonASTDim.String()
	case *GoASTProps:
		matchType = e.dimension == GoASTDim
		propDim = GoASTDim.String()
	default:
		return errors.Errorf("unknown properties of type %T: %#v", props, props)
	}

	if !matchType {
		return errors.Errorf("trying to set %s properties on a %s edge", propDim, e.dimension)
	}

	return nil
}
