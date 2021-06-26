package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	tracing "github.com/codeandcode0x/traceandtrace-go"
	pb "github.com/codeandcode0x/traceandtrace-go/example/helloworld/proto"
	"google.golang.org/grpc"
)

func main() {
	httpExample()
	gRPCExample()
}

// http tracing example
func httpExample() {
	_, cancel := tracing.AddHttpTracing(
		"HttpTracingTest",
		http.Header{},
		map[string]string{"version": "v1"})
	defer cancel()
}

// gRPC tracing example
func gRPCExample() {
	address := "localhost:22530"
	defaultName := "ethan"
	rpcOption, closer := tracing.AddRpcClientTracing("RpcClientTest")
	defer closer.Close()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), rpcOption)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
