package main

//
//import (
//	"context"
//	"github.com/smallnest/rpcx/client"
//	"log"
//)
//
//func main() {
//
//	// TODO: 客户端需要有一套同样的类型，不能用服务端的
//	type Args struct {
//		A int
//		B int
//	}
//
//	type Reply struct {
//		C int
//	}
//	// #1
//	addr := "127.0.0.1"
//	port := "8972"
//	d, _ := client.NewPeer2PeerDiscovery("tcp@"+addr+":"+port, "")
//	// #2
//	xclient := client.NewXClient("Arith", client.Failtry, client.RandomSelect, d, client.DefaultOption)
//	defer xclient.Close()
//
//	// #3
//	args := &Args{
//		A: 10,
//		B: 20,
//	}
//
//	// #4
//	reply := &Reply{}
//
//	// #5
//	err := xclient.Call(context.Background(), "Mul", args, reply) // 这里一看就是反射实现的
//	if err != nil {
//		log.Fatalf("failed to call: %v", err)
//	}
//	log.Printf("%d * %d = %d", args.A, args.B, reply.C)
//}
