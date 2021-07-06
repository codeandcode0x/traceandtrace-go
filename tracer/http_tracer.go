package traceandtracego

import (
	"log"
	logger "log"
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	spanLog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

// add jaeger http tracer
func AddHttpTracer(svcName, spanName string,
	ctx context.Context,
	header http.Header,
	tracer opentracing.Tracer,
	tags map[string]string) context.Context {
	//初始化 tracer
	opentracing.InitGlobalTracer(tracer)
	var sp opentracing.Span
	//从 header 中获取 span
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))
	if spanCtx != nil {
		sp = opentracing.GlobalTracer().StartSpan(spanName, opentracing.ChildOf(spanCtx))
	} else {
		//如果 header 中没有携带 context, 则新建 span
		sp = tracer.StartSpan(spanName)
	}
	//写入 tag 或者 日志
	for k, v := range tags {
		sp.LogFields(
			spanLog.String(k, v),
		)
	}

	// setting ext
	if _, exist := tags["spanKind"]; exist {
		// enum: client,server,producer,consumer
		ext.SpanKind.Set(sp, ext.SpanKindEnum(tags["spanKind"]))
	}

	if _, exist := tags["component"]; exist {
		ext.Component.Set(sp, tags["component"])
	}

	if _, exist := tags["samplingPriority"]; exist {
		spUint, err := strconv.Atoi(tags["samplingPriority"])
		if err != nil {
			log.Fatalf("sampling priority strconv error %v", err)
		}
		ext.SamplingPriority.Set(sp, uint16(spUint))
	}

	if _, exist := tags["peerService"]; exist {
		ext.PeerService.Set(sp, tags["peerService"])
	}

	if _, exist := tags["peerAddress"]; exist {
		ext.PeerAddress.Set(sp, tags["peerAddress"])
	}

	if _, exist := tags["peerHostname"]; exist {
		ext.PeerHostname.Set(sp, tags["peerHostname"])
	}

	if _, exist := tags["peerIpv4"]; exist {
		pi, _ := strconv.ParseUint(tags["peerIpv4"], 10, 32)
		ext.PeerHostIPv4.Set(sp, uint32(pi))
	}

	if _, exist := tags["peerIpv6"]; exist {
		ext.PeerHostIPv6.Set(sp, tags["peerIpv6"])
	}

	if _, exist := tags["peerPort"]; exist {
		ppInt, _ := strconv.Atoi(tags["peerPort"])
		ext.PeerPort.Set(sp, uint16(ppInt))
	}

	if _, exist := tags["httpUrl"]; exist {
		ext.HTTPUrl.Set(sp, tags["httpUrl"])
	}

	if _, exist := tags["httpStatusCode"]; exist {
		hscInt, _ := strconv.Atoi(tags["httpStatusCode"])
		ext.HTTPStatusCode.Set(sp, uint16(hscInt))
	}

	if _, exist := tags["dbStatement"]; exist {
		ext.DBStatement.Set(sp, tags["dbStatement"])
	}

	if _, exist := tags["dbInstance"]; exist {
		ext.DBInstance.Set(sp, tags["dbInstance"])
	}

	if _, exist := tags["dbType"]; exist {
		ext.DBType.Set(sp, tags["dbType"])
	}

	if _, exist := tags["httpMethod"]; exist {
		ext.HTTPMethod.Set(sp, tags["httpMethod"])
	}

	if _, exist := tags["dbUser"]; exist {
		ext.DBUser.Set(sp, tags["dbUser"])
	}

	if _, exist := tags["messageBusDestination"]; exist {
		ext.MessageBusDestination.Set(sp, tags["messageBusDestination"])
	}

	//注入span (用于传递)
	if err := opentracing.GlobalTracer().Inject(
		sp.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header)); err != nil {
		logger.Fatalln("inject failed", err)
	}
	//关闭连接
	defer sp.Finish()
	//返回带有 span 的 context
	return opentracing.ContextWithSpan(ctx, sp)
}
