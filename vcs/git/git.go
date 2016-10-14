package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/codelingo/lingo/app/util"
	"github.com/gogits/go-gogs-client"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/codelingo/lingo/vcs/backing"
)

// TODO(waigani) pass in owner/name here and set them on Repo.
func New() backing.Repo {
	return &Repo{}
}

type Repo struct {
}

func (r *Repo) SetRemote(repoOwner, repoName string) (string, string, error) {
	cfg, err := config.Platform()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	remoteName, err := cfg.GitRemoteName()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	addr, err := cfg.GitServerAddr()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	remoteAddr := fmt.Sprintf("%s/%s/%s.git", addr, repoOwner, repoName)
	out, err := gitCMD("remote", "add", remoteName, remoteAddr)
	if err != nil {
		return "", "", errors.Annotate(err, out)
	}
	return remoteName, remoteAddr, nil
}

func gogsClientForCurrentUser() (*gogs.Client, error) {

	cfg, err := config.Platform()
	if err != nil {
		return nil, errors.Trace(err)
	}

	addr, err := cfg.GitServerAddr()
	if err != nil {
		return nil, errors.Trace(err)
	}

	authCfg, err := util.AuthConfig()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// TODO(waigani) change "password" to "token"
	token, err := authCfg.Get("gitserver.user.password")
	if err != nil {
		return nil, errors.Trace(err)
	}

	return gogs.NewClient(addr, token), nil
}

func (r *Repo) Exists(name string) (bool, error) {
	gogsClient, err := gogsClientForCurrentUser()
	if err != nil {
		return false, errors.Trace(err)
	}

	repos, err := gogsClient.ListMyRepos()
	if err != nil {
		return false, errors.Trace(err)
	}
	for _, repo := range repos {
		if repo.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repo) CreateRemote(name string) error {

	gogsClient, err := gogsClientForCurrentUser()
	if err != nil {
		return errors.Trace(err)
	}

	_, err = gogsClient.CreateRepo(gogs.CreateRepoOption{
		Name: name,
		// TODO(waigani) make all repos private
		Private:  false,
		AutoInit: false,
		//	Readme:   "Default",
	})
	return errors.Trace(err)
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
