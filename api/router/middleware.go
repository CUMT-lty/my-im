package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lty/my-go-chat/api/dto"
	"github.com/lty/my-go-chat/api/rpc"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
	"net/http"
)

// 中间件，检查 session
func CheckSessionId() gin.HandlerFunc {
	return func(c *gin.Context) {
		var formCheckSessionId dto.FormCheckSessionId
		if err := c.ShouldBindBodyWith(&formCheckSessionId, binding.JSON); err != nil {
			c.Abort()
			utils.ResponseWithCode(c, utils.CodeSessionError, nil, nil)
			return
		}
		authToken := formCheckSessionId.AuthToken
		req := &proto.CheckAuthRequest{
			AuthToken: authToken,
		}
		code, userId, userName := rpc.RpcLogicObj.CheckAuth(req) // TODO: 向下一层发送 rpc 请求，调用 CheckAuth
		if code == utils.CodeFail || userId <= 0 || userName == "" {
			c.Abort()
			utils.ResponseWithCode(c, utils.CodeSessionError, nil, nil)
			return
		}
		c.Next() // TODO: Nest 只能在 Gin 的中间件中使用。它会在调用处理程序内部执行链中的待处理程序
		return
	}
}

// TODO: 这个中间件是用来干嘛的？
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}
