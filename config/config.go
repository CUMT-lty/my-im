package config

import (
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
	"sync"
)

var once sync.Once

// var realPath string

var Conf *Config // TODO: 全局唯一配置对象，只被实例化一次

const (
	SuccessReplyCode      = 0
	FailReplyCode         = 1
	SuccessReplyMsg       = "success"
	QueueName             = "gochat_queue"
	RedisBaseValidTime    = 86400 // 24 小时
	RedisPrefix           = "gochat_"
	RedisRoomPrefix       = "gochat_room_"
	RedisRoomOnlinePrefix = "gochat_room_online_count_"
	MsgVersion            = 1 // TODO: 这个是干嘛的
	OpSingleSend          = 2 // single user
	OpRoomSend            = 3 // send to room
	OpRoomCountSend       = 4 // get online user count
	OpRoomInfoSend        = 5 // send info to room
	//	OpBuildTcpConn        = 6 // build tcp conn
)

type Config struct {
	Common  Common // 通用配置
	Connect ConnectConfig
	Logic   LogicConfig
	Task    TaskConfig
	Api     ApiConfig // API 层的配置
	//Site    SiteConfig
}

type Common struct {
	CommonEtcd  CommonEtcd  `mapstructure:"common-etcd"`
	CommonRedis CommonRedis `mapstructure:"common-redis"`
}

// TODO: etcd 服务发现相关配置
type CommonEtcd struct {
	Host              string `mapstructure:"host"`
	BasePath          string `mapstructure:"basePath"`
	ServerPathLogic   string `mapstructure:"serverPathLogic"`
	ServerPathConnect string `mapstructure:"serverPathConnect"`
	UserName          string `mapstructure:"userName"`
	Password          string `mapstructure:"password"`
	ConnectionTimeout int    `mapstructure:"connectionTimeout"`
}

// TODO: 拿到和 gin 模式相关的环境变量
func GetMode() string {
	env := os.Getenv("RUN_MODE")
	if env == "" {
		env = "dev"
	}
	return env
}

// TODO: gin 模式
func GetGinRunMode() string {
	env := GetMode()
	//gin have debug,test,release mode
	if env == "dev" {
		return "debug"
	}
	if env == "test" {
		return "debug"
	}
	if env == "prod" {
		return "release"
	}
	return "release"
}

// TODO: redis 相关配置
type CommonRedis struct {
	RedisAddress  string `mapstructure:"redisAddress"` // // TODO: mapstructure用于将通用的 map[string]interface{} 解码到对应的 Go 结构体中，或者执行相反的操作。
	RedisPassword string `mapstructure:"redisPassword"`
	Db            int    `mapstructure:"db"`
}

type ConnectBase struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}

type ConnectRpcAddressWebsockts struct {
	Address string `mapstructure:"address"`
}

//	type ConnectRpcAddressTcp struct {
//		Address string `mapstructure:"address"`
//	}
type ConnectBucket struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	Channel       int    `mapstructure:"channel"`
	Room          int    `mapstructure:"room"`
	SrvProto      int    `mapstructure:"svrProto"`
	RoutineAmount uint64 `mapstructure:"routineAmount"`
	RoutineSize   int    `mapstructure:"routineSize"`
}

type ConnectWebsocket struct {
	ServerId string `mapstructure:"serverId"`
	Bind     string `mapstructure:"bind"`
}

//	type ConnectTcp struct {
//		ServerId      string `mapstructure:"serverId"`
//		Bind          string `mapstructure:"bind"`
//		SendBuf       int    `mapstructure:"sendbuf"`
//		ReceiveBuf    int    `mapstructure:"receivebuf"`
//		KeepAlive     bool   `mapstructure:"keepalive"`
//		Reader        int    `mapstructure:"reader"`
//		ReadBuf       int    `mapstructure:"readBuf"`
//		ReadBufSize   int    `mapstructure:"readBufSize"`
//		Writer        int    `mapstructure:"writer"`
//		WriterBuf     int    `mapstructure:"writerBuf"`
//		WriterBufSize int    `mapstructure:"writeBufSize"`
//	}
type ConnectConfig struct {
	ConnectBase                ConnectBase                `mapstructure:"connect-base"`
	ConnectRpcAddressWebSockts ConnectRpcAddressWebsockts `mapstructure:"connect-rpcAddress-websockts"`
	//ConnectRpcAddressTcp       ConnectRpcAddressTcp       `mapstructure:"connect-rpcAddress-tcp"`
	ConnectBucket    ConnectBucket    `mapstructure:"connect-bucket"`
	ConnectWebsocket ConnectWebsocket `mapstructure:"connect-websocket"`
	//ConnectTcp                 ConnectTcp                 `mapstructure:"connect-tcp"`
}
type LogicBase struct {
	//ServerId   string `mapstructure:"serverId"`
	CpuNum     int    `mapstructure:"cpuNum"`
	RpcAddress string `mapstructure:"rpcAddress"`
	CertPath   string `mapstructure:"certPath"`
	KeyPath    string `mapstructure:"keyPath"`
}
type LogicConfig struct {
	LogicBase LogicBase `mapstructure:"logic-base"`
}

type TaskBase struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	RedisAddr     string `mapstructure:"redisAddr"`
	RedisPassword string `mapstructure:"redisPassword"`
	RpcAddress    string `mapstructure:"rpcAddress"`
	PushChan      int    `mapstructure:"pushChan"`
	PushChanSize  int    `mapstructure:"pushChanSize"`
}

type TaskConfig struct {
	TaskBase TaskBase `mapstructure:"task-base"`
}

// api 层配置
type ApiBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

// api 层配置
type ApiConfig struct {
	ApiBase ApiBase `mapstructure:"api-base"`
}

//type SiteBase struct {
//	ListenPort int `mapstructure:"listenPort"`
//}
//
//type SiteConfig struct {
//	SiteBase SiteBase `mapstructure:"site-base"`
//}

func getCurrentDir() string {
	_, fileName, _, _ := runtime.Caller(1)
	aPath := strings.Split(fileName, "/")
	dir := strings.Join(aPath[0:len(aPath)-1], "/")
	return dir
}

// TODO: init 函数在导包时自动执行
func init() {
	Init()
}

func Init() {
	once.Do(func() {
		env := GetMode()
		//realPath, _ := filepath.Abs("./")
		realPath := getCurrentDir()
		configFilePath := realPath + "/" + env + "/"
		viper.SetConfigType("toml")
		viper.SetConfigName("/connect")
		viper.AddConfigPath(configFilePath)
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/common")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/task")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/logic")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/api")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		viper.SetConfigName("/site")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}
		Conf = new(Config)
		viper.Unmarshal(&Conf.Common)
		viper.Unmarshal(&Conf.Connect)
		viper.Unmarshal(&Conf.Task)
		viper.Unmarshal(&Conf.Logic)
		viper.Unmarshal(&Conf.Api)
		//viper.Unmarshal(&Conf.Site)
	})
}
