package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	tracing "github.com/codeandcode0x/traceandtrace-go"
	pb "github.com/codeandcode0x/traceandtrace-go/example/protos/helloworld"
	"google.golang.org/grpc"
)

func main() {
	httpServer()
}

// http to gRPC
func httpServer() {
	http.HandleFunc("/rpc/tracing", func(w http.ResponseWriter, r *http.Request) {
		pctx, cancel := tracing.AddHttpTracing(
			"HttpServer",
			"/rpc/tracing GET", r.Header,
			map[string]string{"version": "v1"})
		defer cancel()
		// rpc tracing
		result := RpcClient(pctx)
		// return http request
		io.WriteString(w, "server: "+result)
	})
	log.Fatal(http.ListenAndServe(":9090", nil))
}

//grpc request
func RpcClient(ptx context.Context) string {
	rpcOption, closer := tracing.AddRpcClientTracing(
		"RpcClient",
		map[string]string{"version": "v1"})
	// or map[string]string{"traceType": "zipkin", "version": "v1"}), traceType : jaeger (default) or zipkin
	// or export TRACE_TYPE=zipkin or jaeger
	defer closer.Close()
	address := "localhost:22530"
	conn, err := grpc.Dial(address, grpc.WithInsecure(), rpcOption)
	if err != nil {
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	// Contact the server and print out its response.
	name := "rpc test"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	// use parent context
	ctx, cancel := context.WithTimeout(ptx, time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Println("error:", err)
	}

	log.Printf("Greeting: %s", r.Message)
	return r.Message
}
