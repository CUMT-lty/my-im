package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lty/my-go-chat/api/handler"
	"github.com/lty/my-go-chat/utils"
)

func Register() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware()) // 全局中间件
	initUserRouter(r)       // /user 路由组
	initPushRouter(r)       // /push 路由组
	r.NoRoute(func(c *gin.Context) {
		utils.FailWithMsg(c, "please check request url !")
	})
	return r
}

func initUserRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)
	// TODO: 注意中间件的作用范围
	userGroup.Use(CheckSessionId())
	{
		userGroup.POST("/checkAuth", handler.CheckAuth)
		userGroup.POST("/logout", handler.Logout)
	}

}

func initPushRouter(r *gin.Engine) {
	pushGroup := r.Group("/push")
	pushGroup.Use(CheckSessionId()) // TODO: 中间件的作用对象
	// 这些路由走了 CheckSessionId 这个中间件，请求 json 中要加上 authToken
	{
		pushGroup.POST("/push", handler.Push)
		pushGroup.POST("/pushRoom", handler.PushRoom)
		pushGroup.POST("/count", handler.Count)
		pushGroup.POST("/getRoomInfo", handler.GetRoomInfo)
	}
}
