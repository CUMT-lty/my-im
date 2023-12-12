package logic

import (
	"context"
	"fmt"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/utils"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"strings"
	"time"
)

func (logic *Logic) InitPublishRedisClient() (err error) {
	redisOpt := utils.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	RedisClient = utils.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping(context.Background()).Result(); err != nil { // TODO: ping 中的参数要怎么传
		logrus.Infof("RedisCli Ping Result pong: %s,  err: %s", pong, err)
	}
	// this can change use another redis save session data
	RedisSessClient = RedisClient // TODO: 管理 session 也用这个连接
	return err
}

func (logic *Logic) InitRpcServer() (err error) {
	var network, addr string
	rpcAddrList := strings.Split(config.Conf.Logic.LogicBase.RpcAddress, ",")
	for _, bind := range rpcAddrList {
		if network, addr, err = utils.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogicRpc ParseNetwork error : %s", err.Error())
		}
		logrus.Infof("logic start run at-->%s:%s", network, addr)
		go logic.createRpcServer(network, addr) // TODO: 启了一个新 goroutine，内部注册并启动一个 logic 层服务结点
	}
	return
}

func (logic *Logic) createRpcServer(network string, addr string) {
	s := server.NewServer() // TODO: 配置文件中每有一个结点就创建一个新的 server
	// 添加 etcd 插件
	logic.addRegistryPlugin(s, network, addr) // TODO: 任何插件都必须在注册服务之前添加到 server 中
	// serverId must be unique
	//err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", config.Conf.Logic.LogicBase.ServerId))
	// 注册 rpc 服务，TODO: 最后一个 metadata 参数应该是用来标识该服务结点是哪一个服务器的
	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(LogicRpcServer), fmt.Sprintf("%s", logic.ServerId))
	if err != nil {
		logrus.Errorf("register error:%s", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) { // 优雅关闭服务
		s.UnregisterAll()
	})
	s.Serve(network, addr) // 启动服务，监听 rpc 请求
}

func (logic *Logic) addRegistryPlugin(s *server.Server, network string, addr string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,                         // 对外暴露的本机监听地址
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host}, // TODO: etcd 集群地址
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,       // 服务前缀，为当前服务设置命名空间
		Metrics:        metrics.NewRegistry(),                        // 用来更新服务的tps
		UpdateInterval: time.Minute,                                  // 服务的刷新间隔
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}
