package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lty/my-go-chat/config"
	"github.com/sirupsen/logrus"
	"runtime"
)

// Logic 中的 ServerId 标识这个服务器
// Logic 也是本层中可以用来操作消息队列的类型
type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	//read config
	logicConfig := config.Conf.Logic

	runtime.GOMAXPROCS(logicConfig.LogicBase.CpuNum)
	// 从这里开始 logic 层的 run 逻辑
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())
	// 连接 redis
	if err := logic.InitPublishRedisClient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail,err:%s", err.Error())
	}
	// 启动 rpc 服务
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init rpc server fail")
	}
}
