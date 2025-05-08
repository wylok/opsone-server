package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/monitor"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func MonitorGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/monitor")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.GET("/rule", monitor.QueryRule)
		v1.POST("/rule", monitor.CreateRule)
		v1.PUT("/rule", monitor.ModifyRule)
		v1.DELETE("/rule", monitor.DeleteRule)
		v1.GET("/data/converge", monitor.ConvergeData)
		v1.GET("/data/detail", monitor.DetailData)
		v1.GET("/metric", monitor.Metrics)
		v1.GET("/metric/cn", monitor.MetricCn)
		v1.GET("/stages", monitor.QueryStages)
		v1.POST("/stages", monitor.CreateStages)
		v1.GET("/alarm/history", monitor.AlarmHistory)
		v1.DELETE("/alarm/history", monitor.DeleteAlarmHistory)
		v1.GET("/alarm/send", monitor.AlarmSend)
		v1.DELETE("/alarm/send", monitor.DeleteAlarmSend)
		v1.GET("/alarm/count", monitor.AlarmCount)
		v1.GET("/rule/groups", monitor.QueryRuleGroups)
		v1.POST("/rule/groups", monitor.RelationRuleGroups)
		v1.POST("/alarm/pause", monitor.AlarmPause)
		v1.GET("/alarm/pause", monitor.QueryAlarmPause)
		v1.GET("/metric/top", monitor.MetricTop)
		v1.GET("/process", monitor.QueryProcess)
		v1.DELETE("/process", monitor.DeleteProcess)
		v1.POST("/process", monitor.AddProcess)
		v1.GET("/group/process", monitor.QueryGroupProcess)
		v1.POST("/group/process", monitor.AddGroupProcess)
		v1.DELETE("/group/process", monitor.DeleteGroupProcess)
		v1.GET("/job", monitor.QueryMonitorJobs)
		v1.POST("/job", monitor.CreateMonitorJob)
		v1.DELETE("/job", monitor.DeleteMonitorJobs)
		v1.GET("/custom/metric", monitor.QueryCustom)
		v1.GET("/custom/group/metric", monitor.QueryGroupCustom)
		v1.POST("/custom/metric", monitor.ConfigCustom)
		v1.DELETE("/custom/metric", monitor.DeleteCustom)
		v1.PUT("/custom/script", monitor.RefreshScript)
		v1.GET("/process/top", monitor.QueryProcessTop)
		v1.GET("/rule/contacts", monitor.QueryRuleContacts)
	}
}

func init() {
	platform_conf.RouteNames["monitor.QueryRule"] = "查看报警规则"
	platform_conf.RouteNames["monitor.CreateRule"] = "新建报警规则"
	platform_conf.RouteNames["monitor.ModifyRule"] = "修改报警规则"
	platform_conf.RouteNames["monitor.DeleteRule"] = "删除报警规则"
	platform_conf.RouteNames["monitor.ConvergeData"] = "监控数据聚合"
	platform_conf.RouteNames["monitor.DetailData"] = "查看监控详情"
	platform_conf.RouteNames["monitor.Metrics"] = "查看监控指标"
	platform_conf.RouteNames["monitor.MetricCn"] = "查看监控指标中文"
	platform_conf.RouteNames["monitor.QueryStages"] = "查看报警步骤"
	platform_conf.RouteNames["monitor.CreateStages"] = "新建报警步骤"
	platform_conf.RouteNames["monitor.AlarmHistory"] = "查看报警记录"
	platform_conf.RouteNames["monitor.AlarmSend"] = "查看报警发送记录"
	platform_conf.RouteNames["monitor.DeleteAlarmSend"] = "删除报警发送记录"
	platform_conf.RouteNames["monitor.AddProcess"] = "新增进程监控"
	platform_conf.RouteNames["monitor.QueryGroupProcess"] = "查看资源组进程监控"
	platform_conf.RouteNames["monitor.AddGroupProcess"] = "新增资源组进程监控"
	platform_conf.RouteNames["monitor.DeleteGroupProcess"] = "删除资源组进程监控"
	platform_conf.RouteNames["monitor.QueryMonitorJobs"] = "查看故障自愈"
	platform_conf.RouteNames["monitor.CreateMonitorJob"] = "新建故障自愈"
	platform_conf.RouteNames["monitor.DeleteMonitorJobs"] = "删除故障自愈"
	platform_conf.RouteNames["monitor.QueryCustom"] = "查看自定义指标"
	platform_conf.RouteNames["monitor.QueryGroupCustom"] = "查看资源组自定义指标"
	platform_conf.RouteNames["monitor.ConfigCustom"] = "修改自定义指标"
	platform_conf.RouteNames["monitor.DeleteCustom"] = "删除自定义指标"
	platform_conf.RouteNames["monitor.RefreshScript"] = "重新分发指标收集器"
	platform_conf.RouteNames["monitor.QueryRuleGroups"] = "查询规则关联资源组"
	platform_conf.RouteNames["monitor.AlarmCount"] = "查询报警统计"
	platform_conf.RouteNames["monitor.RelationRuleGroups"] = "规则关联资源组"
	platform_conf.RouteNames["monitor.MetricTop"] = "查询监控指标TOP"
	platform_conf.RouteNames["monitor.DeleteAlarmHistory"] = "删除报警记录"
	platform_conf.RouteNames["monitor.QueryProcess"] = "查询进程监控"
	platform_conf.RouteNames["monitor.AlarmPause"] = "暂停监控通知"
	platform_conf.RouteNames["monitor.DeleteProcess"] = "删除进程监控"
	platform_conf.RouteNames["monitor.QueryAlarmPause"] = "查询暂停监控通知"
	platform_conf.RouteNames["monitor.QueryProcessTop"] = "查询进程TOP"
}
