package task

import (
	"encoding/json"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/sirupsen/logrus"
	"math/rand"
)

// 推送消息的相关参数
type PushParams struct {
	ServerId string
	UserId   int
	Msg      []byte
	RoomId   int
}

// TODO: 推送通道，是一个通道列表
var pushChannels []chan *PushParams

// 初始化通道列表
func init() {
	pushChannels = make([]chan *PushParams, config.Conf.Task.TaskBase.PushChan)
}

func (task *Task) GoPush() {
	for i := 0; i < len(pushChannels); i++ {
		pushChannels[i] = make(chan *PushParams, config.Conf.Task.TaskBase.PushChanSize)
		go task.processSinglePush(pushChannels[i]) // 每个通道都开启一个线程去阻塞读
	}
}

func (task *Task) processSinglePush(ch chan *PushParams) {
	var arg *PushParams
	for {
		arg = <-ch // TODO: 阻塞读，通道中的 PushParams 有个字段 serverId 为空
		// TODO: 消息漏读问题如下:
		// TODO: when arg.ServerId server is down, user could be reconnect other serverId but msg in queue no consume
		task.pushSingleToConnect(arg.ServerId, arg.UserId, arg.Msg)
	}
}

func (task *Task) Push(msg string) {
	m := &proto.RedisMsg{}
	if err := json.Unmarshal([]byte(msg), m); err != nil {
		logrus.Infof(" json.Unmarshal err:%v ", err)
	}
	logrus.Infof("push msg info %d,op is:%d", m.RoomId, m.Op)
	switch m.Op {
	case config.OpSingleSend: // 单点发送消息
		pushChannels[rand.Int()%config.Conf.Task.TaskBase.PushChan] <- &PushParams{ // 写入通道
			ServerId: m.ServerId,
			UserId:   m.UserId,
			Msg:      m.Msg,
		}
	case config.OpRoomSend: // 向房间发送消息
		task.broadcastRoomToConnect(m.RoomId, m.Msg)
	case config.OpRoomCountSend:
		task.broadcastRoomCountToConnect(m.RoomId, m.Count)
	case config.OpRoomInfoSend: // 获取房间相关信息
		task.broadcastRoomInfoToConnect(m.RoomId, m.RoomUserInfo)
	}
}
