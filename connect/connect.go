package connect

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lty/my-go-chat/config"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

type Connect struct { // TODO: ServerId
	ServerId string
}

func New() *Connect {
	return new(Connect)
}

var DefaultServer *Server

func (c *Connect) Run() {
	// get Connect layer config
	connectConfig := config.Conf.Connect
	//set the maximum number of CPUs that can be executing
	runtime.GOMAXPROCS(connectConfig.ConnectBucket.CpuNum)
	//init logic layer rpc client, call logic layer rpc server
	if err := c.InitLogicRpcClient(); err != nil { // 初始化 logic 层 rpc 服务的请求客户端
		logrus.Panicf("connect --> InitLogicRpcClient err: %s", err.Error())
	}
	//init Connect layer rpc server, logic client will call this
	Buckets := make([]*Bucket, connectConfig.ConnectBucket.CpuNum)
	for i := 0; i < connectConfig.ConnectBucket.CpuNum; i++ {
		Buckets[i] = NewBucket(BucketOptions{
			ChannelSize:   connectConfig.ConnectBucket.Channel,
			RoomSize:      connectConfig.ConnectBucket.Room,
			RoutineAmount: connectConfig.ConnectBucket.RoutineAmount,
			RoutineSize:   connectConfig.ConnectBucket.RoutineSize,
		})
	}
	operator := new(DefaultOperator)
	DefaultServer = NewServer(Buckets, operator, ServerOptions{ // 实例化本层 server 对象
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	c.ServerId = fmt.Sprintf("%s-%s", "ws", uuid.New().String()) // 生成 connect 层服务器的唯一id
	//init Connect layer rpc server ,task layer will call this
	if err := c.InitConnectWebsocketRpcServer(); err != nil { // 初始化本层的 rpc 服务
		logrus.Panicf("connect --> InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
	}
	//start Connect layer server handler persistent connection
	if err := c.InitWebsocket(); err != nil { // 初始化 websocket，保持连接
		logrus.Panicf("connect --> Connect layer InitWebsocket() error: %s \n", err.Error())
	}
}
