package dao

import (
	"testing"
	"time"
)

func Test_User(t *testing.T) {
	u := User{
		Id:         1,
		UserName:   "test",
		Password:   "123456",
		CreateTime: time.Now(),
	}
	id, err := u.Add()
	if id != 1 || err != nil {
		t.Fail()
	}
	name := u.GetUserNameByUserId(1)
	if name != "test" {
		t.Fail()
	}
}
