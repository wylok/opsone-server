package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/platform"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func PlatformGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/heartbeat")
	{
		v1.GET("/ws/", platform.Heartbeat)
	}
	v2 := r.Group("/api/v1/platform")
	{
		v2.Use(middleware.VerifyToken())
		v2.Use(middleware.VerifyPermission())
		v2.Use(middleware.Audit())
		v2.GET("/agent/conf", platform.AgentConfig)
		v2.GET("/overview", platform.Overview)
		v2.GET("/offline_time", platform.OfflineTime)
		v2.PUT("/agent/conf", platform.ModifyAgent)
		v2.GET("/platform/config", platform.Config)
		v2.GET("/agent/alive", platform.AgentAlive)
		v2.DELETE("/agent/alive", platform.DeleteAgentAlive)
	}
}

func init() {
	platform_conf.RouteNames["platform.AgentConfig"] = "查询Agent配置"
	platform_conf.RouteNames["platform.Overview"] = "查看平台总览"
	platform_conf.RouteNames["platform.OfflineTime"] = "查询设备离线"
	platform_conf.RouteNames["platform.ModifyAgent"] = "修改Agent配置"
	platform_conf.RouteNames["platform.Config"] = "查看平台配置"
	platform_conf.RouteNames["platform.AgentAlive"] = "查看Agent在线"
	platform_conf.RouteNames["platform.DeleteAgentAlive"] = "删除离线Agent"
}
