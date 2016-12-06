package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/codelingo/lingo/service/grpc/codelingo"
)

type Message string

type Messagec chan Message

type Ingestc chan string

type Issuec chan *codelingo.Issue

type CodeLingoService interface {
	Session(*SessionRequest) (string, error)
	Query(src string) (string, error)
	Review(*ReviewRequest) (Issuec, Messagec, Ingestc, error)
	ListLexicons() ([]string, error)
	ListFacts(lexicon string) (map[string][]string, error)
}

func (mc Messagec) Send(msgFmt string, vars ...interface{}) error {
	select {
	case mc <- Message(fmt.Sprintf(msgFmt, vars...)):
	case <-time.After(time.Second * 5):
		// TODO(waigani) error type
		return errors.New("timeout Messagec.Send: " + fmt.Sprintf(msgFmt, vars...))
	}
	return nil
}

func (ingc Ingestc) Send(s string) error {
	select {
	case ingc <- s:
	case <-time.After(time.Second * 5):
		return errors.New("timeout Ingestc.Send")
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
