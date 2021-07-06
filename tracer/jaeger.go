package traceandtracego

import (
	"io"
	logger "log"
	"os"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

//初始化 Jaeger
func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	// trace default setting
	cfg, err := jaegercfg.FromEnv()
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1
	cfg.Reporter.LogSpans = false
	// setting trace sampler type
	if traceSamplerType := os.Getenv("TRACE_SAMPLER_TYPE"); traceSamplerType != "" {
		cfg.Sampler.Type = traceSamplerType
	}
	// setting trace sampler type param
	if traceSamplerParam := os.Getenv("TRACE_SAMPLER_PARAM"); traceSamplerParam != "" {
		cfg.Sampler.Param, _ = strconv.ParseFloat(traceSamplerParam, 10)
	}
	// setting trace host
	if traceAgentHost := os.Getenv("TRACE_AGENT_HOST"); traceAgentHost != "" {
		cfg.Reporter.LocalAgentHostPort = traceAgentHost
	} else if traceCollectorEndpoint := os.Getenv("TRACE_ENDPOINT"); traceCollectorEndpoint != "" {
		// cfg.Reporter.CollectorEndpoint = "http://localhost:14268/api/traces"
		cfg.Reporter.CollectorEndpoint = traceCollectorEndpoint
	}

	if traceLogSpans := os.Getenv("TRACE_REPORTER_LOG_SPANS"); traceLogSpans != "" {
		cfg.Reporter.LogSpans, _ = strconv.ParseBool(traceLogSpans)
	}

	tracer, closer, err := cfg.New(service, jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		logger.Fatalln("cannot init Jaeger", err)
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

//write sub span
// func WriteSubSpan(span opentracing.Span, subSpanName string) {
// 	//use context
// 	ctx := context.Background()
// 	ctx = opentracing.ContextWithSpan(ctx, span)
// 	//其他过程获取并开始子 span
// 	newSpan, _ := opentracing.StartSpanFromContext(ctx, subSpanName)
// 	//StartSpanFromContext 会将新span保存到ctx中更新
// 	defer newSpan.Finish()
// }
