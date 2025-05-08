package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/cloud"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func CloudGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/cloud")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.GET("/oss", cloud.QueryOss)
		v1.GET("/key", cloud.QueryCloudKey)
		v1.POST("/key", cloud.AddCloudKey)
		v1.DELETE("/key", cloud.DelCloudKey)
		v1.GET("/servers", cloud.QueryCloudServer)
		v1.POST("/server/operate", cloud.OperateCloudServer)
	}
}

func init() {
	platform_conf.RouteNames["cloud.QueryOss"] = "查看OSS"
	platform_conf.RouteNames["cloud.QueryCloudKey"] = "查看密钥"
	platform_conf.RouteNames["cloud.AddCloudKey"] = "新增密钥"
	platform_conf.RouteNames["cloud.DelCloudKey"] = "删除密钥"
	platform_conf.RouteNames["cloud.QueryCloudServer"] = "查询云主机"
	platform_conf.RouteNames["cloud.OperateCloudServer"] = "操作云主机"
}
