package task

import (
	"github.com/lty/my-go-chat/config"
	"github.com/sirupsen/logrus"
	"runtime"
)

type Task struct { // 本层的操作代理对象
}

func New() *Task {
	return new(Task)
}

func (task *Task) Run() {
	//read config
	taskConfig := config.Conf.Task
	runtime.GOMAXPROCS(taskConfig.TaskBase.CpuNum)
	//read from redis queue
	if err := task.InitQueueRedisClient(); err != nil {
		logrus.Panicf("task --> init publishRedisClient fail,err:%s", err.Error())
	}
	//rpc call connect layer send msg
	if err := task.InitConnectRpcClient(); err != nil {
		logrus.Panicf("task --> init InitConnectRpcClient fail,err:%s", err.Error())
	}
	//GoPush
	task.GoPush()
}
