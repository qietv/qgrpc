package qgrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/qietv/qgrpc/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcHealth "google.golang.org/grpc/health"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Config gRPC Server config for qietv
type Config struct {
	Name              string                   `yaml:"name,omitempty"`
	Network           string                   `yaml:"network,omitempty"`
	Addr              string                   `yaml:"addr,omitempty"`
	Listen            string                   `yaml:"listen,omitempty"`
	HealthCheck       bool                     `yaml:"health,omitempty"`
	Timeout           pkg.Duration             `yaml:"timeout,omitempty"`
	IdleTimeout       pkg.Duration             `yaml:"idleTimeout,omitempty"`
	MaxLifeTime       pkg.Duration             `yaml:"maxLifeTime,omitempty"`
	ForceCloseWait    pkg.Duration             `yaml:"forceCloseWait,omitempty"`
	KeepAliveInterval pkg.Duration             `yaml:"keepAliveInterval,omitempty"`
	KeepAliveTimeout  pkg.Duration             `yaml:"keepAliveTimeout,omitempty"`
	AccessLog         string                   `yaml:"access,omitempty"`
	ErrorLog          string                   `yaml:"error,omitempty"`
	Interceptor       []map[string]interface{} `yaml:"interceptors,omitempty"`
}

// Server gRPC server for qietv mico-service server
type Server struct {
	conf *Config
	*grpc.Server
	listener  net.TCPListener
	healthSrv *grpcHealth.Server
}

func (s *Server) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	if s.healthSrv == nil {
		return nil, status.Error(codes.Unimplemented, "health check service is not enabled")
	}
	return s.healthSrv.Check(ctx, in)
}

func (s *Server) Watch(req *health.HealthCheckRequest, hW health.Health_WatchServer) error {
	if s.healthSrv == nil {
		return status.Error(codes.Unimplemented, "health check service is not enabled")
	}
	return s.healthSrv.Watch(req, hW)
}

func (s *Server) List(ctx context.Context, req *health.HealthListRequest) (*health.HealthListResponse, error) {
	if s.healthSrv == nil {
		return nil, status.Error(codes.Unimplemented, "health check service is not enabled")
	}
	return s.healthSrv.List(ctx, req)
}

func Default(registerFunc func(s *grpc.Server)) (s *Server, err error) {
	return New(&Config{
		Name:              "qietv",
		Network:           "tcp",
		Addr:              ":8808",
		Timeout:           pkg.Duration(time.Second * 20),
		IdleTimeout:       0,
		MaxLifeTime:       0,
		ForceCloseWait:    0,
		KeepAliveInterval: 0,
		KeepAliveTimeout:  0,
		AccessLog:         "./access.log",
		ErrorLog:          "./error.log",
		Interceptor:       nil,
	}, registerFunc)
}

// New  creates a gRPC server for Qietv's mico service Server
// err when listen fail
func New(c *Config, registerFunc func(s *grpc.Server)) (s *Server, err error) {
	var (
		listener net.Listener
	)

	s = &Server{
		Server: grpc.NewServer(
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     time.Duration(c.IdleTimeout),
				MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
				Time:                  time.Duration(c.KeepAliveInterval),
				Timeout:               time.Duration(c.Timeout),
				MaxConnectionAge:      time.Duration(c.MaxLifeTime),
			}),
			initInterceptor(c.Name, c.AccessLog, c.ErrorLog, c.Interceptor),
		),
		conf: c,
	}
	if c.Addr == "" {
		c.Addr = c.Listen
	}
	listener, err = net.Listen(c.Network, c.Addr)
	if err != nil {
		err = fmt.Errorf("create server fail, %s", err.Error())
		return
	}
	registerFunc(s.Server)
	if c.HealthCheck {
		s.healthSrv = grpcHealth.NewServer()
		if c.Name != "" {
			s.healthSrv.SetServingStatus(c.Name, health.HealthCheckResponse_SERVING)
		}
		health.RegisterHealthServer(s.Server, s)
	}
	go func() {
		err = s.Serve(listener)
		if err != nil {
			err = fmt.Errorf("grpc server fail, %s", err.Error())
		}
	}()
	return
}

func initInterceptor(serviceName, access, error string, interceptors []map[string]interface{}) grpc.ServerOption {
	var (
		chain []grpc.UnaryServerInterceptor
	)
	//init interceptor chain
	for _, interceptor := range interceptors {
		var (
			name interface{}
			has  bool
		)
		if name, has = interceptor["name"]; !has {
			println("qgRPC interceptor conf fail, %+v", interceptor)
			continue
		}
		switch name {
		case "trace":
			var (
				service string
				tracer  string
				ok      bool
			)
			if service, ok = interceptor["service"].(string); !ok {
				service = serviceName
			}
			if tracer, ok = interceptor["agent"].(string); !ok {
				tracer = ""
			}
			chain = append(chain, NewTracerInterceptor(service, tracer))
		case "metric":
			var (
				histogram    bool
				histogramVal interface{}
			)
			histogramVal = interceptor["histogram"]
			switch histogramVal.(type) {
			case string:
				if "true" == histogramVal.(string) {
					histogram = true
				}
			case nil:
				histogram = false
			case bool:
				histogram = histogramVal.(bool)
			}
			chain = append(chain, NewMetricInterceptor(serviceName, histogram))
		default:
			println("qgRPC interceptor not support, %+v", interceptor)
		}
	}
	if access != "" {
		chain = append(chain, NewLoggingInterceptor(access))
	}
	chain = append(chain, NewRecoveryInterceptor(error))
	return grpc.ChainUnaryInterceptor(chain...)
}
