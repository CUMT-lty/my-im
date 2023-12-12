package main

import (
	"flag"
	"fmt"
	"github.com/lty/my-go-chat/api"
	"github.com/lty/my-go-chat/connect"
	"github.com/lty/my-go-chat/logic"
	"github.com/lty/my-go-chat/task"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var module string
	flag.StringVar(&module, "module", "", "assign run module")
	flag.Parse()
	fmt.Println(fmt.Sprintf("start run %s module", module))
	switch module {
	case "logic":
		logic.New().Run() // 1. 启动 logic 层
	case "connect_websocket":
		connect.New().Run() // 2. 启动 connect 层
	case "task":
		task.New().Run() // 3. 启动 task 层
	case "api":
		api.New().Run() // 4. 启动 api 层
	default:
		fmt.Println("exiting, module param error!")
		return
	}
	fmt.Println(fmt.Sprintf("run %s module done!", module))
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("Server exiting")
}
