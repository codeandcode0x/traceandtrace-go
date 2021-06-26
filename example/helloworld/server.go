package main

import (
	"context"
	"log"
	"net"

	pb "github.com/codeandcode0x/traceandtrace-go/example/helloworld/proto"

	tracing "github.com/codeandcode0x/traceandtrace-go"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port        = ":22530"
	addr        = "localhost:22530"
	serviceName = "RpcServerTest"
)

var tracer opentracing.Tracer
var ctxShare context.Context
var rpcCtx string

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	rpcOption, closer, _ := tracing.AddRpcServerTracing("RpcServerTest")
	defer closer.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(rpcOption)
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
