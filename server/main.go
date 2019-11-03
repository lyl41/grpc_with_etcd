package main

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	grpcnaming "google.golang.org/grpc/naming"
	"grpc_with_etcd/server/impl"
	hello "grpc_with_etcd/server/protobuf"
	"log"
	"net"
	"time"
)

type Server struct {
	etcdClient    *clientv3.Client
	grpcServer    *grpc.Server
	lease         clientv3.Lease //租约，绑定此服务注册的key
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	Addr          string
}

func (s *Server) Serve(listener net.Listener) (err error) {
	s.Addr = "localhost:61933"
	s.etcdClient, err = getEtcdClient(getEtcdEndPoints())
	if err != nil {
		return
	}
	err = s.registerService2Etcd()
	if err != nil {
		return
	}
	go s.etcdLeaseKeepAlive()
	if err = s.grpcServer.Serve(listener); err != nil {
		return
	}
	return
}

func (s *Server) registerService2Etcd() (err error) {
	if s.etcdClient == nil || s.grpcServer == nil {
		err = errors.New("etcd_client or grpcServer is nil")
		return
	}
	s.lease = clientv3.NewLease(s.etcdClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	//创建一个租约
	leaseGrantReponse, err := s.lease.Grant(ctx, 3)
	if err != nil {
		err = errors.Wrap(err, "Grant fail")
		return
	}
	//ETCD的命名解析
	resolver := naming.GRPCResolver{Client: s.etcdClient}
	//grpc命名数据
	data := grpcnaming.Update{
		Op:   grpcnaming.Add,
		Addr: s.Addr,
	}
	//etcd更新，加上租约选项
	err = resolver.Update(ctx, "/root/service/hello", data, clientv3.WithLease(leaseGrantReponse.ID))
	if err != nil {
		err = errors.Wrap(err, "resolver.Update fail")
		return
	}
	s.keepAliveChan, err = s.lease.KeepAlive(context.Background(), leaseGrantReponse.ID)
	if err != nil {
		err = errors.Wrap(err, "lease.KeepAlive fail")
		return
	}
	return
}

func (s *Server) etcdLeaseKeepAlive() (err error) {
	if s.keepAliveChan == nil {
		err = errors.Wrap(errors.New("keepAliveChan is nil"), "")
		return
	}
	//for {
	//	msg := <-s.keepAliveChan
	//	fmt.Println(msg, time.Now().Unix())
	//}
	return
}

const etcd_addr = `localhost:2379`

func main() {
	s := new(Server)
	implServer := new(impl.Server)
	grpcServer := grpc.NewServer()
	hello.RegisterHelloServer(grpcServer, implServer) //TODO
	s.grpcServer = grpcServer

	port := ":61933"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	e := s.Serve(lis)
	if e != nil {
		log.Fatal(e)
	}
}

func getEtcdEndPoints() []string {
	return []string{etcd_addr}
}

func getEtcdClient(endPoints []string) (client *clientv3.Client, err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: time.Second * 3,
	})
	return
}
