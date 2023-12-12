package rpc

import (
	"context"
	"github.com/lty/my-go-chat/proto"
)

// 发送单点消息
func (rpc *RpcLogic) Push(req *proto.Send) (code int, msg string) {
	reply := &proto.SuccessReply{} // TODO: 这是什么
	LogicRpcClient.Call(context.Background(), "Push", req, reply)
	code = reply.Code
	msg = reply.Msg
	return
}

// 向房间发送消息
func (rpc *RpcLogic) PushRoom(req *proto.Send) (code int, msg string) {
	reply := &proto.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "PushRoom", req, reply)
	code = reply.Code
	msg = reply.Msg
	return
}

// TODO: 这个到底是干嘛的
func (rpc *RpcLogic) Count(req *proto.Send) (code int, msg string) {
	reply := &proto.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "Count", req, reply)
	code = reply.Code
	msg = reply.Msg
	return
}

// TODO: 这个到底是干嘛的
func (rpc *RpcLogic) GetRoomInfo(req *proto.Send) (code int, msg string) {
	reply := &proto.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "GetRoomInfo", req, reply)
	code = reply.Code
	msg = reply.Msg
	return
}
