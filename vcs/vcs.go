package vcs

import (
	"github.com/codelingo/lingo/vcs/backing"
	"github.com/codelingo/lingo/vcs/git"
)

func New(b backing.VCSBacking) backing.Repo {
	switch b {
	case backing.Git:
		return git.New()

	}
	return nil
}
