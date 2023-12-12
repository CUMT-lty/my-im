package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// TODO: 和消息队列的交互

var RedisClient *redis.Client
var RedisSessClient *redis.Client // TODO: 管理 session 的链接

// 发送 redis 消息到队列（这里队列用的是双向链表）
func (logic *Logic) RedisPublishChannel(ctx context.Context, serverId string, toUserId int, msg []byte) (err error) {
	redisMsg := proto.RedisMsg{
		Op:       config.OpSingleSend, // TODO: 单点发送
		ServerId: serverId,            // TODO: logic 层哪个节点发送的？看下这个 serverId 是干嘛的
		UserId:   toUserId,            // TODO: 发给哪个用户的
		Msg:      msg,                 // TODO: 消息
	}
	redisMsgStr, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishChannel Marshal err:%s", err.Error())
		return err
	}
	redisChannel := config.QueueName
	if err := RedisClient.LPush(ctx, redisChannel, redisMsgStr).Err(); err != nil {
		logrus.Errorf("logic,lpush err:%s", err.Error())
		return err
	}
	return
}

func (logic *Logic) RedisPublishRoomInfo(ctx context.Context, roomId int, count int, RoomUserInfo map[string]string, msg []byte) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomSend, // TODO: 发送消息到指定房间
		RoomId:       roomId,            // TODO: 房间编号
		Count:        count,             // TODO: 这个是干嘛的
		Msg:          msg,               // TODO: 消息本体
		RoomUserInfo: RoomUserInfo,      // TODO: 是一个 map，应该是用来遍历房间成员的
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(ctx, config.QueueName, redisMsgByte).Err() // 队列名一样，使用的是同一个队列
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

func (logic *Logic) RedisPushRoomCount(ctx context.Context, roomId int, count int) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:     config.OpRoomCountSend,
		RoomId: roomId, // TODO: 房间 Id
		Count:  count,  // TODO: 房间内的在线用户数量?
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(ctx, config.QueueName, redisMsgByte).Err() // TODO: 在线用户数也要放到队列中吗
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	return
}

// TODO: 不知道这个方法是用来干啥的
func (logic *Logic) RedisPushRoomInfo(ctx context.Context, roomId int, count int, roomUserInfo map[string]string) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomInfoSend,
		RoomId:       roomId,
		Count:        count,
		RoomUserInfo: roomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(ctx, config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

// TODO: authKey 是哪来的
// 加前缀
func (logic *Logic) getRoomUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

// TODO: authKey 是哪来的
// 加前缀
func (logic *Logic) getRoomOnlineCountKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

// TODO: authKey 是哪来的
// 加前缀
func (logic *Logic) getUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
