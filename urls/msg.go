package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/msg"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func MsgGroup(r *gin.Engine) {
	v2 := r.Group("/api/v1/msg")
	{
		v2.Use(middleware.VerifyToken())
		v2.Use(middleware.VerifyPermission())
		v2.Use(middleware.Audit())
		v2.GET("", msg.QueryMsg)
		v2.GET("/detail", msg.Detail)
		v2.DELETE("", msg.DeleteMsg)
	}
}

func init() {
	platform_conf.RouteNames["msg.QueryMsg"] = "查看消息"
	platform_conf.RouteNames["msg.Detail"] = "查看消息详情"
	platform_conf.RouteNames["msg.DeleteMsg"] = "删除消息"
}
