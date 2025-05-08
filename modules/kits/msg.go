package kits

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"inner/conf/msg_conf"
	"inner/modules/databases"
	"time"
)

func RecordMsg(JsonData msg_conf.RMsg) bool {
	//消息记录
	var (
		err    error
		sqlErr error
		db     = databases.DB
		now    = time.Now()
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
	}()
	if err == nil {
		// 初始化数据库连接
		MsgId := RandString(12)
		// 写入表数据
		err = db.Transaction(func(tx *gorm.DB) error {
			m := databases.Msg{MsgId: MsgId, MsgType: JsonData.MsgType, Title: JsonData.Title,
				Level:  JsonData.Level,
				Status: "Unread", CreateTime: now}
			if err = tx.Create(&m).Error; err != nil {
				sqlErr = err
			}
			// 写入表数据
			mc := databases.MsgContent{MsgId: MsgId, Content: JsonData.Content}
			if err = tx.Create(&mc).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
	if err != nil {
		return false
	}
	return true
}
