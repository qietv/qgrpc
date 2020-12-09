package main

import (
	"context"
	"fmt"
	"github.com/qietv/qgrpc"
	"google.golang.org/grpc"
)

type demoServer struct{}

func main() {
	s, err := qgrpc.New(&qgrpc.Config{
		Name:        "demo",
		Network:     "tcp",
		Addr:        ":8809",
		AccessLog:   "access.log",
		ErrorLog:    "error.log",
		Interceptor: nil,
	}, func(s *grpc.Server) {
		demo.RegisterGRPCDemoServer(s, &demoServer{})
	})
	if err != nil {
		panic(fmt.Sprintf("grpc server start fail, %s", err.Error()))
	}
	defer s.Server.GracefulStop()

	//Test server
	conn, err := grpc.Dial("localhost:8809", grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}

	gRPCClient := demoServer.NewGRPCDemoClient(conn)
	ret, err := gRPCClient.GetList(context.Background(), &demo.ListReq{
		Start: 0,
		Limit: 10,
	})
}
