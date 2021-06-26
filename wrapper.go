package traceandtracego

import (
	"context"
	"io"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

//add tracing
func AddHttpTracing(r *http.Request, tags map[string]string, param ...map[string]string) (context.Context, context.CancelFunc) {
	var svcName, tType string
	//启动 trace 任务
	ctx, cancel := context.WithCancel(context.Background())
	//创建通道
	ch := make(chan context.Context, 0)
	//选择类型和服务
	tType = "Jaeger"
	svcName = r.URL.Path
	if len(param) > 0 {
		if _, exist := param[0]["serviceName"]; exist {
			svcName = param[0]["serviceName"]
		}

		if _, exist := param[0]["traceType"]; exist {
			tType = param[0]["traceType"]
		}
	}
	//创建任务
	go GenerateTracingJobs(ch, ctx, r, svcName, tType, tags)
	//返回通道
	return <-ch, cancel
}

//add rpc client tracing
func AddRpcClientTracing(serviceName string) (grpc.DialOption, io.Closer) {
	//初始化 jaeger
	tracer, closer := InitJaeger(serviceName)
	//返回 rpc options
	return ClientDialOption(tracer), closer
}

//add rpc server tracing
func AddRpcServerTracing(serviceName string) (grpc.ServerOption, io.Closer, opentracing.Tracer) {
	//初始化 jaeger
	tracer, closer := InitJaeger(serviceName)
	//返回 rpc options
	return ServerDialOption(tracer), closer, tracer
}

//zipkin
func AddZipkinTracer(serviceName string) {
	//TO-DO
}

//skywalking
func AddSkyWalkingTracer(serviceName string) {
	//TO-DO
}
