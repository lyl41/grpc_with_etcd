package impl

import (
	"fmt"
	"golang.org/x/net/context"
	"grpc_with_etcd/server/protobuf"
)

type Server struct {
}

func (Server) Hello(ctx context.Context, req *hello.HelloRequest) (reply *hello.HelloReply, err error) {
	fmt.Println("recv msg", req.Msg)
	reply = new(hello.HelloReply)
	reply.Msg = "hello, " + req.Msg
	return
}
