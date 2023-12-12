package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lty/my-go-chat/api/dto"
	"github.com/lty/my-go-chat/api/rpc"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
	"strconv"
)

// 最上层的消息推送，只负责推送，不负责读取
func Push(c *gin.Context) {
	var formPush dto.FormPush
	if err := c.ShouldBindBodyWith(&formPush, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	authToken := formPush.AuthToken
	msg := formPush.Msg
	toUserId := formPush.ToUserId
	toUserIdInt, _ := strconv.Atoi(toUserId)
	getUserNameReq := &proto.GetUserInfoRequest{UserId: toUserIdInt}
	code, toUserName := rpc.RpcLogicObj.GetUserNameByUserId(getUserNameReq) // 先找消息接收方用户是否存在
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "rpc fail get friend userName")
		return
	}
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: authToken} // 验证消息发送方的登陆状态
	code, fromUserId, fromUserName := rpc.RpcLogicObj.CheckAuth(checkAuthReq)
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "rpc fail get self info")
		return
	}
	// TODO: 这个业务逻辑是什么，是只能私发给同一房间内的人吗
	roomId := formPush.RoomId
	req := &proto.Send{
		Msg:          msg,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		ToUserId:     toUserIdInt,
		ToUserName:   toUserName,
		RoomId:       roomId,
		Op:           config.OpSingleSend, // 操作类型：单点发送
	}
	code, rpcMsg := rpc.RpcLogicObj.Push(req) // 向下一层发起 rpc 调用
	if code == utils.CodeFail {
		utils.FailWithMsg(c, rpcMsg)
		return
	}
	utils.SuccessWithMsg(c, "ok", nil)
	return
}

func PushRoom(c *gin.Context) {
	var formRoom dto.FormRoom
	if err := c.ShouldBindBodyWith(&formRoom, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	authToken := formRoom.AuthToken
	msg := formRoom.Msg
	roomId := formRoom.RoomId
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: authToken}
	authCode, fromUserId, fromUserName := rpc.RpcLogicObj.CheckAuth(checkAuthReq) // 验证发送方登陆状态
	if authCode == utils.CodeFail {
		utils.FailWithMsg(c, "rpc fail get self info")
		return
	}
	req := &proto.Send{
		Msg:          msg,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		RoomId:       roomId,
		Op:           config.OpRoomSend, // 操作类型：发送到指定房间
	}
	code, msg := rpc.RpcLogicObj.PushRoom(req) // 向下一层发起 rpc 调用
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "rpc push room msg fail!")
		return
	}
	utils.SuccessWithMsg(c, "ok", msg)
	return
}

func Count(c *gin.Context) {
	var formCount dto.FormCount
	if err := c.ShouldBindBodyWith(&formCount, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	roomId := formCount.RoomId
	req := &proto.Send{
		RoomId: roomId,
		Op:     config.OpRoomCountSend, // TODO: 获取房间人数？
	}
	code, msg := rpc.RpcLogicObj.Count(req) // 向下一层发起 rpc 调用
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "rpc get room count fail!")
		return
	}
	utils.SuccessWithMsg(c, "ok", msg)
	return
}

func GetRoomInfo(c *gin.Context) {
	var formRoomInfo dto.FormRoomInfo
	if err := c.ShouldBindBodyWith(&formRoomInfo, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	roomId := formRoomInfo.RoomId
	req := &proto.Send{
		RoomId: roomId,
		Op:     config.OpRoomInfoSend, // TODO: 这个业务逻辑是干嘛的，没懂
	}
	code, msg := rpc.RpcLogicObj.GetRoomInfo(req) // 向下一层发起 rpc 调用
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "rpc get room info fail!")
		return
	}
	utils.SuccessWithMsg(c, "ok", msg)
	return
}
