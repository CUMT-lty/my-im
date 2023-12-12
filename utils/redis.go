package utils

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// redis 初始化相关

var RedisClientMap = map[string]*redis.Client{}
var syncLock sync.Mutex

type RedisOption struct {
	Address  string
	Password string
	Db       int
}

func GetRedisInstance(redisOpt RedisOption) *redis.Client {
	address := redisOpt.Address
	db := redisOpt.Db
	password := redisOpt.Password
	addr := fmt.Sprintf("%s", address)
	syncLock.Lock()
	// 防止并发请求来时，给一个 redis 服务地址创建多个连接
	if redisCli, ok := RedisClientMap[addr]; ok { // 一个 redis 服务地址使用一个 redis 连接
		// 这里返回之前是否需要释放锁
		return redisCli // 如果已有就直接返回
	}
	newRedisCli := redis.NewClient(&redis.Options{ // 否则要新建连接
		Addr:            addr,
		Password:        password,
		DB:              db,
		ConnMaxLifetime: 20 * time.Second,
		//MaxConnAge: 20 * time.Second, 原来是这个，含义应该是一样的
	})
	RedisClientMap[addr] = newRedisCli
	syncLock.Unlock()
	return RedisClientMap[addr]
}
