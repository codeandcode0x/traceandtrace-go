package traceandtracego

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	tracing "github.com/codeandcode0x/traceandtrace-go/tracer"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const (
	JAEGER_TRACER     = "jaeger"
	ZIPKIN_TRACER     = "zipkin"
	SKYWALKING_TRACER = "skywalking"
)

//Add http tracing , tags is k-v map which can set in span log, param map can set trace type .
func AddHttpTracing(svcName, spanName string, header http.Header, tags map[string]string, param ...map[string]string) (context.Context, context.CancelFunc) {
	//trace type
	var traceType string
	//start trace task
	ctx, cancel := context.WithCancel(context.Background())
	//create chan
	ch := make(chan context.Context, 0)
	//trace type
	traceType = JAEGER_TRACER
	//get trace type env
	if tType := os.Getenv("TRACE_TYPE"); tType != "" {
		traceType = tType
	} else if _, exist := tags["traceType"]; exist {
		traceType = strings.ToLower(tags["traceType"])
	}
	//create goroutine job
	go GenerateTracingJobs(ch, ctx, svcName, spanName, header, tags, traceType)
	//return chan
	return <-ch, cancel
}

//add rpc client tracing
func AddRpcClientTracing(serviceName string, param ...map[string]string) (grpc.DialOption, io.Closer) {
	//init tracer
	tracer, closer := SelectInitTracer(serviceName, param...)
	//return rpc options
	return tracing.ClientDialOption(tracer), closer
}

//add rpc server tracing
func AddRpcServerTracing(serviceName string, param ...map[string]string) (grpc.ServerOption, io.Closer, opentracing.Tracer) {
	//init jaeger
	tracer, closer := SelectInitTracer(serviceName, param...)
	//return rpc options
	return tracing.ServerDialOption(tracer), closer, tracer
}
