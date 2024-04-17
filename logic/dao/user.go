package dao

import (
	"errors"
	"gochat/db"
	"time"
)

var dbIns = db.GetDb()

type User struct {
	Id         int `gorm:"primary_key"`
	UserName   string
	Password   string
	CreateTime time.Time
	db.DbGoChat
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) Add() (userId int, err error) {
	if u.UserName == "" || u.Password == "" {
		return 0, errors.New("user_name or password empty")
	}
	u.CreateTime = time.Now()
	if err = dbIns.Create(&u).Error; err != nil {
		return 0, err
	}
	return u.Id, nil
}

// TODO CheckHaveUserName 可以使用 bool 过滤器，减轻数据库的压力

func (u *User) CheckHaveUserName(userName string) (data User) {
	dbIns.Where("user_name=?", userName).Take(&data)
	return
}

func (u *User) GetUserNameByUserId(userId int) (userName string) {
	var data User
	dbIns.Where("id=?", userId).Take(&data)
	return data.UserName
}
