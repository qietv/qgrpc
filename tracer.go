package qgrpc

import (
	"context"
	"encoding/base64"
	"github.com/hanskorg/logkit"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-lib/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"io"
	"strings"
	"time"
)

//MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	metadata.MD
}

// Set implements Set() of opentracing.TextMapWriter
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, "-bin") {
		val := base64.StdEncoding.EncodeToString([]byte(val))
		val = string(val)
	}
	c.MD[key] = append(c.MD[key], val)
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// DialOption grpc client option
func DialOption(tracer opentracing.Tracer) grpc.DialOption {
	return grpc.WithUnaryInterceptor(clientInterceptor(tracer))
}

// NewTracer NewJaegerTracer for current service
func NewTracer(serviceName string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	tracer = opentracing.GlobalTracer()
	return
}

func NewJaegerTracer(serviceName string, jagentHost string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	cfg := config.Configuration{
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName:   "x-trace-id",
			TraceBaggageHeaderPrefix: "x-qt-",
		},
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
	}
	if jagentHost != "" {
		cfg.Reporter = &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  jagentHost,
		}
	}
	zkPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	tracer, closer, err = cfg.NewTracer(
		config.Logger(jaeger.NullLogger),
		config.Metrics(metrics.NullFactory),
		config.Extractor(opentracing.HTTPHeaders, zkPropagator),
		config.Injector(opentracing.HTTPHeaders, zkPropagator),
		config.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		grpclog.Errorf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	return
}

type clientSpanTagKey struct{}

//clientInterceptor grpc client wrapper
func clientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	if tracer == nil {
		zkPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
		tracer, _, _ = config.Configuration{Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName:   "x-trace-id",
			TraceBaggageHeaderPrefix: "x-qt-",
		}}.NewTracer(
			config.Logger(jaeger.NullLogger),
			config.Metrics(metrics.NullFactory),
			config.Extractor(opentracing.HTTPHeaders, zkPropagator),
			config.Injector(opentracing.HTTPHeaders, zkPropagator),
		)
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var (
			parentCtx opentracing.SpanContext
			//uid       int64
			//gateway   net.IP
			err    error
			tracer opentracing.Tracer
		)
		if QTracer != nil {
			tracer = QTracer
		} else {
			tracer = opentracing.GlobalTracer()
		}
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
		}
		traceOpts := []opentracing.StartSpanOption{
			opentracing.ChildOf(parentCtx),
			ext.SpanKindRPCClient,
		}
		if tagx := ctx.Value(clientSpanTagKey{}); tagx != nil {
			if opt, ok := tagx.(opentracing.StartSpanOption); ok {
				traceOpts = append(traceOpts, opt)
			}
		}
		span := tracer.StartSpan(method, traceOpts...)
		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		mdWriter := MDReaderWriter{md}
		if err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, mdWriter); err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}
		newCtx := opentracing.ContextWithSpan(metadata.NewOutgoingContext(ctx, md), span)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.String("call-error", err.Error()))
		}
		return err
	}
}

//serverInterceptor grpc server wrapper
func serverInterceptor(tracer opentracing.Tracer, withError bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			md   metadata.MD
			ok   bool
			span opentracing.Span
		)

		md, ok = metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			opentracing.SpanFromContext(ctx)
			logkit.Errorf("extract from metadata err: %v", err)
		}
		span = tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(parentSpanContext),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCServer,
		)
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)

		resp, err = handler(ctx, req)
		if err != nil && withError {
			span.LogFields(log.Object("error", err))
		}
		return
	}
}
