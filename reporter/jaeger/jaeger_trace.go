package trace

import (
	opentracing "github.com/opentracing/opentracing-go"
	spanLog "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "golang.org/x/net/context"
    "os"
    "io"
    "net/http"
    logger "log"
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
    cfg, err := jaegercfg.FromEnv()
    cfg.Sampler.Type = "const"
    cfg.Sampler.Param = 1
    cfg.Reporter.LocalAgentHostPort = "jaeger-production-query.istio-system:6831"
    //设置 trace agent host
    if traceAgentHost := os.Getenv("TRACE_AGENT_HOST"); traceAgentHost != "" {
        cfg.Reporter.LocalAgentHostPort = traceAgentHost
    }
    cfg.Reporter.LogSpans = true
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
func AddTracer(ctx context.Context, r *http.Request, tracer opentracing.Tracer, tags map[string]string) context.Context {
	//初始化 tracer
	opentracing.InitGlobalTracer(tracer)
	var sp opentracing.Span
	//从 header 中获取 span
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, 
		opentracing.HTTPHeadersCarrier(r.Header))
	if spanCtx != nil {
		sp = opentracing.GlobalTracer().StartSpan(r.URL.Path, opentracing.ChildOf(spanCtx))
	}else{
	//如果 header 中没有携带 context, 则新建 span
		sp = tracer.StartSpan(r.URL.Path)
	}
	//写入 tag 或者 日志
	for k, v := range tags {
		// sp.LogKV(k, v)
		// sp.SetTag(k, v)
		sp.LogFields(
		    spanLog.String(k, v),
		)
	}
	//注入span (用于传递)
	if err := opentracing.GlobalTracer().Inject(
		sp.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header)); err != nil {
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
func grpcClientInterceptor (
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
    }else{
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




