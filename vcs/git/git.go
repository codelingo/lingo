package git

import (
	"fmt"
	"os/exec"
	"regexp"
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

func (r *Repo) OwnerAndNameFromRemote() (string, string, error) {
	pCfg, err := config.Platform()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	remoteName, err := pCfg.GitRemoteName()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	cmd := exec.Command("git", "remote", "show", "-n", remoteName)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	result := regexp.MustCompile(`.*[\/:](.*)\/(.*)\.git`)
	m := result.FindStringSubmatch(string(b))
	if len(m) < 2 || m[1] == "" {
		return "", "", errors.New("could not find repository owner, have you run `lingo init`?")
	}
	if len(m) < 3 || m[2] == "" {
		return "", "", errors.New("could not find repository name, have you run `lingo init?`")
	}
	return m[1], m[2], nil

	// TODO(waigani) user may have added remote, but not commited code. In
	// that case, "git remote show" will give the following output:
	//
	// 	fatal: ambiguous argument 'remote': unknown revision or path not in the working tree.
	// Use '--' to separate paths from revisions, like this:
	// 'git <command> [<revision>...] -- [<file>...]'
	//
	// In this case, we need to tell the user to make an initial commit and
	// push to the remote. The steps are:
	//
	// 1. Create remote repo on codelingo git server
	// 2. Add remote as git remote
	// 3. Commit code and push to remote: `git push codelingo_dev master`
	//
}

// AssertNotTracked checks for the existence of the appropriate
// codelingo remote to avoid duplications on GOGS.
func (r *Repo) AssertNotTracked() error {
	platCfg, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}

	remote, err := platCfg.GitRemoteName()
	if err != nil {
		return errors.Trace(err)
	}

	out, err := gitCMD("remote", "show", "-n")
	if err != nil {
		return errors.Annotate(err, out)
	}

	parts := strings.Split(out, "\n")
	for _, p := range parts {
		if p == remote {
			return errors.Errorf("%s git remote already exists", r)
		}
	}
	return nil
}

func (r *Repo) CreateRemote(name string) error {

	gogsClient, err := gogsClientForCurrentUser()
	if err != nil {
		return errors.Trace(err)
	}

	_, err = gogsClient.CreateRepo(gogs.CreateRepoOption{
		Name:     name,
		Private:  true,
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
	_, err = gitCMD("push", remote, "HEAD", "--force", "--no-verify")
	return errors.Trace(err)
}

func (r *Repo) CurrentCommitId() (string, error) {
	out, err := gitCMD("rev-parse", "HEAD")
	if err != nil {
		return "", errors.Trace(err)
	}

	return out, nil
}

// TODO(benjamin-rood) Check git version to ensure expected cmd and behaviour
// by any git command-line actions

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
