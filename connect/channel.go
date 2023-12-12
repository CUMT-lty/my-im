package connect

import (
	"github.com/gorilla/websocket"
	"github.com/lty/my-go-chat/proto"
)

// in fact, Channel it's a user Connect session

// 一个 Channel 表示一个用户会话连接
type Channel struct {
	Room      *Room    // TODO: 所属房间？
	Next      *Channel // 一个房间中的所有会话连接用双向链表组织
	Prev      *Channel
	broadcast chan *proto.Msg // TODO: 用来广播消息的通道？
	userId    int             // 用户 id，标识是哪个用户的会话链接
	conn      *websocket.Conn // websocket 连接
	//connTcp   *net.TCPConn  // 不用 tcp 连接
}

// 创建新会话对象，参数消息通道的缓存容量
func NewChannel(size int) (c *Channel) {
	c = new(Channel)
	c.broadcast = make(chan *proto.Msg, size)
	c.Next = nil
	c.Prev = nil
	return
}

func (ch *Channel) Push(msg *proto.Msg) (err error) { // 向房间内投送消息
	select {
	case ch.broadcast <- msg: // 将消息放入通道
	default:
	}
	return
}
