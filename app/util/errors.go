package util

import (
	"regexp"
	"strings"

	"github.com/codelingo/lexicon/lib/util"
	"github.com/juju/errors"
	"gopkg.in/fatih/color.v1"
)

var repoNotFoundRegexp = regexp.MustCompile("fatal: repository '.*' not found.*")

func UserFacingError(err error) {
	if err == nil {
		util.Logger.Debugf("got a nil error - this shouldn't be happening: %s", errors.ErrorStack(err))
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

func userFacingErrMsg(mainErr error) string {
	message := mainErr.Error()

	switch {
	// Connection
	case strings.Contains(message, "all SubConns are in TransientFailure"):
		return "Sorry, the client failed to make a connection to the server. Please check your internet connection and try again."
	case strings.Contains(message, "transport is closing"):
		return "Sorry, a server error occurred and the connection was broken. Please try again."
	// Config
	case repoNotFoundRegexp.MatchString(message):
		return "please run `lingo config setup`"
	}

	return message
}
