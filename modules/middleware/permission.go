package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"net/http"
)

var (
	Log      kits.Log
	db       = databases.DB
	rc, ctx  = common.RedisConnect()
	enforcer = kits.CasBin()
)

func VerifyPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				log.Error(err)
			}
		}()
		userId := c.GetString("user_id")
		token := c.GetString("token")
		err = enforcer.LoadPolicy()
		if err == nil && userId != "" && token != "" {
			if userId == "platform" && token == platform_conf.PublicToken {
				c.Next()
				return
			}
			allowed, err := enforcer.Enforce(userId, platform_conf.TenantId, c.Request.URL.Path, c.Request.Method)
			if err == nil && allowed {
				c.Next()
				return
			}
		}
		log.Error(" user_id:" + userId + " Permission deny")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false, "message": "Forbidden", "data": map[string]interface{}{}})
		c.Abort()
		return
	}
}
