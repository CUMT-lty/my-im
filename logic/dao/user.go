package dao

import (
	"github.com/lty/my-go-chat/db"
	"github.com/pkg/errors"
	"time"
)

var dbIns = db.GetDb("gochat")

type User struct {
	Id         int `gorm:"primary_key"`
	UserName   string
	Password   string
	CreateTime time.Time
	//db.DbGoChat // TODO: 这个是啥
}

// 创建 user 表
func (u *User) TableName() string {
	return "user"
}

func (u *User) AddUser() (userId int, err error) {
	if u.UserName == "" || u.Password == "" {
		return 0, errors.New("username or password empty!")
	}
	oUser := u.CheckHaveUserName(u.UserName)
	if oUser.Id > 0 {
		return oUser.Id, nil
	}
	u.CreateTime = time.Now()
	//dbIns.Create(&u)
	if err = dbIns.Table(u.TableName()).Create(&u).Error; err != nil {
		return 0, err
	}
	return u.Id, nil
}

func (u *User) CheckHaveUserName(userName string) (data User) {
	dbIns.Table(u.TableName()).Where("user_name=?", userName).Take(&data)
	return
}

func (u *User) GetUserNameByUserId(userId int) (userName string) {
	var data User
	dbIns.Table(u.TableName()).Where("id=?", userId).Take(&data)
	return data.UserName
}
