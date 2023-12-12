package dto

// 用户登陆
type FormLogin struct {
	// TODO: struct tag
	// struct tag 是结构体在编译阶段关联到成员的元信息字符串，在运行的时候通过反射的机制读取出来。
	// 由一个或多个键值对组成；键与值使用冒号分隔，值用双引号括起来；键值对之间使用一个空格分隔。
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
	// TODO: json tag
	// json tag 为键名的标签对应的值用于控制 encoding/json 包的编码和解码的行为
	// 并且 encoding/...下面其它的包也遵循这个约定。
	// TODO: form tag
	// form tag 为 Gin 中使用的 tag，将表单数据和模型进行绑定，方便参数校验和使用。
	// TODO: binding tag
	// Gin 对于数据的校验使用的是 validator.v10 包进行参数验证，该包提供多种数据校验方法，通过 binding tag 来进行数据校验。
}

// 用户注册
type FormRegister struct {
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
}

// 验证登陆状态
type FormCheckAuth struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

// 用户登出
type FormLogout struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

// 登陆状态校验
type FormCheckSessionId struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}
