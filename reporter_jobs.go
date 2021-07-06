/** Package traceandtrace-go is the go language tracing lib, which can integrate different tracers such as: jeager, zipkin, skywalking, etc.
reporter_job can report tracing data and optimize performance to achieve low business intrusion and high performance.

// quick start
	import (
    	tracing "github.com/codeandcode0x/traceandtrace-go"
	)

	// Add in func or middleware
	_, cancel := tracing.AddHttpTracing("HttpTracingTest", [your http Header], map[string]string{"version": "v1"})
	defer cancel()

	...

	reporter_job optimizes the reported tracing data (using goroutine processing), has little business intrusion, and reports with high performance.
	After the job ends, resources can be released.
*/
package traceandtracego

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tracing "github.com/codeandcode0x/traceandtrace-go/tracer"
	opentracing "github.com/opentracing/opentracing-go"
)

//generate trace jobs (goroutine), use context and chan to control jobs and release goroutine .
func GenerateTracingJobs(pch chan<- context.Context, parent context.Context, svc, spanName string, header http.Header, tags map[string]string, traceType string) {
	// setting context
	ctx, cancel := context.WithCancel(parent)
	// setting chan
	ch := make(chan context.Context, 0)
	go doTask(ch, ctx, svc, spanName, header, tags, traceType)
	// receive signal
	pctx := <-ch
	pch <- pctx
	// destroy resources
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

// do trace reporter
func doTask(ch chan context.Context, parent context.Context,
	svc, spanName string, header http.Header, tags map[string]string, traceType string) {
	//define tracer, closer
	var tracer opentracing.Tracer
	var closer io.Closer
	var ctx context.Context
	//init tracer
	tracer, closer = SelectInitTracer(svc, map[string]string{"traceType": traceType})
	ctx = tracing.AddHttpTracer(svc, spanName, parent, header, tracer, tags)
	//close
	defer closer.Close()
	ch <- ctx
}

//select init tracer
func SelectInitTracer(svc string, param ...map[string]string) (opentracing.Tracer, io.Closer) {
	var tracer opentracing.Tracer
	var closer io.Closer
	// get tracer type
	traceType := JAEGER_TRACER
	if tType := os.Getenv("TRACE_TYPE"); tType != "" {
		traceType = tType
	} else if len(param) > 0 {
		if _, exist := param[0]["traceType"]; exist {
			traceType = strings.ToLower(param[0]["traceType"])
		}
	}

	// select reporter type
	switch traceType {
	case "jaeger":
		tracer, closer = tracing.InitJaeger(svc)
		break
	case "zipkin":
		tracer, closer = tracing.InitZipkin(svc)
		break
	case "skywalking":
		log.Println("create skywalking tracing job")
		break
	default:
		break
	}
	//return
	return tracer, closer
}
