package trace

import (
	"context"
	"net/http"
	traceJob "traceandtrace-go/trace-jobs"
)

//add tracing
func AddTracing(r *http.Request, tags map[string]string, param ...map[string]string) (context.Context, context.CancelFunc) {
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
	go traceJob.GenerateTracingJobs(ch, ctx, r, svcName, tType, tags)
	//返回通道
	return <-ch, cancel
}
