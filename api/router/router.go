package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lty/my-go-chat/api/handler"
	"github.com/lty/my-go-chat/utils"
)

// TODO: 注册所有的路由，这个方法还没写完
func Register() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware()) // TODO: 全局中间件？
	initUserRouter(r)       // /user 路由组
	initPushRouter(r)       // /push 路由组
	r.NoRoute(func(c *gin.Context) {
		utils.FailWithMsg(c, "please check request url !")
	})
	return r
}

func initUserRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.POST("/login", handler.Login)       // ✅
	userGroup.POST("/register", handler.Register) // ✅
	// TODO: 中间件的作用范围，是不是只有下面两个路由才走中间件
	userGroup.Use(CheckSessionId())
	{
		userGroup.POST("/checkAuth", handler.CheckAuth) // ✅
		userGroup.POST("/logout", handler.Logout)       // ✅
	}

}

func initPushRouter(r *gin.Engine) {
	pushGroup := r.Group("/push")
	pushGroup.Use(CheckSessionId()) // TODO: 中间件的作用对象
	// 这些路由走了 CheckSessionId 这个中间件，请求 json 中要加上 authToken
	{
		pushGroup.POST("/push", handler.Push)               // ✅
		pushGroup.POST("/pushRoom", handler.PushRoom)       // ✅
		pushGroup.POST("/count", handler.Count)             // ✅
		pushGroup.POST("/getRoomInfo", handler.GetRoomInfo) // ✅
	}
	// TODO: 是否需要在这些路由处理函数中将 http 连接升级为 websocket 连接，否则如何接收消息
}
