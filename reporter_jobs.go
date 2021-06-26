package traceandtracego

import (
	"context"
	"io"
	"log"
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go/tracer"
	opentracing "github.com/opentracing/opentracing-go"
)

//生成 trace jobs
func GenerateTracingJobs(pch chan<- context.Context, parent context.Context,
	svc string, header http.Header, tags map[string]string, traceType string) {
	//设置 context
	ctx, cancel := context.WithCancel(parent)
	//设置通道
	ch := make(chan context.Context, 0)
	go doTask(ch, ctx, svc, header, tags, traceType)
	//接受信号
	pctx := <-ch
	pch <- pctx
	//销毁资源
	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		default:
			break
		}
	}
}

//执行 trace reporter
func doTask(ch chan context.Context, parent context.Context,
	svc string, header http.Header, tags map[string]string, traceType string) {
	//定义 tracer, closer
	var tracer opentracing.Tracer
	var closer io.Closer
	var ctx context.Context
	//选择 reporter 类别
	switch traceType {
	case "jaeger":
		tracer, closer = tracing.InitJaeger(svc)
		ctx = tracing.AddTracer(svc, parent, header, tracer, tags)
		break
	case "zipkin":
		log.Println("create zipkin tracing job")
		tracer, closer = tracing.InitZipkin(svc)
		ctx = tracing.AddTracer(svc, parent, header, tracer, tags)
		break
	case "skyWalking":
		log.Println("create skywalking tracing job")
		break
	default:
		break
	}

	defer closer.Close()
	ch <- ctx
}
