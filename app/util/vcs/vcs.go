package vcs

import (
	"github.com/codelingo/lingo/app/util/vcs/backing"
	"github.com/codelingo/lingo/app/util/vcs/git"
)

func New(b backing.VCSBacking) backing.Repo {
	switch b {
	case backing.Git:
		return git.New()

	}
	return nil
}
