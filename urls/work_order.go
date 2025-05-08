package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/work_order"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func WorkOrderGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/work_order")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.GET("", work_order.QueryWorkOrder)
		v1.POST("", work_order.AddWorkOrder)
		v1.PUT("", work_order.ModifyWorkOrder)
		v1.DELETE("", work_order.DelWorkOrder)
		v1.GET("/approve/ready", work_order.ReadyApproveWorkOrder)
		v1.GET("/approve/pend", work_order.PendApproveWorkOrder)
		v1.POST("/approve", work_order.ApproveWorkOrder)
		v1.GET("/approve/flow", work_order.QueryApproveFlow)
		v1.GET("/flow", work_order.QueryWorkOrderFlow)
		v1.POST("/flow", work_order.AddWorkOrderFLow)
		v1.PUT("/flow", work_order.ModifyWorkOrderFLow)
		v1.DELETE("/flow", work_order.DelWorkOrderFLow)
		v1.GET("/type", work_order.QueryWorkOrderType)
	}
}

func init() {
	platform_conf.RouteNames["work_order.QueryWorkOrder"] = "查看工单列表"
	platform_conf.RouteNames["work_order.AddWorkOrder"] = "新建工单"
	platform_conf.RouteNames["work_order.ModifyWorkOrder"] = "修改工单"
	platform_conf.RouteNames["work_order.DelWorkOrder"] = "删除工单"
	platform_conf.RouteNames["work_order.ReadyApproveWorkOrder"] = "查看已审批工单"
	platform_conf.RouteNames["work_order.PendApproveWorkOrder"] = "查看待审批工单"
	platform_conf.RouteNames["work_order.QueryWorkOrderFlow"] = "查看审批流程"
	platform_conf.RouteNames["work_order.AddWorkOrderFLow"] = "新建审批流程"
	platform_conf.RouteNames["work_order.ModifyWorkOrderFLow"] = "修改审批流程"
	platform_conf.RouteNames["work_order.DelWorkOrderFLow"] = "删除审批流程"
	platform_conf.RouteNames["work_order.QueryWorkOrderType"] = "查询工单类型"
	platform_conf.RouteNames["work_order.ApproveWorkOrder"] = "审批工单"
	platform_conf.RouteNames["work_order.QueryApproveFlow"] = "查询审批进度"
}
