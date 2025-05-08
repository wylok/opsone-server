package databases

import (
	"time"
)

type Msg struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	MsgId      string    `gorm:"column:msg_id;type:varchar(100);uniqueIndex" json:"msg_id"`
	MsgType    string    `gorm:"column:msg_type;type:varchar(100)" json:"msg_type"`
	Level      string    `gorm:"column:level;type:varchar(50)" json:"level"`
	Title      string    `gorm:"column:title;type:varchar(100);index" json:"title"`
	Status     string    `gorm:"column:status;type:varchar(100);index" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (Msg) TableName() string {
	return "msg"
}

type MsgContent struct {
	Id      uint64 `gorm:"primary_key" json:"id"`
	MsgId   string `gorm:"column:msg_id;type:varchar(100);uniqueIndex" json:"msg_id"`
	Content string `gorm:"column:content;type:varchar(1000)" json:"content"`
}

func (MsgContent) TableName() string {
	return "msg_content"
}
