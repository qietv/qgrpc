#### gRPC 封装

#### 目标
    

#### 使用方法

> go get github.com/qietv/qgrpc/pkg

```golang
    s, err := qgrpc.New(&qgrpc.Config{
        Name:              "qietv",
        Network:           "tcp",
        Addr:              ":8808",
        AccessLog:         "access.log",
        ErrorLog:          "error.log",
        Interceptor:       nil,
    }, func(s *grpc.Server) {
        user.RegisterGRPCBanServer(s, &banServer{})
    })
    if err != nil {
        panic("grpc server start fail")
    }
    defer s.Server.GracefulStop()

```

#### 里程碑
- [x] 健康检查
- [x] access && error log