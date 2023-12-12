package rpc

import (
	"context"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client" // TODO: 应该是第三方 etcd 插件？
	"github.com/sirupsen/logrus"                // TODO: 用来做日志处理的
	"github.com/smallnest/rpcx/client"
	"sync"
	"time"
)

// rpc 客户端
var LogicRpcClient client.XClient // TODO: 单例，只会有一个且不会被覆盖

var once sync.Once

type RpcLogic struct{} // TODO: 命名含义：向 logic 层的 rpc 请求

var RpcLogicObj *RpcLogic // 进行 rpc 操作的对象 TODO: 单例，只会有一个且不会被覆盖，依靠 once.Do() 的单次执行机制实现

func InitLogicRpcClient() { // TODO: 该方法在本层入口文件处调用
	once.Do(func() { // TODO: 不管 Do 中的方法是什么，once.Do() 只能执行一次
		// TODO: 这些选项都是什么
		etcdConfigOption := &store.Config{
			ClientTLS:         nil,
			TLS:               nil,
			ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
			Bucket:            "",
			PersistConnection: true,
			Username:          config.Conf.Common.CommonEtcd.UserName,
			Password:          config.Conf.Common.CommonEtcd.Password,
		}
		// TODO: 下面这些参数的值具体是什么
		d, err := etcdV3.NewEtcdV3Discovery( // TODO: etcd 服务发现
			config.Conf.Common.CommonEtcd.BasePath,
			config.Conf.Common.CommonEtcd.ServerPathLogic, // TODO: 这个是 rpc 服务的名称吗
			[]string{config.Conf.Common.CommonEtcd.Host},  // TODO: 这个是 etcd 集群的地址吗
			true,
			etcdConfigOption,
		)
		if err != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail:%s", err.Error())
		}
		// 实例化 rpc 客户端
		// 随机选择，单点重试
		// d 是服务发现方式，这里使用的是 etcd 服务发现（应该是用的第三方插件，官网不是这么写的）
		// 上面是客户端设置了 EtcdDiscovery 插件，设置 basepath 和 etcd 集群的地址。
		LogicRpcClient = client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathLogic, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		RpcLogicObj = new(RpcLogic)
		if LogicRpcClient == nil {
			logrus.Fatalf("get logic rpc client nil")
		}
	})
}

func (rpc *RpcLogic) Login(req *proto.LoginRequest) (code int, authToken string, msg string) {
	reply := &proto.LoginResponse{}
	// req 是参数，reply 是 rpc 远程调用返回的结果
	err := LogicRpcClient.Call(context.Background(), "Login", req, reply) // TODO: 服务注册在哪，服务方法定义在哪
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) Register(req *proto.RegisterRequest) (code int, authToken string, msg string) {
	reply := &proto.RegisterReply{}
	err := LogicRpcClient.Call(context.Background(), "Register", req, reply) // TODO: 服务注册在哪，服务方法定义在哪
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) GetUserNameByUserId(req *proto.GetUserInfoRequest) (code int, userName string) {
	reply := &proto.GetUserInfoResponse{}
	LogicRpcClient.Call(context.Background(), "GetUserInfoByUserId", req, reply)
	code = reply.Code
	userName = reply.UserName
	return
}

func (rpc *RpcLogic) CheckAuth(req *proto.CheckAuthRequest) (code int, userId int, userName string) {
	reply := &proto.CheckAuthResponse{}
	LogicRpcClient.Call(context.Background(), "CheckAuth", req, reply)
	code = reply.Code
	userId = reply.UserId
	userName = reply.UserName
	return
}

func (rpc *RpcLogic) Logout(req *proto.LogoutRequest) (code int) {
	reply := &proto.LogoutResponse{}
	LogicRpcClient.Call(context.Background(), "Logout", req, reply)
	code = reply.Code
	return
}
