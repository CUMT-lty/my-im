package dto

// TODO:
type FormPush struct {
	Msg       string `form:"msg" json:"msg" binding:"required"`
	ToUserId  string `form:"toUserId" json:"toUserId" binding:"required"`
	RoomId    int    `form:"roomId" json:"roomId" binding:"required"`
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

// TODO:
type FormRoom struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
	Msg       string `form:"msg" json:"msg" binding:"required"`
	RoomId    int    `form:"roomId" json:"roomId" binding:"required"`
}

// TODO:
type FormCount struct {
	RoomId int `form:"roomId" json:"roomId" binding:"required"`
}

// TODO:
type FormRoomInfo struct {
	RoomId int `form:"roomId" json:"roomId" binding:"required"`
}
