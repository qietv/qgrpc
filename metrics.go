package qgrpc

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mOnce sync.Once
)

const MetricsNS = "qietv"

type QGrpcMetrics struct {
	serverRequestCounter  *prometheus.CounterVec
	serverCompleteCounter *prometheus.CounterVec
	serverPanicCounter    *prometheus.CounterVec
	handleHistogram       *prometheus.HistogramVec
	handleHistogramEnable bool
	ServiceName           string
}

func NewMetrics(serviceName string, histogram bool) (m *QGrpcMetrics) {
	m = &QGrpcMetrics{
		serverRequestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: MetricsNS,
				Name:      "grpc_server_request_total",
				Help:      "Total number of grpc request from started server.",
			},
			[]string{"service", "method"},
		),
		serverCompleteCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: MetricsNS,
				Name:      "grpc_server_completed_total",
				Help:      "Total number of grpc completed from started server.",
			},
			[]string{"service", "method", "code"},
		),
		serverPanicCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: MetricsNS,
				//Subsystem: serviceName,
				Name: "grpc_server_panic_total",
				Help: "Total number of grpc method panic from started server.",
			},
			[]string{"service", "method"},
		),
		handleHistogram:       nil,
		handleHistogramEnable: false,
		ServiceName:           serviceName,
	}
	if histogram {
		m.handleHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: MetricsNS,
			//Subsystem: serviceName,
			Name:    "grpc_server_complete_millisecond",
			Help:    "Histogram of response latency (milliseconds) of gRPC that had been application-level handled by the server.",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "method"})
	}
	m.handleHistogramEnable = histogram
	return
}

func (m *QGrpcMetrics) ServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			startTime  time.Time
			grpcStatus *status.Status
			codeStr    string
			service    string
			method     string
		)
		if m.handleHistogramEnable {
			startTime = time.Now()
		}
		if m.ServiceName == "" {
			service, method = m.serviceMethodName(info.FullMethod)
		} else {
			_, method = m.serviceMethodName(info.FullMethod)
			service = m.ServiceName
		}
		m.serverRequestCounter.WithLabelValues(service, method).Inc()
		resp, err = handler(ctx, req)
		grpcStatus, _ = status.FromError(err)
		if grpcStatus.Code() > codes.Unauthenticated {
			codeStr = strconv.FormatInt(int64(grpcStatus.Code()), 10)
		} else {
			codeStr = grpcStatus.Code().String()
		}
		m.serverCompleteCounter.WithLabelValues(service, method, codeStr).Inc()
		if m.handleHistogramEnable {
			m.handleHistogram.WithLabelValues(service, method).Observe(float64(time.Since(startTime).Milliseconds()))
		}
		return
	}
}

func (m *QGrpcMetrics) AddPanicNumber(info *grpc.UnaryServerInfo) {
	service, method := m.serviceMethodName(info.FullMethod)
	m.serverPanicCounter.WithLabelValues(service, method).Inc()
	return
}

func (m *QGrpcMetrics) serviceMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return m.ServiceName, "unknown"
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *QGrpcMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.serverPanicCounter.Describe(ch)
	m.serverRequestCounter.Describe(ch)
	m.serverCompleteCounter.Describe(ch)
	if m.handleHistogramEnable {
		m.handleHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *QGrpcMetrics) Collect(ch chan<- prometheus.Metric) {
	m.serverPanicCounter.Collect(ch)
	m.serverRequestCounter.Collect(ch)
	m.serverCompleteCounter.Collect(ch)
	if m.handleHistogramEnable {
		m.handleHistogram.Collect(ch)
	}
}
