package review

import "os/exec"

// TODO(waigani) provide sync for github repos

// Sync HEAD of the current branch with remote.
func sync() error {

	cmd := exec.Command("git", "push", "codelingo", "HEAD")
	return cmd.Run()

}
