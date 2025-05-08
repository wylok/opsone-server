package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/job"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func JobGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/job")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.POST("/exec", job.ExecUpdate)
		v1.GET("/exec", job.ExecList)
		v1.POST("/file/upload", job.FileUpdate)
		v1.POST("/file/send", job.FileSend)
		v1.GET("/file", job.FileList)
		v1.GET("/results", job.Results)
		v1.GET("/overview", job.Overview)
		v1.DELETE("/overview", job.OverviewDelete)
		v1.POST("/script/upload", job.ScriptUpdate)
		v1.GET("/script", job.ScriptList)
		v1.GET("/script/detail", job.ScriptDetail)
		v1.DELETE("/script", job.ScriptDelete)
		v1.POST("/script/run", job.ScriptRun)
		v1.GET("/script/run", job.ScriptRunList)
		v1.PUT("/script", job.ScriptModify)
	}
}

func init() {
	platform_conf.RouteNames["job.ExecUpdate"] = "命令执行"
	platform_conf.RouteNames["job.ExecList"] = "命令执行列表"
	platform_conf.RouteNames["job.FileUpdate"] = "文件上传"
	platform_conf.RouteNames["job.FileSend"] = "执行文件分发"
	platform_conf.RouteNames["job.FileList"] = "文件分发列表"
	platform_conf.RouteNames["job.Results"] = "查看作业结果"
	platform_conf.RouteNames["job.Overview"] = "查看作业总览"
	platform_conf.RouteNames["job.OverviewDelete"] = "删除作业"
	platform_conf.RouteNames["job.ScriptUpdate"] = "脚本上传"
	platform_conf.RouteNames["job.ScriptList"] = "查看脚本列表"
	platform_conf.RouteNames["job.ScriptDetail"] = "查看脚本详情"
	platform_conf.RouteNames["job.ScriptRun"] = "执行脚本"
	platform_conf.RouteNames["job.ScriptDelete"] = "删除脚本"
	platform_conf.RouteNames["job.ScriptRunList"] = "脚本执行列表"
	platform_conf.RouteNames["job.ScriptModify"] = "修改脚本"
}
