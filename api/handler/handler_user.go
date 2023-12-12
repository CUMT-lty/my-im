package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lty/my-go-chat/api/dto"
	"github.com/lty/my-go-chat/api/rpc"
	"github.com/lty/my-go-chat/proto"
	"github.com/lty/my-go-chat/utils"
)

func Login(c *gin.Context) {
	var formLogin dto.FormLogin
	if err := c.ShouldBindBodyWith(&formLogin, binding.JSON); err != nil { // 失败
		utils.FailWithMsg(c, err.Error()) // 这里面有响应
		return
	}
	req := &proto.LoginRequest{
		Name:     formLogin.UserName,
		Password: utils.Sha1(formLogin.Password), // 密码要哈希处理
	}
	// TODO: RpcLogicObj 是 rpc 请求代理对象？本服务中发出的所有 rpc 请求都交给这个对象处理
	code, authToken, msg := rpc.RpcLogicObj.Login(req) // TODO: 向下一层发送 rpc 请求，调用 Login
	if code == utils.CodeFail || authToken == "" {     // 失败
		utils.FailWithMsg(c, msg)
		return
	}
	// 成功
	utils.SuccessWithMsg(c, "login success", authToken) // TODO: 登陆功能的最上层返回结果
}

func Register(c *gin.Context) {
	var formRegister dto.FormRegister
	// TODO; 请求是form表单还是 json 字符串
	if err := c.ShouldBindBodyWith(&formRegister, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	req := &proto.RegisterRequest{
		Name:     formRegister.UserName,
		Password: utils.Sha1(formRegister.Password),
	}
	code, authToken, msg := rpc.RpcLogicObj.Register(req) // TODO: 向下一层发送 rpc 请求，调用 Register
	if code == utils.CodeFail || authToken == "" {
		utils.FailWithMsg(c, msg)
		return
	}
	utils.SuccessWithMsg(c, "register success", authToken)
}

func CheckAuth(c *gin.Context) {
	var formCheckAuth dto.FormCheckAuth
	if err := c.ShouldBindBodyWith(&formCheckAuth, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	authToken := formCheckAuth.AuthToken
	req := &proto.CheckAuthRequest{
		AuthToken: authToken,
	}
	code, userId, userName := rpc.RpcLogicObj.CheckAuth(req) // TODO: 向下一层发送 rpc 请求，调用 CheckAuth
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "auth fail")
		return
	}
	var jsonData = map[string]interface{}{
		"userId":   userId,
		"userName": userName,
	}
	utils.SuccessWithMsg(c, "auth success", jsonData)
}

func Logout(c *gin.Context) {
	var formLogout dto.FormLogout
	if err := c.ShouldBindBodyWith(&formLogout, binding.JSON); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	authToken := formLogout.AuthToken
	logoutReq := &proto.LogoutRequest{
		AuthToken: authToken,
	}
	code := rpc.RpcLogicObj.Logout(logoutReq) // TODO: 向下一层发送 rpc 请求，调用 Logout
	if code == utils.CodeFail {
		utils.FailWithMsg(c, "logout fail!")
		return
	}
	utils.SuccessWithMsg(c, "logout ok!", nil)
}
