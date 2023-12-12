package connect

import "github.com/lty/my-go-chat/proto"

type Operator interface {
	Connect(conn *proto.ConnectRequest) (int, error)
	DisConnect(disConn *proto.DisConnectRequest) (err error)
}

type DefaultOperator struct { // TODO: 要对这个注册服务吗
}

// rpc call logic layer
func (o *DefaultOperator) Connect(conn *proto.ConnectRequest) (uid int, err error) {
	rpcConnect := new(RpcConnect)
	uid, err = rpcConnect.Connect(conn)
	return
}

// rpc call logic layer
func (o *DefaultOperator) DisConnect(disConn *proto.DisConnectRequest) (err error) {
	rpcConnect := new(RpcConnect)
	err = rpcConnect.DisConnect(disConn)
	return
}
