package connect

import (
	"context"
	"errors"
	"fmt"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"strings"
	"sync"
	"time"
)

// logic 层的 rpc 客户端，此实例只会被初始化一次
// logic 层和 redis 交互，服务写在 logic 是因为还需要 redis 来辅助存储一些信息
var logicRpcClient client.XClient
var once sync.Once

// 初始化 logicRpcClient 实例，该实例只会被初始化一次
func (c *Connect) InitLogicRpcClient() (err error) {
	etcdConfigOption := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
		Bucket:            "",
		PersistConnection: true,
		Username:          config.Conf.Common.CommonEtcd.UserName,
		Password:          config.Conf.Common.CommonEtcd.Password,
	}
	once.Do(func() { // 只执行一次
		etcdDisc, e := etcdV3.NewEtcdV3Discovery( // etcd 服务发现
			config.Conf.Common.CommonEtcd.BasePath,
			config.Conf.Common.CommonEtcd.ServerPathLogic,
			[]string{config.Conf.Common.CommonEtcd.Host},
			true,
			etcdConfigOption,
		)
		if e != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail:%s", e.Error())
		}
		// TODO: logic 层 rpc 服务对应的客户端对象初始化，etcd 服务发现
		logicRpcClient = client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathLogic, client.Failtry, client.RandomSelect, etcdDisc, client.DefaultOption)
	})
	if logicRpcClient == nil {
		return errors.New("get rpc client nil")
	}
	return
}

// Connect 层向其他层发起 rpc 请求的代理对象
type RpcConnect struct {
}

func (rpc *RpcConnect) Connect(connReq *proto.ConnectRequest) (uid int, err error) {
	reply := &proto.ConnectReply{}
	err = logicRpcClient.Call(context.Background(), "Connect", connReq, reply) // 向 logic 层的 rpc 远程调用
	if err != nil {
		logrus.Fatalf("failed to call: %v", err)
	}
	uid = reply.UserId
	logrus.Infof("connect logic userId :%d", reply.UserId)
	return
}

func (rpc *RpcConnect) DisConnect(disConnReq *proto.DisConnectRequest) (err error) {
	reply := &proto.DisConnectReply{}
	if err = logicRpcClient.Call(context.Background(), "DisConnect", disConnReq, reply); err != nil {
		logrus.Fatalf("failed to call: %v", err)
	} // 向 logic 层的 rpc 远程调用
	return
}

// connect 层 rpc 服务初始化
func (c *Connect) InitConnectWebsocketRpcServer() (err error) {
	var network, addr string
	connectRpcAddress := strings.Split(config.Conf.Connect.ConnectRpcAddressWebSockts.Address, ",")
	for _, bind := range connectRpcAddress {
		if network, addr, err = utils.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer ParseNetwork error : %s", err)
		}
		logrus.Infof("Connect rpc server start run at-->%s:%s", network, addr)
		go c.createConnectWebsocktsRpcServer(network, addr) // TODO: 每一个配置都开启一个独立的 connect 层 rpc 服务 goroutine
	}
	return
}

// 开启一个 rpc 服务，连接类型是 websocket
func (c *Connect) createConnectWebsocktsRpcServer(network string, addr string) {
	s := server.NewServer()
	addRegistryPlugin(s, network, addr) // TODO: 任何插件都要在服务注册之前就添加进去
	//config.Conf.Connect.ConnectTcp.ServerId
	//s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathConnect, new(ConnectPushRpcServer), fmt.Sprintf("%s", config.Conf.Connect.ConnectWebsocket.ServerId))
	// metadata 参数中有服务器 serveId 和 连接类型 serverType=ws
	s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathConnect, new(ConnectPushRpcServer), fmt.Sprintf("serverId=%s&serverType=ws", c.ServerId))
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr) // 启动服务
}

// 给 rpc 服务添加插件
func addRegistryPlugin(s *server.Server, network string, addr string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

// 本层提供的 rpc 服务

// connect 层的服务注册对象
type ConnectPushRpcServer struct {
}

// 单点发送消息
func (rpc *ConnectPushRpcServer) PushSingleMsg(ctx context.Context, pushMsgReq *proto.PushMsgRequest, successReply *proto.SuccessReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
	)
	logrus.Info("rpc PushMsg :%v ", pushMsgReq)
	if pushMsgReq == nil {
		logrus.Errorf("rpc PushSingleMsg() args:(%v)", pushMsgReq)
		return
	}
	bucket = DefaultServer.Bucket(pushMsgReq.UserId)                 // 获取桶
	if channel = bucket.Channel(pushMsgReq.UserId); channel != nil { // 获取用户会话
		err = channel.Push(&pushMsgReq.Msg) // 发送消息
		logrus.Infof("DefaultServer Channel err nil ,args: %v", pushMsgReq)
		return
	}
	successReply.Code = config.SuccessReplyCode
	successReply.Msg = config.SuccessReplyMsg
	logrus.Infof("successReply:%v", successReply)
	return
}

// 发送消息到房间
func (rpc *ConnectPushRpcServer) PushRoomMsg(ctx context.Context, pushRoomMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Code = config.SuccessReplyCode
	successReply.Msg = config.SuccessReplyMsg
	logrus.Infof("PushRoomMsg msg %+v", pushRoomMsgReq)
	for _, bucket := range DefaultServer.Buckets { // 遍历桶，因为本质上桶管理的是会话连接，广播到房间的最终目的是发送给房间内的每一个用户
		bucket.BroadcastRoom(pushRoomMsgReq) // 发送消息到房间
	}
	return
}

// TODO: 没看明白
func (rpc *ConnectPushRpcServer) PushRoomCount(ctx context.Context, pushRoomMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Code = config.SuccessReplyCode
	successReply.Msg = config.SuccessReplyMsg
	logrus.Infof("PushRoomCount msg %v", pushRoomMsgReq)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(pushRoomMsgReq)
	}
	return
}

// TODO: 没看明白
func (rpc *ConnectPushRpcServer) PushRoomInfo(ctx context.Context, pushRoomMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Code = config.SuccessReplyCode
	successReply.Msg = config.SuccessReplyMsg
	logrus.Infof("connect,PushRoomInfo msg %+v", pushRoomMsgReq)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(pushRoomMsgReq)
	}
	return
}
