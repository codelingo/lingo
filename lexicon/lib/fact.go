package lib

type Deducer func(parentID string, args ...interface{}) (nodeID string)

// Fact represents the fact that we're interested in. Normally, this will be
// the name of an AST node type, e.g. "func", "var". Though it can also be a
// more abstract type.
type Fact struct {
	name     string
	deducers map[string]Deducer
	nodeID   string
}

func NewFact(name string, deducers map[string]Deducer) *Fact {

	// set default deducers

	return &Fact{
		name:     name,
		deducers: deducers,
	}
}

func (f *Fact) Name() string {
	return f.name
}
