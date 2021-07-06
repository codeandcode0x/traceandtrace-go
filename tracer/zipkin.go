package traceandtracego

import (
	"log"
	"os"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

//初始化 zipkin
func InitZipkin(service string) (opentracing.Tracer, reporter.Reporter) {
	// setting trace host
	zipkinHost := "http://localhost:9411/api/v2/spans"
	if traceAgentHost := os.Getenv("TRACE_AGENT_HOST"); traceAgentHost != "" {
		zipkinHost = traceAgentHost
	} else if traceCollectorEndpoint := os.Getenv("TRACE_ENDPOINT"); traceCollectorEndpoint != "" {
		zipkinHost = traceCollectorEndpoint
	}
	//设置 span reporter
	reporter := httpreporter.NewReporter(zipkinHost)
	//创建本地 service 节点
	endpoint, err := zipkin.NewEndpoint(service, "")
	//log error
	if err != nil {
		log.Println("unable to create local endpoint ...", err)
	}

	var sampleParam uint64 = 1
	// setting trace sampler type param
	if traceSamplerParam := os.Getenv("TRACE_SAMPLER_PARAM"); traceSamplerParam != "" {
		sampleParam, _ = strconv.ParseUint(traceSamplerParam, 10, 64)
	}
	//setting sampler
	sampler := zipkin.NewModuloSampler(sampleParam)

	//初始化 tracer
	nativeTracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		log.Println("unable to create tracer ...", err)
	}

	//wrap tracer
	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)
	//返回 trace & reporter
	return tracer, reporter
}
