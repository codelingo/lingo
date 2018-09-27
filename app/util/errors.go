package util

import (
	"regexp"
	"strings"

	"github.com/juju/errors"
	"gopkg.in/fatih/color.v1"
)

var repoNotFoundRegexp = regexp.MustCompile("fatal: repository '.*' not found.*")

type RepoExistsError string

func (r RepoExistsError) Error() string {
	return string(r)
}

func IsRepoExistsError(err error) bool {
	_, ok := err.(RepoExistsError)
	return ok
}

type UnauthorisedRepoError string

func (r UnauthorisedRepoError) Error() string {
	return string(r)
}

func IsUnauthorisedRepoError(err error) bool {
	_, ok := err.(UnauthorisedRepoError)
	return ok
}

// UserFacingWarning writes the given string to stderr in a coloured font.
func UserFacingWarning(str string) {
	errColor := color.New(color.FgYellow).SprintfFunc()
	msg := errColor("%s", str)
	Stderr.Write([]byte(msg + "\n"))
}

func UserFacingError(err error) {
	if err == nil {
		Logger.Debugf("got a nil error - this shouldn't be happening: %s", errors.ErrorStack(err))
		return
	}
	errColor := color.New(color.FgHiRed).SprintfFunc()
	msg := errColor("%s", userFacingErrMsg(err))
	Stderr.Write([]byte(msg + "\n"))
}

func FatalOSErr(err error) {
	UserFacingError(err)
	Exiter(1)
}

func userFacingErrMsg(err error) string {
	cause := errors.Cause(err)
	message := err.Error()

	switch {
	// Error types
	case IsUnauthorisedRepoError(cause):
		return "Sorry, you are not authorised to access this repo. Please run `$ lingo config setup` to authorise yourself."
	// Connection
	case strings.Contains(message, "all SubConns are in TransientFailure"):
		return "Sorry, the client failed to make a connection to the server. Please check your internet connection and try again."
	case strings.Contains(message, "transport is closing"):
		return "Sorry, a server error occurred and the connection was broken. Please try again."
	case strings.Contains(message, "ResourceExhausted"):
		return "Sorry, your request is too large and has been rejected by the server."
	// Config
	case repoNotFoundRegexp.MatchString(message):
		return "please run `lingo config setup`"
	// Git
	case strings.Contains(message, "fatal: Not a git repository"):
		return "This command can only be run in a git repository."
	}

	return message
}
