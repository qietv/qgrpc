package main

import (
	"context"
	"example/proto"
	"fmt"
	"github.com/qietv/qgrpc"
	"google.golang.org/grpc"
)

type demoServer struct{}

func (ds *demoServer) Test(c context.Context, req *proto.TestReq) (resp *proto.TestResp, err error) {
	resp = &proto.TestResp{
		Pong: "pong",
	}
	return
}

func main() {
	var (
		err error
		s  *qgrpc.Server
	)
	s, err = qgrpc.New(&qgrpc.Config{
		Name:        "demo",
		Network:     "tcp",
		Addr:        ":8809",
		AccessLog:   "access.log",
		ErrorLog:    "error.log",
		Interceptor: nil,
	}, func(s *qgrpc.Server) {
		proto.RegisterTestServer(s.Server, &demoServer{})
	})
	if err != nil {
		panic(fmt.Sprintf("grpc server start fail, %s", err.Error()))
	}
	defer s.Server.GracefulStop()


	conn, err := grpc.Dial("localhost:8809", grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	client := proto.NewTestClient(conn)
	resp, err := client.Test(context.Background(), &proto.TestReq{
		Ping:                 "aaa",
	})
	println(resp.Pong)
}
