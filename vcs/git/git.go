package git

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gogits/go-gogs-client"

	"github.com/juju/errors"

	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/app/util/common"
	"github.com/codelingo/lingo/app/util/common/config"
)

// TODO(waigani) pass in owner/name here and set them on Repo.
func New() *Repo {
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

	// Attempt to remove existing remote.
	out, err := gitCMD("remote", "remove", remoteName)
	if err != nil && !isKnownRemoteErr(err.Error()) {
		return "", "", errors.Annotate(err, out)
	}

	remoteAddr := fmt.Sprintf("%s/%s/%s.git", addr, repoOwner, repoName)
	out, err = gitCMD("remote", "add", remoteName, remoteAddr)
	if err != nil {
		return "", "", errors.Annotate(err, out)
	}
	return remoteName, remoteAddr, nil
}

func isKnownRemoteErr(errStr string) bool {

	for _, subStr := range []string{
		"No such remote",
		"Could not remove config section",
	} {
		if strings.Contains(errStr, subStr) {
			return true
		}

	}
	return false
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

	authCfg, err := config.Auth()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// TODO(waigani) change "password" to "token"
	token, err := authCfg.GetGitUserPassword()
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
		return "", "", errors.New("could not find repository owner, have you run `lingo config setup`?")
	}
	if len(m) < 3 || m[2] == "" {
		return "", "", errors.New("could not find repository name, have you run `lingo config setup`?")
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

	repoOpts := gogs.CreateRepoOption{
		Name:     name,
		Private:  true,
		AutoInit: false,
		//	Readme:   "Default",
	}

	if _, err = gogsClient.CreateRepo(repoOpts); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "violates unique constraint") {
			return errors.Wrap(err, util.RepoExistsError("failed to create repo, repo already exists"))
		}

		// TODO(waigani) TECHDEBT correct fix is to remove the go-gogs-client
		// client and replace it with gogsclient in
		// bots/clair/resource/gogsclient.go.
		if errMsg == "unexpected end of JSON input" {
			return errors.Wrap(err, util.UnauthorisedRepoError("failed to create repo, user unauthorised"))
		}

		return errors.Trace(err)
	}

	return nil
}

func (r *Repo) Sync(repoOwner string, workingDir string) error {
	cfg, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}
	remote, err := cfg.GitRemoteName()
	if err != nil {
		return errors.Trace(err)
	}
	// sync local and remote before reviewing
	out, err := gitCMD("push", "-4", remote, "HEAD", "--force", "--no-verify")
	if err != nil {
		// For users running older versions of git, run without the -4 flag.
		if strings.Contains(out, "unknown switch `4'") {
			_, err := gitCMD("push", remote, "HEAD", "--force", "--no-verify")
			return errors.Trace(err)
		}
	}
	return errors.Trace(err)
}

func (r *Repo) CurrentCommitId() (string, error) {
	out, err := gitCMD("rev-parse", "HEAD")
	if err != nil {
		return "", errors.Trace(err)
	}

	return strings.TrimSpace(out), nil
}

// WorkingDir returns a string representing the user's current directory in the format of the
// it will be represented in the store plus a trailing "/"
func (r *Repo) WorkingDir() (string, error) {
	dir, err := gitCMD("rev-parse", "--show-prefix")
	if err != nil {
		return "", errors.Trace(err)
	}

	dir = strings.Replace(dir, "\n", "", -1)

	return dir, nil
}

func (r *Repo) ReadFile(filename string) (string, error) {
	// If we are dealing with unstaged changes or the diff from a pull request,
	// just read from the current state of the repo.
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(out), nil
}

func (r *Repo) Clone(path, url string) error {
	out, err := gitCMD("-C", path, "clone", url)
	if err != nil {
		// TODO(waigani) better error handling
		errMsg := err.Error() + " " + string(out)

		// There is a race condition where the same repo may be cloned at
		// the same time.
		if !strings.Contains(errMsg, "already exists") {
			return errors.Annotate(err, "error cloning repo '"+url+"': "+errMsg)
		}
	}
	return nil
}

// Applies a raw diff to the current repo
// TODO: pass diff to stdin without whitespace unrecognised input error
func (r *Repo) ApplyPatch(diff string) error {
	// Create a new patch file containing the diff
	fname := "../temp.patch"
	f, err := os.Create(fname)
	if err != nil {
		return errors.Trace(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = fmt.Fprint(w, diff)
	if err != nil {
		return errors.Trace(err)
	}
	w.Flush()

	// Apply the patch
	_, err = gitCMD("apply", fname)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(os.Remove(fname))
}

func (r *Repo) CheckoutRemote(sha string) error {
	var err error
	var currentSha string
	if currentSha, err = r.CurrentCommitId(); err != nil {
		return errors.Trace(err)
	}
	if currentSha == sha {
		return nil
	}

	// fetch origin
	out, err := gitCMD("fetch", "-4", "origin")
	if err != nil {
		// For users running older versions of git, run without the -4 flag.
		if strings.Contains(out, "unknown switch `4'") {
			_, err := gitCMD("fetch", "origin")
			if err != nil {
				return errors.Trace(err)
			}
		} else {
			return errors.Trace(err)
		}
	}

	// checkout sha
	if _, err = gitCMD("checkout", sha); err != nil {
		return errors.Trace(err)
	}

	// delete master
	if _, err = gitCMD("branch", "-D", "master"); err != nil {
		return errors.Trace(err)
	}

	// checkout new master
	if _, err = gitCMD("checkout", "-b", "master"); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ClearChanges ensures there are no unstaged changes
func (r *Repo) ClearChanges() error {
	// repo already checked out, fetch latest.
	if _, err := gitCMD("clean", "-f"); err != nil {
		return errors.Trace(err)
	}

	if _, err := gitCMD("checkout", "."); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// TODO(benjamin-rood) Check git version to ensure expected cmd and behaviour
// by any git command-line actions
// Error codes:
// 129 incorrect arguments supplied to git
func gitCMD(args ...string) (out string, err error) {
	cmd := exec.Command("git", args...)
	b, err := cmd.CombinedOutput()
	out = string(b)
	return out, errors.Annotate(err, out)
}

func gitCmdInDir(dir string, args ...string) (out string, err error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	b, err := cmd.CombinedOutput()
	out = string(b)
	return out, errors.Annotate(err, out)
}

func (r *Repo) GetDotlingoFilepathsInDir(dir string) ([]string, error) {
	staged, err := gitCmdInDir(dir, "ls-tree", "-r", "--name-only", "HEAD")
	if err != nil {
		return nil, errors.Trace(err)
	}

	unstaged, err := gitCmdInDir(dir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, errors.Trace(err)
	}

	files := strings.Split(staged, "\n")
	files = append(files, strings.Split(unstaged, "\n")...)

	dotlingoFilepaths := []string{}
	for _, filepath := range files {
		if common.IsDotlingoFile(filepath) {
			dotlingoFilepaths = append(dotlingoFilepaths, filepath)
		}
	}

	return dotlingoFilepaths, nil
}
