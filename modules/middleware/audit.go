package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/platform_conf"
	"inner/modules/databases"
	"inner/modules/kits"
	"strings"
	"time"
)

func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			err error
			db  = databases.DB
		)
		go func(c *gin.Context) {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
			}()
			userId := c.GetString("user_id")
			if userId != "" {
				HandlerName := strings.Split(c.HandlerName(), "/")
				name := HandlerName[len(HandlerName)-1]
				_, ok := platform_conf.RouteNames[name]
				if ok && c.Request.Method != "GET" {
					da := databases.Audit{AuditId: kits.RandString(12), UserId: userId, AuditType: "api",
						Handler: name, Action: platform_conf.RouteNames[name], CreateAt: time.Now()}
					err = db.Create(&da).Error
				}
			}
		}(c)
		c.Next()
		return
	}
}
