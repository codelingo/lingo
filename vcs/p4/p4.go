package p4

// TODO (Junyu) remove any unnecessary functions.
import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/juju/errors"

	"regexp"

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

	out, err = p4CMD("remote", "-o", remoteName)
	if err != nil {
		return "", "", errors.Annotate(err, out)
	}
	in := strings.Replace(out, "Address:\tlocalhost:1666", "Address:\t"+remoteAddr, 1)
	cmd := exec.Command("p4", "remote", "-i")
	cmd.Stdin = bytes.NewReader([]byte(in))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Annotate(err, string(output))
	}

	out, err = p4CMD("remote", "-o", remoteName)
	if err != nil {
		return "", "", errors.Annotate(err, out)
	}
	in = strings.Replace(out, "//... //...", "//stream/main/... //depot/"+repoOwner+"/"+repoName+"/...", 1)
	cmd = exec.Command("p4", "remote", "-i")
	cmd.Stdin = bytes.NewReader([]byte(in))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Annotate(err, string(output))
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
	if strings.Contains(userName, "\r") {
		userName = strings.Split(userName, "\r")[0]
	}
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
	latestChangelist, err := p4CMD("changes", "-s", "submitted", "-m1")
	if err != nil {
		return "", errors.Annotate(err, latestChangelist)
	}
	out, err := p4CMD("change", "-o", strings.Split(latestChangelist, " ")[1])
	if err != nil {
		return "", errors.Annotate(err, latestChangelist)
	}
	reg := regexp.MustCompile("(?m)^Identity:.+")
	str := reg.FindString(out)
	if str == "" {
		return "", errors.New("submit identity is missing in the changelist")
	}
	return strings.TrimSpace(strings.Split(str, ":")[1]), nil
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
	if strings.Contains(root, "\r") {
		root = strings.Split(root, "\r")[0]
	}
	out, err = p4CMD("where")
	if err != nil {
		return "", errors.Annotate(err, out)
	}
	rootQM := regexp.QuoteMeta(root)
	if runtime.GOOS == "windows" {

		b, err := exec.Command("pwd").CombinedOutput()
		if err == nil {
			rootQM = strings.Replace(root, "\\", "/", -1)
			out = strings.Replace(out, "C:", "c:", 1)
			root = rootQM
		} else if !strings.Contains(err.Error(), "executable file not found") {
			return "", errors.Annotate(err, string(b))
		}

	}
	reg = regexp.MustCompile(rootQM + ".+")
	str = reg.FindString(out)
	subStr := strings.Split(str, root)[1]
	subStr = subStr[1 : len(subStr)-1]
	workingDir := strings.Split(subStr, "...")[0]
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

func callBash(dir string, args ...string) (out string, err error) {
	cmd := exec.Command("bash", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	b, err := cmd.CombinedOutput()
	out = string(b)
	return out, errors.Annotate(err, out)
}

func (r *Repo) GetDotlingoFilepathsInDir(dir string) ([]string, error) {
	out, err := callBash(dir, "-c", "p4 client -o")
	if err != nil {
		return nil, errors.Annotate(err, out)
	}
	reg := regexp.MustCompile("(?m)^Root:.+")
	str := reg.FindString(out)
	root := strings.Split(str, "\t")[1]

	staged, err := callBash(dir, "-c", "p4 files "+root+"/...")
	if err != nil {
		return nil, errors.Trace(err)
	}

	files := strings.Split(staged, "\n")

	for k, filepath := range files {
		if common.IsDotlingoFile(filepath) {
			serverPath := strings.Split(filepath, "#")[0]
			out, err := callBash(dir, "-c", "p4 where "+serverPath)
			if err != nil {
				return nil, errors.Annotate(err, out)
			}
			reg = regexp.MustCompile(root + ".+")
			files[k] = strings.Split(reg.FindString(out), root+"/")[1]
		}
	}

	dotlingoFilepaths := []string{}
	for _, filepath := range files {
		if common.IsDotlingoFile(filepath) {
			dotlingoFilepaths = append(dotlingoFilepaths, filepath)
		}
	}

	return dotlingoFilepaths, nil
}
