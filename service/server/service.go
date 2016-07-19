package server

import "github.com/codelingo/lingo/service/grpc/codelingo"

type CodeLingoService interface {
	Query(src string) (string, error)
	Review(*ReviewRequest) ([]*codelingo.Issue, error)
}
