package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/logic/dao"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// TODO: 提供对上层的 rpc 服务

type LogicRpcServer struct{} // TODO: 这个类型的名字就是服务的名字

// rpc_user
func (rpc *LogicRpcServer) Register(ctx context.Context, args *proto.RegisterRequest, reply *proto.RegisterReply) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	uData := u.CheckHaveUserName(args.Name) // 不重复注册
	if uData.Id > 0 {
		return errors.New("this user name already have , please login !!!")
	}
	u.UserName = args.Name
	u.Password = args.Password
	userId, err := u.AddUser()
	if err != nil {
		logrus.Infof("register err:%s", err.Error())
		return err
	}
	if userId == 0 {
		return errors.New("register userId empty!")
	}
	// TODO: 设置 token
	randToken := utils.GetRandomToken(32)         // 生成随机 token
	sessionId := utils.CreateSessionId(randToken) // 加上了特殊的 session 前缀，标识这是用作 session 的
	userData := make(map[string]interface{})
	userData["userId"] = userId
	userData["userName"] = args.Name
	// redis 处理 session
	RedisSessClient.Do(ctx, "MULTI") // TODO: 开启 redis 事务
	// TODO: redis 中数据结构的选取
	RedisSessClient.HMSet(ctx, sessionId, userData)           // 使用哈希类型，sessionId 作为 key
	RedisSessClient.Expire(ctx, sessionId, 86400*time.Second) // 给 session key 设置有效时间为一天，一天之内不登陆就会取消登陆状态
	err = RedisSessClient.Do(ctx, "EXEC").Err()               // TODO: 执行事务
	if err != nil {
		logrus.Infof("register set redis token fail!")
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}

func (rpc *LogicRpcServer) Login(ctx context.Context, args *proto.LoginRequest, reply *proto.LoginResponse) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	userName := args.Name
	passWord := args.Password
	data := u.CheckHaveUserName(userName)
	if (data.Id == 0) || (passWord != data.Password) {
		return errors.New("no this user or password error!")
	}
	loginSessionId := utils.GetSessionIdByUserId(data.Id)
	// set token
	// err = redis.HMSet(auth, userData)
	randToken := utils.GetRandomToken(32)         // 生成随机 token
	sessionId := utils.CreateSessionId(randToken) // 获得 session
	userData := make(map[string]interface{})
	userData["userId"] = data.Id
	userData["userName"] = data.UserName
	// 检查登陆状态
	token, _ := RedisSessClient.Get(ctx, loginSessionId).Result()
	if token != "" { // 如果已经登陆
		//logout already login user session
		oldSession := utils.CreateSessionId(token)
		err := RedisSessClient.Del(ctx, oldSession).Err() // 删除旧 session
		if err != nil {
			return errors.New("logout user fail!token is:" + token)
		}
	}
	RedisSessClient.Do(ctx, "MULTI")                                       // 开启事务
	RedisSessClient.HMSet(ctx, sessionId, userData)                        // 加入新 session
	RedisSessClient.Expire(ctx, sessionId, 86400*time.Second)              // 设置过期时间为一天
	RedisSessClient.Set(ctx, loginSessionId, randToken, 86400*time.Second) // TODO: 没明白
	err = RedisSessClient.Do(ctx, "EXEC").Err()                            // 执行事务
	//err = RedisSessClient.Set(authToken, data.Id, 86400*time.Second).Err()
	if err != nil {
		logrus.Infof("register set redis token fail!")
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}

func (rpc *LogicRpcServer) GetUserInfoByUserId(ctx context.Context, args *proto.GetUserInfoRequest, reply *proto.GetUserInfoResponse) (err error) {
	reply.Code = config.FailReplyCode
	userId := args.UserId
	u := new(dao.User)
	userName := u.GetUserNameByUserId(userId)
	reply.UserId = userId
	reply.UserName = userName
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *LogicRpcServer) CheckAuth(ctx context.Context, args *proto.CheckAuthRequest, reply *proto.CheckAuthResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := utils.GetSessionName(authToken)
	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(ctx, sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail!,authToken is:%s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session,authToken is:%s", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	reply.UserId = intUserId
	userName, _ := userDataMap["userName"]
	reply.Code = config.SuccessReplyCode
	reply.UserName = userName
	return
}

func (rpc *LogicRpcServer) Logout(ctx context.Context, args *proto.LogoutRequest, reply *proto.LogoutResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := utils.GetSessionName(authToken)

	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(ctx, sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail!,authToken is:%s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session,authToken is:%s", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	sessIdMap := utils.GetSessionIdByUserId(intUserId)
	//del sess_map like sess_map_1
	err = RedisSessClient.Del(ctx, sessIdMap).Err() // 删除登陆状态，删除相应的 key
	if err != nil {
		logrus.Infof("logout del sess map error:%s", err.Error())
		return err
	}
	// TODO: 没明白这里 serverId 是干嘛的，为什么要删除
	logic := new(Logic)
	serverIdKey := logic.getUserKey(fmt.Sprintf("%d", intUserId))
	err = RedisSessClient.Del(ctx, serverIdKey).Err()
	if err != nil {
		logrus.Infof("logout del server id error:%s", err.Error())
		return err
	}
	err = RedisSessClient.Del(ctx, sessionName).Err()
	if err != nil {
		logrus.Infof("logout error:%s", err.Error())
		return err
	}
	reply.Code = config.SuccessReplyCode
	return
}

// rpc push

// 发送单点消息
func (rpc *LogicRpcServer) Push(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	var bodyBytes []byte
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic --> push msg fail,err:%s", err.Error())
		return
	}
	logic := new(Logic)
	userSidKey := logic.getUserKey(fmt.Sprintf("%d", sendData.ToUserId))
	serverIdStr := RedisSessClient.Get(ctx, userSidKey).Val()
	if err != nil {
		logrus.Errorf("logic --> push parse int fail:%s", err.Error())
		return
	}
	err = logic.RedisPublishChannel(ctx, serverIdStr, sendData.ToUserId, bodyBytes)
	if err != nil {
		logrus.Errorf("logic --> redis publish err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// 发送消息到房间
func (rpc *LogicRpcServer) PushRoom(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	roomId := sendData.RoomId
	logic := new(Logic)
	roomUserInfo := make(map[string]string)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisClient.HGetAll(ctx, roomUserKey).Result()
	if err != nil {
		logrus.Errorf("logic,PushRoom redis hGetAll err:%s", err.Error())
		return
	}
	//if len(roomUserInfo) == 0 {
	//	return errors.New("no this user")
	//}
	var bodyBytes []byte
	sendData.RoomId = roomId
	sendData.Msg = args.Msg
	sendData.FromUserId = args.FromUserId
	sendData.FromUserName = args.FromUserName
	sendData.Op = config.OpRoomSend
	sendData.CreateTime = utils.GetNowDateTime()
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic,PushRoom Marshal err:%s", err.Error())
		return
	}
	err = logic.RedisPublishRoomInfo(ctx, roomId, len(roomUserInfo), roomUserInfo, bodyBytes)
	if err != nil {
		logrus.Errorf("logic,PushRoom err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// 获取一个房间中的在线用户数量
func (rpc *LogicRpcServer) Count(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	roomId := args.RoomId
	logic := new(Logic)
	var count int
	count, err = RedisSessClient.Get(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", roomId))).Int()
	err = logic.RedisPushRoomCount(ctx, roomId, count)
	if err != nil {
		logrus.Errorf("logic,Count err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// 获取房间相关信息
func (rpc *LogicRpcServer) GetRoomInfo(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisClient.HGetAll(ctx, roomUserKey).Result()
	if len(roomUserInfo) == 0 {
		return errors.New("getRoomInfo no this user")
	}
	err = logic.RedisPushRoomInfo(ctx, roomId, len(roomUserInfo), roomUserInfo)
	if err != nil {
		logrus.Errorf("logic,GetRoomInfo err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// TODO: 连接管理

// TODO： 连接，没看明白逻辑
func (rpc *LogicRpcServer) Connect(ctx context.Context, args *proto.ConnectRequest, reply *proto.ConnectReply) (err error) {
	if args == nil {
		logrus.Errorf("logic --> connect args empty")
		return
	}
	logic := new(Logic)
	// key := logic.getUserKey(args.AuthToken)
	logrus.Infof("logic,authToken is:%s", args.AuthToken)
	key := utils.GetSessionName(args.AuthToken)
	userInfo, err := RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		logrus.Infof("RedisCli HGetAll key :%s , err:%s", key, err.Error())
		return err
	}
	if len(userInfo) == 0 {
		reply.UserId = 0
		return
	}
	reply.UserId, _ = strconv.Atoi(userInfo["userId"])
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(args.RoomId))
	if reply.UserId != 0 {
		userKey := logic.getUserKey(fmt.Sprintf("%d", reply.UserId))
		logrus.Infof("logic redis set userKey:%s, serverId : %s", userKey, args.ServerId)
		validTime := config.RedisBaseValidTime * time.Second
		err = RedisClient.Set(ctx, userKey, args.ServerId, validTime).Err() // TODO: 有效时间
		if err != nil {
			logrus.Warnf("logic set err:%s", err)
		}
		if RedisClient.HGet(ctx, roomUserKey, fmt.Sprintf("%d", reply.UserId)).Val() == "" {
			RedisClient.HSet(ctx, roomUserKey, fmt.Sprintf("%d", reply.UserId), userInfo["userName"])
			// add room user count ++
			RedisClient.Incr(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId)))
		}
	}
	logrus.Infof("logic rpc userId:%d", reply.UserId)
	return
}

// TODO: 断开连接, 没看明白逻辑
func (rpc *LogicRpcServer) DisConnect(ctx context.Context, args *proto.DisConnectRequest, reply *proto.DisConnectReply) (err error) {
	logic := new(Logic)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(args.RoomId))
	// room user count --
	if args.RoomId > 0 {
		count, _ := RedisSessClient.Get(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Int()
		if count > 0 {
			RedisClient.Decr(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Result()
		}
	}
	// room login user--
	if args.UserId != 0 {
		err = RedisClient.HDel(ctx, roomUserKey, fmt.Sprintf("%d", args.UserId)).Err()
		if err != nil {
			logrus.Warnf("HDel getRoomUserKey err : %s", err)
		}
	}
	//below code can optimize send a signal to queue,another process get a signal from queue,then push event to websocket
	roomUserInfo, err := RedisClient.HGetAll(ctx, roomUserKey).Result()
	if err != nil {
		logrus.Warnf("RedisCli HGetAll roomUserInfo key:%s, err: %s", roomUserKey, err)
	}
	if err = logic.RedisPublishRoomInfo(ctx, args.RoomId, len(roomUserInfo), roomUserInfo, nil); err != nil {
		logrus.Warnf("publish RedisPublishRoomCount err: %s", err.Error())
		return
	}
	return
}
