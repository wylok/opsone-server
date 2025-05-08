package monitor

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/monitor_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

// @Tags 监控平台
// @Summary 监控指标查询
// @Produce  json
// @Security ApiKeyAuth
// @Param resource query string true "监控资源"
// @Param items query array true "监控项"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/metric [get]
func Metrics(c *gin.Context) {
	//监控指标查询接口
	var (
		JsonData       monitor_conf.Metric
		Response       = common.Response{C: c}
		MonitorMetrics []databases.MonitorMetrics
		Metrics        []string
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		JsonData.Items = kits.FormListFormat(JsonData.Items)
		db.Select("monitor_key").Where("monitor_resource=? and monitor_item in ?",
			JsonData.Resource, JsonData.Items).Find(&MonitorMetrics)
		if len(MonitorMetrics) > 0 {
			for _, v := range MonitorMetrics {
				Metrics = append(Metrics, v.MonitorKey)
			}
		}
		Response.Data = Metrics
	}
}

// @Tags 监控平台
// @Summary 监控指标中文查询
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/metric/cn [get]
func MetricCn(c *gin.Context) {
	//监控指标中文查询接口
	var (
		Response    = common.Response{C: c}
		MonitorKeys []databases.MonitorKeys
		keys        = map[string]map[string]string{}
	)
	// 接口请求返回结果
	defer func() {
		Response.Send()
	}()
	db.Find(&MonitorKeys)
	if len(MonitorKeys) > 0 {
		for _, v := range MonitorKeys {
			keys[v.MonitorKey] = map[string]string{"KeyCn": v.MonitorKeyCn, "KeyUnit": v.MonitorKeyUnit}
		}
	}
	Response.Data = keys
}
