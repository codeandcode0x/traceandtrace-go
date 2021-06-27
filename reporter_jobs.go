/** Package traceandtrace-go 是 go 语言 tracing lib, 可以集成不同的 tracer 如: jeager、zipkin、skywalking 等
reporter_job 可以对 tracing 数据上报，做了性能优化，做到业务侵入小, 高性能等。

// 快速上手
	import (
    	tracing "github.com/codeandcode0x/traceandtrace-go"
	)

	// 在 func 中 或者 middleware 中添加
	_, cancel := tracing.AddHttpTracing("HttpTracingTest", [your http Header], map[string]string{"version": "v1"})
	defer cancel()

	...

	reporter_job 对上报 tracing 数据进行了优化 (采用携程任务处理),对业务侵入小，高性能上报,job 结束后, 可以对资源进行释放。
*/
package traceandtracego

import (
	"context"
	"io"
	"log"
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go/tracer"
	opentracing "github.com/opentracing/opentracing-go"
)

//generate trace jobs (goroutine), use context and chan to control jobs and release goroutine .
func GenerateTracingJobs(pch chan<- context.Context, parent context.Context, svc string, header http.Header, tags map[string]string, traceType string) {
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
		break
	case "skywalking":
		log.Println("create skywalking tracing job")
		break
	default:
		break
	}

	defer closer.Close()
	ch <- ctx
}
