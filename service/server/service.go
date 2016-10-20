package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/codelingo/lingo/service/grpc/codelingo"
)

type Message string

type Messagec chan Message

type Issuec chan *codelingo.Issue

type CodeLingoService interface {
	Session(*SessionRequest) (string, error)
	Query(src string) (string, error)
	Review(*ReviewRequest) (Issuec, Messagec, error)
}

func (mc Messagec) Send(msgFmt string, vars ...interface{}) error {
	select {
	case mc <- Message(fmt.Sprintf(msgFmt, vars...)):
	case <-time.After(time.Second * 5):
		// TODO(waigani) error type
		return errors.New("timeout")
	}
	return nil
}

func (issc Issuec) Send(issue *codelingo.Issue) error {
	select {
	case issc <- issue:
	case <-time.After(time.Second * 30):
		// TODO(waigani) error type
		return errors.New("timeout")
	}
	return nil
}
