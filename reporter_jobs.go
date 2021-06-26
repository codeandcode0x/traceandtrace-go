package traceandtracego

import (
	"context"
	"io"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

//生成 trace jobs
func GenerateTracingJobs(pch chan<- context.Context, parent context.Context, r *http.Request, svc, traceType string, tags map[string]string) {
	//设置 context
	ctx, cancel := context.WithCancel(parent)
	//设置通道
	ch := make(chan context.Context, 0)
	go doTask(ch, ctx, r, svc, traceType, tags)
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
func doTask(ch chan context.Context, parent context.Context, r *http.Request, svc, traceType string, tags map[string]string) {
	//定义 tracer, closer
	var tracer opentracing.Tracer
	var closer io.Closer
	var ctx context.Context
	//选择 reporter 类别
	switch traceType {
	case "Jaeger":
		log.Println("create jaeger tracing job")
		tracer, closer = InitJaeger(svc)
		ctx = AddTracer(parent, r, tracer, tags)
		break
	case "Zinkin":
		log.Println("create zinkin tracing job")
		tracer, closer = InitZipkin(svc)
		ctx = AddTracer(parent, r, tracer, tags)
		break
	case "SkyWalking":
		log.Println("create skywalking tracing job")
		break
	default:
		break
	}

	defer closer.Close()
	log.Println("tracing job finish ...")
	ch <- ctx
}
