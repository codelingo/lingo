package main

import (
	"log"

	pb "github.com/codelingo/lingo/service/grpc/codelingo"
	"github.com/waigani/xxx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:8002"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCodeLingoClient(conn)
	// Contact the server and print out its response.
	r, err := c.ListFacts(context.Background(), &pb.ListFactsRequest{Lexicon: "codelingo/golang"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Facts)
	xxx.Dump(r)
}
