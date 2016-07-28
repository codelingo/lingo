package git

import (
	"os/exec"
	"strings"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/vcs/backing"
)

func New() backing.Repo {
	return &Repo{}
}

type Repo struct {
}

func (r *Repo) Sync() error {

	cfg, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}
	remote, err := cfg.GitRemoteName()
	if err != nil {
		return errors.Trace(err)
	}

	// sync local and remote before reviewing
	_, err = gitCMD("push", remote, "HEAD")
	return errors.Trace(err)
}

func (r *Repo) CurrentCommitId() (string, error) {
	out, err := gitCMD("rev-parse", "HEAD")
	if err != nil {
		return "", errors.Trace(err)
	}

	return out, nil
}

func gitCMD(args ...string) (out string, err error) {
	cmd := exec.Command("git", args...)
	b, err := cmd.CombinedOutput()
	out = strings.TrimSpace(string(b))
	return out, errors.Annotate(err, out)

	// TODO(waigani) stdout is empty?
	// cmd := exec.Command("git", args...)
	// e := &bytes.Buffer{}
	// o := &bytes.Buffer{}
	// cmd.Stderr = e
	// cmd.Stdout = o
	// stderr = string(e.Bytes())
	// stdout = string(o.Bytes())
	// err = cmd.Run()
	// if err != nil {
	// 	gitargs := strings.Join(args, " ")
	// 	return "", stderr, errors.Annotate(err, "git args: `"+gitargs+"` stdout: "+stdout+" stderr: "+stderr)

	// }
	// return strings.TrimSpace(stdout), stderr, nil
}
