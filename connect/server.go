package connect

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lty/my-go-chat/proto"
	"github.com/sirupsen/logrus"
	"github.com/zhenjl/cityhash" // Google提出的 CPU 计算性能较好的一种哈希算法，抗碰撞性能也不错

	"time"
)

// TODO: server 下要分桶，这个分桶机制减少锁竞争还没看明白
type Server struct {
	Buckets   []*Bucket     // TODO: [userId:Bucket]，分桶机制没看明白
	Options   ServerOptions // 配置选项
	bucketIdx uint32        // 这个参数的作用
	operator  Operator
}

type ServerOptions struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

// 返回一个 Server 实例
func NewServer(b []*Bucket, o Operator, options ServerOptions) *Server {
	s := new(Server)
	s.Buckets = b
	s.Options = options
	s.bucketIdx = uint32(len(b))
	s.operator = o
	return s
}

// 通过 userId 获取桶
func (s *Server) Bucket(userId int) *Bucket {
	userIdStr := fmt.Sprintf("%d", userId)
	idx := cityhash.CityHash32([]byte(userIdStr), uint32(len(userIdStr))) % s.bucketIdx // 使用 CityHash 哈希算法
	return s.Buckets[idx]
}

// TODO: 这个方法是干嘛的
func (s *Server) writePump(ch *Channel, c *Connect) {
	//PingPeriod default eq 54s
	// TODO: 发送心跳，保活
	ticker := time.NewTicker(s.Options.PingPeriod)
	defer func() { // 方法退出时，停止发送心跳，并关闭连接
		ticker.Stop()
		ch.conn.Close()
	}()

	for {
		select {
		case message, has := <-ch.broadcast:
			//write data dead time , like http timeout , default 10s
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait)) // TODO: 没看懂
			if !has {
				logrus.Warn("SetWriteDeadline not ok")
				ch.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.conn.NextWriter(websocket.TextMessage) // TODO: 没看懂
			if err != nil {
				logrus.Warn(" ch.conn.NextWriter err :%s  ", err.Error())
				return
			}
			logrus.Infof("message write body:%s", message.Body)
			w.Write(message.Body)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C: // 客户端发来的心跳叫 ping，服务器发送的心跳叫 pong
			//heartbeat，if ping error will exit and close current websocket conn
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait)) // TODO: 没看懂
			logrus.Infof("websocket.PingMessage :%v", websocket.PingMessage)
			// TODO: 这里是不是写错了，服务器发送的心跳应该是 pong
			if err := ch.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// TODO: 读消息
func (s *Server) readPump(ch *Channel, c *Connect) {
	defer func() {
		logrus.Infof("start exec disConnect ...")
		if ch.Room == nil || ch.userId == 0 {
			logrus.Infof("roomId and userId eq 0")
			ch.conn.Close()
			return
		}
		logrus.Infof("exec disConnect ...")
		disConnectRequest := new(proto.DisConnectRequest)
		disConnectRequest.RoomId = ch.Room.Id
		disConnectRequest.UserId = ch.userId
		s.Bucket(ch.userId).DeleteChannel(ch)
		if err := s.operator.DisConnect(disConnectRequest); err != nil {
			logrus.Warnf("DisConnect err :%s", err.Error())
		}
		ch.conn.Close()
	}()

	ch.conn.SetReadLimit(s.Options.MaxMessageSize)              // TODO: 看不懂
	ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait)) // TODO: 看不懂
	ch.conn.SetPongHandler(func(string) error {                 // TODO: 看不懂
		ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait)) // TODO: 看不懂
		return nil
	})

	for {
		_, message, err := ch.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump ReadMessage err:%s", err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		var connReq *proto.ConnectRequest
		logrus.Infof("get a message :%s", message) // 这里可以成功打印消息内容
		if err := json.Unmarshal([]byte(message), &connReq); err != nil {
			logrus.Errorf("message struct %+v", connReq)
		}
		if connReq == nil || connReq.AuthToken == "" {
			logrus.Errorf("s.operator.Connect no authToken")
			return
		}
		connReq.ServerId = c.ServerId //config.Conf.Connect.ConnectWebsocket.ServerId
		userId, err := s.operator.Connect(connReq)
		if err != nil {
			logrus.Errorf("s.operator.Connect error %s", err.Error())
			return
		}
		if userId == 0 {
			logrus.Error("Invalid AuthToken ,userId empty")
			return
		}
		logrus.Infof("websocket rpc call return userId:%d,RoomId:%d", userId, connReq.RoomId)
		b := s.Bucket(userId)
		//insert into a bucket
		err = b.Put(userId, connReq.RoomId, ch)
		if err != nil {
			logrus.Errorf("conn close err: %s", err.Error())
			ch.conn.Close()
		}
	}
}
