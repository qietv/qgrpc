package qgrpc

import (
	"context"
	"fmt"
	"github.com/qietv/qgrpc/pkg"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"net"
	"sync"
	"time"
)

//gRPC Server config for qietv
type Config struct {
	Name              string                   `yaml:"name,omitempty"`
	Network           string                   `yaml:"network,omitempty"`
	Addr              string                   `yaml:"addr,omitempty"`
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

//Server gRPC server for qietv mico-service server
type Server struct {
	conf *Config
	mu   sync.Mutex
	*grpc.Server
	listener net.TCPListener
}

func (s *Server) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.Service == s.conf.Name {
		// check the server overall health status.
		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_SERVING,
		}, nil
	}
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_UNKNOWN,
	}, nil
}

func (s *Server) Watch(req *health.HealthCheckRequest, hW health.Health_WatchServer) error {
	return nil
}

func Default(registerFunc func(s *grpc.Server)) (s *Server, err error) {
	return New(&Config{
		Name:              "qgrpc",
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

// NewServer creates a gRPC server for qietv's mico service Server
// err when listen fail
func New(c *Config, registerFunc func(s *grpc.Server)) (s *Server, err error) {
	var (
		listener net.Listener
	)

	s = &Server{
		Server: grpc.NewServer(
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: time.Duration(c.IdleTimeout),
				//MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
				Time:             time.Duration(c.KeepAliveInterval),
				Timeout:          time.Duration(c.Timeout),
				MaxConnectionAge: time.Duration(c.MaxLifeTime),
			}),
			initInterceptor(c.Name, c.AccessLog, c.ErrorLog, c.Interceptor),
		),
	}
	listener, err = net.Listen(c.Network, c.Addr)
	if err != nil {
		err = fmt.Errorf("create server fail, %s", err.Error())
		return
	}
	registerFunc(s.Server)
	health.RegisterHealthServer(s.Server, s)
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
				service interface{}
				tracer  interface{}
				ok      bool
			)
			if service, ok = interceptor["service"]; !ok {
				if serviceName == "" {
					println("qgRPC trace conf fail, %+v", interceptor)
					continue
				}
				service = serviceName
			}
			if tracer, ok = interceptor["tracer"]; !ok {
				println("qgRPC trace conf fail, %+v", interceptor)
				continue
			}
			chain = append(chain, NewTracerInterceptor(service.(string), tracer.(string)))
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
