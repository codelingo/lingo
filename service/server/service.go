package server

import "github.com/codelingo/lingo/service/grpc/codelingo"

type CodeLingoService interface {
	Session(*SessionRequest) (string, error)
	Query(src string) (string, error)
	Review(*ReviewRequest) (<-chan *codelingo.Issue, error)
}
