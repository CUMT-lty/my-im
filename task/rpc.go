package task

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"strings"
	"sync"
	"time"
)

var RClient = &RpcConnectClient{ // TODO：每个 task 层结点维护一个自己的 RClient
	ServerInsMap: make(map[string][]Instance), // TODO: connect 层每个服务器的客户端连接映射
	IndexMap:     make(map[string]int),
}

type RpcConnectClient struct {
	lock         sync.Mutex
	ServerInsMap map[string][]Instance //serverId--[]ins
	IndexMap     map[string]int        //serverId--index
}

type Instance struct {
	ServerType string
	ServerId   string         // TODO: connect 层服务器的唯一 Id
	Client     client.XClient // TODO: 连接 connect 层该服务器的 rpc 客户端
}

// 初始化 connect 层 rpc 服务对应的客户端
func (task *Task) InitConnectRpcClient() (err error) {
	etcdConfigOption := &store.Config{ // etcd 配置对象
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
		Bucket:            "",
		PersistConnection: true,
		Username:          config.Conf.Common.CommonEtcd.UserName,
		Password:          config.Conf.Common.CommonEtcd.Password,
	}
	etcdConfig := config.Conf.Common.CommonEtcd
	etcdDisc, e := etcdV3.NewEtcdV3Discovery( // etcd 服务发现对象，用于从 etcd 中获取 rpc 服务器的地址信息
		etcdConfig.BasePath,
		etcdConfig.ServerPathConnect, // 写在配置文件中了
		[]string{etcdConfig.Host},
		true,
		etcdConfigOption,
	)
	if e != nil {
		logrus.Fatalf("init task rpc etcd discovery client fail:%s", e.Error())
	}
	if len(etcdDisc.GetServices()) <= 0 {
		logrus.Panicf("no etcd server find!")
	}
	for _, connectConf := range etcdDisc.GetServices() { // TODO: 遍历从 etcd 中获取到的 rpc 服务器的连接配置信息
		logrus.Infof("key is:%s,value is:%s", connectConf.Key, connectConf.Value)
		// 从每个地址信息
		serverType := getParamByKey(connectConf.Value, "serverType")          // 获取服务器类型
		serverId := getParamByKey(connectConf.Value, "serverId")              // 获取服务器 id
		logrus.Infof("serverType is:%s,serverId is:%s", serverType, serverId) // 这里有打印的信息，可以看到变量的具体值
		if serverType == "" || serverId == "" {
			continue
		}
		p2pDisc, e := client.NewPeer2PeerDiscovery(connectConf.Key, "")
		if e != nil {
			logrus.Errorf("init task client.NewPeer2PeerDiscovery client fail:%s", e.Error())
			continue
		}
		c := client.NewXClient(etcdConfig.ServerPathConnect, client.Failtry, client.RandomSelect, p2pDisc, client.DefaultOption)
		ins := Instance{
			ServerType: serverType,
			ServerId:   serverId,
			Client:     c,
		}
		if _, has := RClient.ServerInsMap[serverId]; !has { // 将该连接实例加入对应的 serverId 映射中
			RClient.ServerInsMap[serverId] = []Instance{ins}
		} else {
			RClient.ServerInsMap[serverId] = append(RClient.ServerInsMap[serverId], ins)
		}
	}
	// watch connect server change && update RpcConnectClientList
	go task.watchServicesChange(etcdDisc)
	return
}

/*
允许根据服务器ID获取对应的RPC客户端，以便进行后续的RPC调用操作
通过维护 ServerInsMap 和 IndexMap 两个映射关系，实现了根据 服务器ID 查找 RPC 客户端的功能
TODO: 在多个连接层服务器提供相同服务的情况下，通过轮询选择一个 RPC 客户端，实现了负载均衡的效果。
*/
func (rc *RpcConnectClient) GetRpcClientByServerId(serverId string) (c client.XClient, err error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, has := rc.ServerInsMap[serverId]; !has || len(rc.ServerInsMap[serverId]) <= 0 { // 如果连接层没有这个 ip
		return nil, errors.New("no connect layer ip:" + serverId)
	}
	if _, has := rc.IndexMap[serverId]; !has {
		rc.IndexMap = map[string]int{
			serverId: 0,
		}
	}
	idx := rc.IndexMap[serverId] % len(rc.ServerInsMap[serverId])
	ins := rc.ServerInsMap[serverId][idx]
	rc.IndexMap[serverId] = (rc.IndexMap[serverId] + 1) % len(rc.ServerInsMap[serverId])
	return ins.Client, nil
}

// 获取所有 connect 层服务器的 rpc 客户端连接
func (rc *RpcConnectClient) GetAllConnectTypeRpcClient() (rpcClientList []client.XClient) {
	for serverId, _ := range rc.ServerInsMap {
		c, err := rc.GetRpcClientByServerId(serverId)
		if err != nil {
			logrus.Infof("GetAllConnectTypeRpcClient err:%s", err.Error())
			continue
		}
		rpcClientList = append(rpcClientList, c)
	}
	return
}

// etcd 监控 connect 层的服务结点变化
func (task *Task) watchServicesChange(d client.ServiceDiscovery) {
	etcdConfig := config.Conf.Common.CommonEtcd
	for kvChan := range d.WatchService() {
		if len(kvChan) <= 0 {
			logrus.Errorf("connect services change, connect alarm, no abailable ip")
		}
		logrus.Infof("connect services change trigger...")
		insMapNew := make(map[string][]Instance)
		for _, kv := range kvChan { // 这里的逻辑和上面是一样的
			logrus.Infof("connect services change,key is:%s,value is:%s", kv.Key, kv.Value)
			serverType := getParamByKey(kv.Value, "serverType")
			serverId := getParamByKey(kv.Value, "serverId")
			logrus.Infof("serverType is:%s,serverId is:%s", serverType, serverId)
			if serverType == "" || serverId == "" {
				continue
			}
			d, e := client.NewPeer2PeerDiscovery(kv.Key, "")
			if e != nil {
				logrus.Errorf("init task client.NewPeer2PeerDiscovery watch client fail:%s", e.Error())
				continue
			}
			c := client.NewXClient(etcdConfig.ServerPathConnect, client.Failtry, client.RandomSelect, d, client.DefaultOption)
			ins := Instance{
				ServerType: serverType,
				ServerId:   serverId,
				Client:     c,
			}
			if _, ok := insMapNew[serverId]; !ok {
				insMapNew[serverId] = []Instance{ins}
			} else {
				insMapNew[serverId] = append(insMapNew[serverId], ins)
			}
		}
		RClient.lock.Lock()
		RClient.ServerInsMap = insMapNew
		RClient.lock.Unlock()
	}
}

// 解析参数
func getParamByKey(s string, key string) string {
	logrus.Infof("%s", s)
	params := strings.Split(s, "&")
	for _, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) == 2 && kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

// 发起 rpc 调用来推送消息
// 单点发送消息
func (task *Task) pushSingleToConnect(serverId string, userId int, msg []byte) {
	logrus.Infof("pushSingleToConnect Body %s", string(msg))
	pushMsgReq := &proto.PushMsgRequest{
		UserId: userId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpSingleSend, // 操作类型：单点发送消息
			SeqId:     utils.GetSnowflakeId(),
			Body:      msg, // 消息体
		},
	}
	reply := &proto.SuccessReply{}
	connectRpc, err := RClient.GetRpcClientByServerId(serverId)
	if err != nil {
		logrus.Infof("get rpc client err : %v", err)
	}
	err = connectRpc.Call(context.Background(), "PushSingleMsg", pushMsgReq, reply) // TODO: 向 connect 层发起 rpc 调用
	if err != nil {
		logrus.Infof("pushSingleToConnect Call err %v", err)
	}
	logrus.Infof("reply %s", reply.Msg)
}

// 广播消息到房间
func (task *Task) broadcastRoomToConnect(roomId int, msg []byte) {
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomSend, // 操作类型：广播消息到房间
			SeqId:     utils.GetSnowflakeId(),
			Body:      msg, // 消息体
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient()
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomToConnect rpc  %v", rpc)
		rpc.Call(context.Background(), "PushRoomMsg", pushRoomMsgReq, reply) // TODO: 向 connect 层发起 rpc 调用
		logrus.Infof("reply %s", reply.Msg)
	}
}

func (task *Task) broadcastRoomCountToConnect(roomId, count int) {
	msg := &proto.RedisRoomCountMsg{
		Count: count,
		Op:    config.OpRoomCountSend,
	}
	var body []byte
	var err error
	if body, err = json.Marshal(msg); err != nil {
		logrus.Warnf("broadcastRoomCountToConnect  json.Marshal err :%s", err.Error())
		return
	}
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomCountSend,
			SeqId:     utils.GetSnowflakeId(), // 生成消息的唯一序列号
			Body:      body,                   // 消息体
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient() // 获取一个客户端
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomCountToConnect rpc  %v", rpc)
		rpc.Call(context.Background(), "PushRoomCount", pushRoomMsgReq, reply) // TODO: 向 connect 层发起 rpc 调用
		logrus.Infof("reply %s", reply.Msg)
	}
}

// 向 connect 层广播房间的用户信息
func (task *Task) broadcastRoomInfoToConnect(roomId int, roomUserInfo map[string]string) {
	msg := &proto.RedisRoomInfo{
		Count:        len(roomUserInfo), // 房间内的用户数量
		Op:           config.OpRoomInfoSend,
		RoomUserInfo: roomUserInfo, // 房间内的用户信息
		RoomId:       roomId,       // 房间 roomId
	}
	var body []byte
	var err error
	if body, err = json.Marshal(msg); err != nil {
		logrus.Warnf("broadcastRoomInfoToConnect  json.Marshal err :%s", err.Error())
		return
	}
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomInfoSend,
			SeqId:     utils.GetSnowflakeId(), // 生成消息的唯一序列号
			Body:      body,                   // 消息体
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient() // 获取一个 rpc 客户端
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomInfoToConnect rpc  %v", rpc)
		rpc.Call(context.Background(), "PushRoomInfo", pushRoomMsgReq, reply) // TODO: 向 connect 层发起 rpc 调用
		logrus.Infof("broadcastRoomInfoToConnect rpc  reply %v", reply)
	}
}
