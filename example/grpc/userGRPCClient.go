package main

import (
	"context"
	"fmt"
	"log"

	tracing "github.com/codeandcode0x/traceandtrace-go"
	user "github.com/codeandcode0x/traceandtrace-go/example/protos/user"

	"google.golang.org/grpc"
)

func main() {
	// add gRPC client tracing
	rpcOption, closer := tracing.AddRpcClientTracing("UserRpcClient")
	defer closer.Close()

	conn, err := grpc.Dial(":22530", grpc.WithInsecure(), rpcOption)
	if err != nil {
		log.Printf("faild to connect: %v", err)
	}
	defer conn.Close()

	c := user.NewUserRPCClient(conn)
	r, err := c.GetAllUsers(context.Background(), &user.UserMsgRequest{Count: 100})
	if err != nil {
		log.Printf("could not request: %v", err)
	}

	fmt.Printf("get user count : %s !\n", r.Message)
}
