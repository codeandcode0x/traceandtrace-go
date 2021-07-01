package traceandtracego

import (
	"io"
	"log"
	logger "log"
	"net/http"
	"os"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	spanLog "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//声明 tracer
var tracer opentracing.Tracer
var ctxShare context.Context
var rpcCtx string
var sf = 100

type TextMapReader struct {
	metadata.MD
}

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
func WriteSubSpan(span opentracing.Span, subSpanName string) {
	//use context
	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)
	//其他过程获取并开始子 span
	newSpan, _ := opentracing.StartSpanFromContext(ctx, subSpanName)
	//StartSpanFromContext 会将新span保存到ctx中更新
	defer newSpan.Finish()
}

// TracerWrapper tracer wrapper
func AddTracer(svcName, spanName string,
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

//RPC Client Dial Option
func ClientDialOption(parentTracer opentracing.Tracer) grpc.DialOption {
	tracer = parentTracer
	return grpc.WithUnaryInterceptor(grpcClientInterceptor)
}

//text map writer
type TextMapWriter struct {
	metadata.MD
}

//text map writer set
func (t TextMapWriter) Set(key, val string) {
	t.MD[key] = append(t.MD[key], val)
}

//RPC Client 拦截器
func grpcClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) (err error) {

	//从context中获取metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		//如果对metadata进行修改，那么需要用拷贝的副本进行修改
		md = md.Copy()
	}
	//carrier := opentracing.TextMapCarrier{}
	carrier := TextMapWriter{md}
	//父类 context
	var currentContext opentracing.SpanContext
	//从 context 中获取原始的 span
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		currentContext = parentSpan.Context()
	} else {
		//否则创建 span
		span := tracer.StartSpan(method)
		defer span.Finish()
		currentContext = span.Context()
	}
	//将 span 的 context 信息注入到 carrier 中
	e := tracer.Inject(currentContext, opentracing.TextMap, carrier)
	if e != nil {
		logger.Fatalln("tracer inject failed", e)
	}
	//创建一个新的 context，把 metadata 附带上
	ctx = metadata.NewOutgoingContext(ctx, md)
	return invoker(ctx, method, req, reply, cc, opts...)
}

//RPC Server Dial Option
func ServerDialOption(tracer opentracing.Tracer) grpc.ServerOption {
	return grpc.UnaryInterceptor(jaegerGrpcServerInterceptor)
}

//读取metadata中的span信息
func (t TextMapReader) ForeachKey(handler func(key, val string) error) error { //不能是指针
	for key, val := range t.MD {
		for _, v := range val {
			if err := handler(key, v); err != nil {
				return err
			}
		}
	}
	return nil
}

//RPC Server 拦截器
func jaegerGrpcServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	//从context中获取metadata。md.(type) == map[string][]string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		//如果对metadata进行修改，那么需要用拷贝的副本进行修改。（FromIncomingContext的注释）
		md = md.Copy()
	}
	carrier := TextMapReader{md}
	tracer := opentracing.GlobalTracer()
	spanContext, e := tracer.Extract(opentracing.TextMap, carrier)
	if e != nil {
		logger.Fatalln("extract span context err", e)
	}

	span := tracer.StartSpan(info.FullMethod, opentracing.ChildOf(spanContext))
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return handler(ctx, req)
}
