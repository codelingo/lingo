package git

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/codelingo/lingo/app/util/vcs/backing"
	"github.com/waigani/xxx"
)

const (
	// remote   = "origin"
	remote = "codelingo"
)

func New() backing.Repo {
	return &Repo{}
}

type Repo struct {
}

func (r *Repo) Sync() error {
	// sync local and remote before reviewing
	cmd := exec.Command("git", "push", remote, "HEAD")
	b := &bytes.Buffer{}
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		xxx.Print("err hit")
		return err
	}
	xxx.Print(string(b.Bytes()))
	return nil
}

func (r *Repo) CurrentCommitId() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	b := &bytes.Buffer{}
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b.Bytes())), nil
}
