package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	//"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	hello "grpc_with_etcd/client/protobuf"
	"log"
	"time"
)

func ordinaryCall(target string) {
	conn, err := grpc.Dial(target,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := hello.NewHelloClient(conn)
	req := &hello.HelloRequest{
		Msg: "lyl",
	}
	reply, err := client.Hello(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}

func etcdCall(target string) {
	etcdClient, err := getEtcdClient(getEtcdEndPoints())
	if err != nil {
		fmt.Println("err", err)
		return
	}
	//etcd中实现了grpc naming包中的Resolver，可以使用这个Resolver创建一个grpc Balancer
	resolver := naming.GRPCResolver{Client: etcdClient}
	grpcBalancer := grpc.RoundRobin(&resolver)
	conn, err := grpc.Dial(target,
		grpc.WithInsecure(),
		grpc.WithBalancer(grpcBalancer),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := hello.NewHelloClient(conn)
	req := &hello.HelloRequest{
		Msg: "lyl",
	}
	//此处等待Watcher监听到etcd变化，添加服务端地址
	//可以grpc dial时传递failFirst参数，这样grpc会等待可用地址直到超时。
	time.Sleep(time.Millisecond * 20)
	reply, err := client.Hello(context.Background(), req)
	fmt.Println("err", err)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}

const etcd_addr = `localhost:2379`

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

func main() {
	//ordinaryCall("localhost:61933")
	etcdCall("/root/service/hello")
	//time.Sleep(time.Second)
}
