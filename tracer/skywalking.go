package traceandtracego

import (
	logger "log"

	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

// var ctxShare context.Context
// var tracer opentracing.Tracer

//初始化 zipkin
func InitSkyWalking(service string) (opentracing.Tracer, reporter.Reporter) {
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

//添加 tracer
// func AddTracer(r *http.Request, tracer opentracing.Tracer) {
// 	opentracing.InitGlobalTracer(tracer)
// 	sp := tracer.StartSpan(r.URL.Path)
// 	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,
// 		opentracing.HTTPHeadersCarrier(r.Header))
// 	if spanCtx != nil {
// 		sp = opentracing.GlobalTracer().StartSpan(r.URL.Path, opentracing.ChildOf(spanCtx))
// 	} else {
// 		//http inject
// 		if err := opentracing.GlobalTracer().Inject(
// 			sp.Context(),
// 			opentracing.HTTPHeaders,
// 			opentracing.HTTPHeadersCarrier(r.Header)); err != nil {
// 			logger.Println("[traceandtrace] [Error] inject failed ...", err)
// 		}
// 	}
// 	//上下文记录父spanContext
// 	ctxShare = context.WithValue(context.Background(), "usergRpcCtx", opentracing.ContextWithSpan(context.Background(), sp))
// 	//close span
// 	defer sp.Finish()
// }
