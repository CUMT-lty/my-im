package task

import (
	"context"
	"github.com/lty/my-go-chat/config"
	"github.com/lty/my-go-chat/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

// task 层负责操作 redis 消息队列的客户端连接
var RedisClient *redis.Client

// 初始化 redis 客户端连接
func (task *Task) InitQueueRedisClient() (err error) {
	redisOpt := utils.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	RedisClient = utils.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		logrus.Infof("task --> RedisClient Ping Result pong: %s,  err: %s", pong, err)
	}
	go func() { // 不断消费消息并通过 rpc 调用推送到 connect 层
		for {
			var result []string
			//10s timeout
			result, err = RedisClient.BRPop(context.Background(), time.Second*10, config.QueueName).Result() // 阻塞读
			if err != nil {
				logrus.Infof("task queue block timeout,no msg err:%s", err.Error())
			}
			if len(result) >= 2 {
				task.Push(result[1])
			}
		}
	}()
	return
}
