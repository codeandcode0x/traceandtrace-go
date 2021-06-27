package traceandtracego

import (
	logger "log"

	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

//初始化 zipkin
func InitZipkin(service string) (opentracing.Tracer, reporter.Reporter) {
	//设置 span reporter
	reporter := httpreporter.NewReporter("http://localhost:9411/api/v2/spans")
	//创建本地 service 节点
	endpoint, err := zipkin.NewEndpoint(service, "")
	//log error
	if err != nil {
		logger.Println("[traceandtrace] [Error] unable to create local endpoint ...", err)
	}
	//初始化 tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logger.Println("[traceandtrace] [Error] unable to create tracer ...", err)
	}
	//wrap tracer
	tracer := zipkinot.Wrap(nativeTracer)
	//返回 trace & reporter
	return tracer, reporter
}
