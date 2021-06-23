package trace

import (
	"io"
	jaegertrace "traceandtrace-go/reporter/jaeger"

	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

//add rpc client tracing
func AddRpcClientTracing(serviceName string) (grpc.DialOption, io.Closer) {
	//初始化 jaeger
	tracer, closer := jaegertrace.InitJaeger(serviceName)
	//返回 rpc options
	return jaegertrace.ClientDialOption(tracer), closer
}

//add rpc server tracing
func AddRpcServerTracing(serviceName string) (grpc.ServerOption, io.Closer, opentracing.Tracer) {
	//初始化 jaeger
	tracer, closer := jaegertrace.InitJaeger(serviceName)
	//返回 rpc options
	return jaegertrace.ServerDialOption(tracer), closer, tracer
}

//zipkin
func AddZipkinTracer(serviceName string) {
	//TO-DO
}

//skywalking
func AddSkyWalkingTracer(serviceName string) {
	//TO-DO
}
