package traceandtracego

import (
	logger "log"

	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//声明 tracer
var tracer opentracing.Tracer

//text map reader
type TextMapReader struct {
	metadata.MD
}

//text map writer
type TextMapWriter struct {
	metadata.MD
}

//RPC Client Dial Option
func ClientDialOption(parentTracer opentracing.Tracer) grpc.DialOption {
	tracer = parentTracer
	return grpc.WithUnaryInterceptor(grpcClientInterceptor)
}

//RPC Server Dial Option
func ServerDialOption(tracer opentracing.Tracer) grpc.ServerOption {
	return grpc.UnaryInterceptor(grpcServerInterceptor)
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

//RPC Server 拦截器
func grpcServerInterceptor(
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

//text map writer set
func (t TextMapWriter) Set(key, val string) {
	t.MD[key] = append(t.MD[key], val)
}

// 读取 metadata 中的 span 信息
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
