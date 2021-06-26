package traceandtracego

import (
	"context"
	"io"
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go/tracer"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

//add tracing
func AddHttpTracing(svcName string, header http.Header, tags map[string]string, param ...map[string]string) (context.Context, context.CancelFunc) {
	var tType string
	//启动 trace 任务
	ctx, cancel := context.WithCancel(context.Background())
	//创建通道
	ch := make(chan context.Context, 0)
	//选择类型和服务
	tType = "Jaeger"
	if len(param) > 0 {
		if _, exist := param[0]["traceType"]; exist {
			tType = param[0]["traceType"]
		}
	}
	//创建任务
	go GenerateTracingJobs(ch, ctx, svcName, header, tags, tType)
	//返回通道
	return <-ch, cancel
}

//add rpc client tracing
func AddRpcClientTracing(serviceName string) (grpc.DialOption, io.Closer) {
	//初始化 jaeger
	tracer, closer := tracing.InitJaeger(serviceName)
	//返回 rpc options
	return tracing.ClientDialOption(tracer), closer
}

//add rpc server tracing
func AddRpcServerTracing(serviceName string) (grpc.ServerOption, io.Closer, opentracing.Tracer) {
	//初始化 jaeger
	tracer, closer := tracing.InitJaeger(serviceName)
	//返回 rpc options
	return tracing.ServerDialOption(tracer), closer, tracer
}

//zipkin
func AddZipkinTracer(serviceName string) {
	//TO-DO
}

//skywalking
func AddSkyWalkingTracer(serviceName string) {
	//TO-DO
}