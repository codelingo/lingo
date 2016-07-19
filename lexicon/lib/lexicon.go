package lib

import (
	"errors"

	"github.com/codelingo/lexicon/lib/graph"
)

type lexicon struct {
	facts map[string]*Fact
}

func NewLexicon(facts ...*Fact) *lexicon {
	factMap := map[string]*Fact{}
	for _, f := range facts {
		factMap[f.Name()] = f
	}

	return &lexicon{
		facts: factMap,
	}
}

func (l *lexicon) Deduce(factName string) (*graph.Node, error) {
	if f, ok := l.facts[factName]; ok {
		return f.Deduce()
	}
	// TODO(waigani) used typed error.
	return nil, errors.New("fact not found")
}
func (l *lexicon) DeduceExists(factName, attribName string) (*graph.Node, error) {
	if f, ok := l.facts[factName]; ok {
		return f.DeduceExists(attribName)
	}
	// TODO(waigani) used typed error.
	return nil, errors.New("fact not found")
}
func (l *lexicon) DeduceString(factName, attribName, val string) (*graph.Node, error) {
	if f, ok := l.facts[factName]; ok {
		return f.DeduceString(attribName, val)
	}
	// TODO(waigani) used typed error.
	return nil, errors.New("fact not found")
}
func (l *lexicon) DeduceRegex(factName, attribName, reg string) (*graph.Node, error) {
	if f, ok := l.facts[factName]; ok {
		return f.DeduceRegex(attribName, reg)
	}
	// TODO(waigani) used typed error.
	return nil, errors.New("fact not found")
}
func (l *lexicon) DeduceComp(factName, attribName, op string, n int) (*graph.Node, error) {
	if f, ok := l.facts[factName]; ok {
		return f.DeduceComp(attribName, op, n)
	}
	// TODO(waigani) used typed error.
	return nil, errors.New("fact not found")
}

func (l *lexicon) Fact(name string) (*Fact, error) {
	if fct, ok := l.facts[name]; ok {
		return fct, nil
	}

	// TODO(waigani) make a typed error
	return nil, errors.New("fact not found")
}
