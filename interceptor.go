package qgrpc

import (
	"context"
	"fmt"
	"github.com/hanskorg/logkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"os"
	"runtime/debug"
	"time"
)

// NewLoggingInterceptor create log interceptor
// format {time}\t{origin}\t{method}\t{http code}\t{status.code}\t{status.message}
func NewLoggingInterceptor(logfile string) grpc.UnaryServerInterceptor {
	accessLogger, err := logkit.NewFileLogger(logfile, "", time.Second, uint64(1204*1024*1800), 0)
	if err != nil {
		println("logger conf fail, %s", err.Error())
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			startTime time.Time
			code      int32
			message   string
			remoteIp  string
			ret       *status.Status
			durTime   float32
		)
		if accessLogger == nil {
			return handler(ctx, req)
		}
		startTime = time.Now()
		resp, err = handler(ctx, req)
		if err != nil {
			ret = status.Convert(err)
			code = int32(ret.Code())
			message = ret.Message()
		}

		if pr, ok := peer.FromContext(ctx); ok {
			if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
				remoteIp = tcpAddr.IP.String()
			} else {
				remoteIp = pr.Addr.String()
			}
		}

		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			rips := md.Get("x-forward-for")
			if len(rips) != 0 {
				remoteIp = rips[0]
			}
		}
		if remoteIp == "" {
			remoteIp = "-"
		}
		if message == "" {
			message = "-"
		}
		durTime = float32(time.Since(startTime).Microseconds()) / 1000.0
		if _, err := accessLogger.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%s\t%-.3f\t%d\t%s\n", startTime.Format(time.RFC3339), remoteIp, info.FullMethod, "-", durTime, code, message))); err != nil {
			println("write access log fail, %s", err.Error())
		}
		return resp, err
	}
}

// NewRecoveryInterceptor grpc server ServerInterceptor
// create union revovery handler for qietv gRPC sevice
func NewRecoveryInterceptor(logfile string) grpc.UnaryServerInterceptor {
	var (
		errorLogger io.Writer
		err         error
		startTime   time.Time
		remoteIp    string
	)

	if logfile != "" {
		errorLogger, err = logkit.NewFileLogger(logfile, "", time.Second, uint64(1204*1024*1800), 0)
	} else {
		errorLogger, err = os.Open("/dev/stderr")
	}
	if err != nil {
		println("logger conf fail, %s", err.Error())
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		defer func() {
			if r := recover(); r != nil {
				if _, err := errorLogger.Write([]byte(fmt.Sprintf("%s\t%s\t%s\n fatal ===> %s %s \n===============\n", startTime.Format(time.RFC3339), remoteIp, info.FullMethod, r, debug.Stack()))); err != nil {
					println("write access log fail, %s", err.Error())
				}
				err = status.Errorf(codes.Internal, "fatal err: %+v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// NewTracerInterceptor grpc server ServerInterceptor
func NewTracerInterceptor(serviceName string, agentHost string) grpc.UnaryServerInterceptor {
	tracer, _, err := NewTracer(serviceName, agentHost)
	if err != nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			println("tracer setup fail, %s", err.Error())
			return handler(ctx, req)
		}
	}
	return serverInterceptor(tracer)
}
