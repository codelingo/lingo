package p4

// TODO (Junyu) remove any unnecessary functions.
import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juju/errors"

	"regexp"

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
	remoteName, err := cfg.P4RemoteName()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	remoteAddr, err := cfg.P4ServerAddr()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	// remove existing remote setting
	out, err := p4CMD("remotes")
	if err != nil {
		return "", "", errors.Annotate(err, out)
	}
	if strings.Contains(out, remoteName) {
		out, err = p4CMD("remote", "-d", remoteName)
		if err != nil {
			return "", "", errors.Annotate(err, out)
		}
	}

	c1 := exec.Command("p4", "remote", "-o", remoteName)
	c2 := exec.Command("sed", "-e", "s/Address:\tlocalhost:1666/Address:\t"+remoteAddr+"/")
	c3 := exec.Command("p4", "remote", "-i")
	c2.Stdin, err = c1.StdoutPipe()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	c3.Stdin, err = c2.StdoutPipe()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	c3.Stdout = os.Stdout
	if err := c3.Start(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c2.Start(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c1.Run(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c2.Wait(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c3.Wait(); err != nil {
		return "", "", errors.Trace(err)
	}
	c1 = exec.Command("p4", "remote", "-o", remoteName)
	c2 = exec.Command("sed", "-e", "s@//... //...@//stream/main/... //depot/"+repoOwner+"/"+repoName+"/...@")
	c3 = exec.Command("p4", "remote", "-i")
	c2.Stdin, err = c1.StdoutPipe()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	c3.Stdin, err = c2.StdoutPipe()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	c3.Stdout = os.Stdout
	if err := c3.Start(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c2.Start(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c1.Run(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c2.Wait(); err != nil {
		return "", "", errors.Trace(err)
	}
	if err := c3.Wait(); err != nil {
		return "", "", errors.Trace(err)
	}
	return remoteName, remoteAddr, nil
}

func currentUser() (string, error) {
	out, err := p4CMD("user", "-o")
	if err != nil {
		return "", errors.Annotate(err, "Cannot find a active user")
	}
	reg := regexp.MustCompile("(?m)^User:.+")
	str := reg.FindString(out)
	userName := strings.Split(str, "\t")[1]
	return userName, nil
}

func (r *Repo) Exists(name string) (bool, error) {
	cfg, err := config.Platform()
	if err != nil {
		return false, errors.Trace(err)
	}

	addr, err := cfg.P4ServerAddr()
	if err != nil {
		return false, errors.Trace(err)
	}
	authCfg, err := config.Auth()
	if err != nil {
		return false, errors.Trace(err)
	}

	password, err := authCfg.GetP4UserPassword()
	if err != nil {
		return false, errors.Trace(err)
	}
	out, err := p4CMD("-p", addr, "-u", name, "-P", password, "users")
	if err != nil {
		return false, errors.Annotate(err, "unable to validate user: "+name)
	}

	if strings.Contains(out, name) {
		return true, nil
	}
	return false, nil
}

func (r *Repo) OwnerAndNameFromRemote() (string, string, error) {
	authCfg, err := config.Auth()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	repoOwner, err := authCfg.GetP4UserName()
	if err != nil {
		return "", "", errors.Trace(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", "", errors.Trace(err)
	}

	return repoOwner, filepath.Base(dir), nil
}

// AssertNotTracked checks for the existence of the appropriate
// codelingo remote to avoid duplications on GOGS.
func (r *Repo) AssertNotTracked() error {
	return nil
}

func (r *Repo) CreateRemote(name string) error {
	user, err := currentUser()
	if err != nil {
		return errors.Trace(err)
	}
	isRemoteUserExist, err := r.Exists(user)
	if !isRemoteUserExist {
		return errors.Trace(err)
	}
	return nil
}

func (r *Repo) Sync(repoOwner string, workingDir string) error {
	cfg, err := config.Platform()
	if err != nil {
		return errors.Trace(err)
	}
	remoteName, err := cfg.P4RemoteName()
	if err != nil {
		return errors.Trace(err)
	}
	authCfg, err := config.Auth()
	if err != nil {
		return errors.Trace(err)
	}
	password, err := authCfg.GetP4UserPassword()
	if err != nil {
		return errors.Trace(err)
	}
	out, err := p4CMD("-u", repoOwner, "-P", password, "push", "-r", remoteName)
	if err != nil {
		return errors.Annotate(errors.Trace(err), out)
	}
	return nil
}

func (r *Repo) CurrentCommitId() (string, error) {
	// TODO (Junyu) check if submit identity is set. If it is, then return the unique ID, otherwise return ""
	return "", nil
}

// WorkingDir returns a string representing the user's current directory in the format of the
// it will be represented in the store plus a trailing "/"
func (r *Repo) WorkingDir() (string, error) {
	out, err := p4CMD("client", "-o")
	if err != nil {
		return "", errors.Annotate(err, out)
	}
	reg := regexp.MustCompile("(?m)^Root:.+")
	str := reg.FindString(out)
	root := strings.Split(str, "\t")[1]

	out, err = p4CMD("where")
	if err != nil {
		return "", errors.Annotate(err, out)
	}
	reg = regexp.MustCompile(root + ".+")
	str = reg.FindString(out)
	workingDir := strings.Split(strings.Split(str, root+"/")[1], "...")[0]
	return workingDir, nil
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
	return nil
}

func (r *Repo) ApplyPatch(diff string) error {
	return nil
}

func (r *Repo) CheckoutRemote(sha string) error {
	return nil
}

func (r *Repo) ClearChanges() error {
	return nil
}

func p4CMD(args ...string) (out string, err error) {
	cmd := exec.Command("p4", args...)
	b, err := cmd.CombinedOutput()
	return string(b), errors.Annotate(err, string(b))
}
