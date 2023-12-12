package connect

import (
	"github.com/gorilla/websocket"
	"github.com/lty/my-go-chat/config"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (c *Connect) InitWebsocket() error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c.serveWs(DefaultServer, w, r)
	}) // ✅
	err := http.ListenAndServe(config.Conf.Connect.ConnectWebsocket.Bind, nil)
	return err
}

func (c *Connect) serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  server.Options.ReadBufferSize,
		WriteBufferSize: server.Options.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool { //cross origin domain support
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil) // 升级 http 连接为 websocket 连接
	if err != nil {
		logrus.Errorf("serverWs err:%s", err.Error())
		return
	}
	var ch *Channel
	//default broadcast size eq 512
	ch = NewChannel(server.Options.BroadcastSize)
	ch.conn = conn
	// 每有一个连接进来，都会给这个连接开启独立的读消息 goroutine 和写消息 goroutine
	//send data to websocket conn
	go server.writePump(ch, c)
	//get data from websocket conn
	go server.readPump(ch, c)
}
