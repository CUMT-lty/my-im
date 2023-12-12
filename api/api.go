package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lty/my-go-chat/api/router"
	"github.com/lty/my-go-chat/api/rpc"
	"github.com/lty/my-go-chat/config"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 本层入口，对外暴露服务

type Api struct {
}

func New() *Api {
	return &Api{}
}

func (api *Api) Run() {
	rpc.InitLogicRpcClient()          // 初始化本层的 rpc 客户端
	r := router.Register()            // 路由处理器
	runMode := config.GetGinRunMode() // TODO: gin 工作模式
	logrus.Info("server start , now run mode is ", runMode)
	gin.SetMode(runMode)
	apiConfig := config.Conf.Api
	port := apiConfig.ApiBase.ListenPort // 端口 7070
	fmt.Println("api层端口：", port)
	//flag.Parse()                         // TODO: 这里是在干嘛

	// 开启 http 服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r, // TODO: 这里使用了 r
	}
	go func() { // TODO: 这里启了一个新的 goroutine
		// TODO: 为什么不用 gin 的 r.Run，是否是为了兼容 tcp 连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { // 启动监听，
			logrus.Errorf("start listen : %s\n", err)
		}
	}()

	// TODO: 优雅关闭服务
	quit := make(chan os.Signal)
	// TODO: go 中的信号处理：
	// SIGUP: 终端控制进程结束
	// SIGINT: 用户发送 INTR 字符 (Ctrl+C) 触发
	// SIGTERM: 程序结束，可以被捕获、阻塞或忽略
	// SIGQUIT: 用户发送 QUIT 字符 (Ctrl+/) 触发
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-quit // TODO: 从这里开始 Run() 方法的执行会被阻塞，直到有信号进入 quit 通道
	// TODO: 要知道这些信号量是怎么被读到的
	logrus.Infof("Exit signal:", s)

	logrus.Infof("Shut down server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server Shutdown:", err)
	}
	logrus.Infof("Server exiting")
	os.Exit(0) // TODO: 退出的是进程
	// TODO: 这里好麻烦，为什么不用 gin 的优雅关停
}
