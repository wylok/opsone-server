package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/monitor_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 监控平台
// @Summary 监控报警查询
// @Produce  json
// @Security ApiKeyAuth
// @Param host_id query string false "主机ID"
// @Param rule_id query string false "规则ID"
// @Param rule_name query string false "规则名称"
// @Param status query string false "报警状态"
// @Param resource query string false "监控资源"
// @Param item query string false "监控项"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/history [get]
func AlarmHistory(c *gin.Context) {
	//监控报警查询接口
	var (
		JsonData     monitor_conf.AlarmHistory
		Response     = common.Response{C: c}
		AlarmHistory []databases.AlarmHistory
		data         []interface{}
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Order("alarm_history.start_time desc")
		if JsonData.HostId != "" {
			tx = tx.Where("alarm_history.host_id=?", JsonData.HostId)
		}
		if JsonData.RuleId != "" {
			tx = tx.Where("alarm_history.rule_id=?", JsonData.RuleId)
		}
		if JsonData.RuleName != "" {
			tx = tx.Where("alarm_history.rule_name like ?", "%"+JsonData.RuleName+"%")
		}
		if JsonData.Status != "" {
			tx = tx.Where("alarm_history.status=?", JsonData.Status)
		}
		if JsonData.Resource != "" {
			tx = tx.Where("alarm_history.monitor_resource=?", JsonData.Resource)
		}
		if JsonData.Item != "" {
			tx = tx.Where("alarm_history.monitor_item=?", JsonData.Item)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AlarmHistory)
		if len(AlarmHistory) > 0 {
			for _, v := range AlarmHistory {
				s, _ := json.Marshal(v)
				d := kits.StringToMap(string(s))
				d["host_name"] = rc.HGet(ctx, platform_conf.ServerNameKey, v.HostId).Val()
				data = append(data, d)
			}
			Response.Data = data
		}
	}
}

// @Tags 监控平台
// @Summary 删除监控报警接口
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteAlarm true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/history [delete]
func DeleteAlarmHistory(c *gin.Context) {
	//删除监控报警接口
	var (
		JsonData     monitor_conf.DeleteAlarm
		Response     = common.Response{C: c}
		AlarmHistory []databases.AlarmHistory
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Ids != nil {
			err = db.Where("id in ?", JsonData.Ids).Delete(&AlarmHistory).Error
		}
		if JsonData.RuleIds != nil {
			err = db.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmHistory).Error
		}
		if JsonData.HostIds != nil {
			err = db.Where("host_id in ??", JsonData.HostIds).Delete(&AlarmHistory).Error
		}
	}
}

// @Tags 监控平台
// @Summary 报警通知查询
// @Produce  json
// @Security ApiKeyAuth
// @Param host_ids query array false "主机ID"
// @Param rule_ids query array false "规则ID"
// @Param channel query string false "报警渠道"
// @Param trace_id query string false "trace_id"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/send [get]
func AlarmSend(c *gin.Context) {
	//报警通知查询接口
	var (
		JsonData     monitor_conf.AlarmSend
		Response     = common.Response{C: c}
		AlarmSend    []databases.AlarmSend
		MonitorRules []databases.MonitorRules
		rules        []string
		ruleNames    = map[string]string{"BuiltInRule": "内置规则"}
		data         []interface{}
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
		JsonData.RuleIds = kits.FormListFormat(JsonData.RuleIds)
		JsonData.HostIds = kits.FormListFormat(JsonData.HostIds)
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Order("alarm_send.send_time desc")
		if JsonData.HostIds != nil {
			tx = tx.Where("alarm_send.host_id in ?", JsonData.HostIds)
		}
		if JsonData.RuleIds != nil {
			tx = tx.Where("alarm_send.rule_id in ?", JsonData.RuleIds)
		}
		if JsonData.Channel != "" {
			tx = tx.Where("alarm_send.channel=?", JsonData.Channel)
		}
		if JsonData.TraceId != "" {
			tx = tx.Where("alarm_send.trace_id=?", JsonData.TraceId)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AlarmSend)
		if len(AlarmSend) > 0 {
			db.Select("rule_id", "rule_name").Where("rule_id in ?", rules).Find(&MonitorRules)
			for _, v := range MonitorRules {
				ruleNames[v.RuleId] = v.RuleName
			}
			for _, v := range AlarmSend {
				s, _ := json.Marshal(v)
				d := kits.StringToMap(string(s))
				d["host_name"] = rc.HGet(ctx, platform_conf.ServerNameKey, v.HostId).Val()
				d["rule_name"] = ruleNames[v.RuleId]
				data = append(data, d)
			}
			Response.Data = data
		}
	}
}

// @Tags 监控平台
// @Summary 删除报警通知接口
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteAlarm true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/send [delete]
func DeleteAlarmSend(c *gin.Context) {
	//删除报警通知接口
	var (
		JsonData  monitor_conf.DeleteAlarm
		Response  = common.Response{C: c}
		AlarmSend []databases.AlarmSend
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Ids != nil {
			err = db.Where("id in ?", JsonData.Ids).Delete(&AlarmSend).Error
		}
		if JsonData.RuleIds != nil {
			err = db.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmSend).Error
		}
		if JsonData.HostIds != nil {
			err = db.Where("host_id in ?", JsonData.HostIds).Delete(&AlarmSend).Error
		}
	}
}

// @Tags 监控平台
// @Summary 报警通知暂停接口
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.PauseAlarm true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/pause [post]
func AlarmPause(c *gin.Context) {
	//报警通知暂停接口
	var (
		JsonData monitor_conf.PauseAlarm
		Response = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		key := monitor_conf.PauseAlarmKey + "_" + JsonData.TraceId
		if JsonData.Duration == 0 {
			JsonData.Duration = 30
		}
		if JsonData.Action == "pause" {
			rc.Set(ctx, key, JsonData.TraceId, time.Duration(JsonData.Duration)*time.Minute)
		}
		if JsonData.Action == "cancel" && rc.Exists(ctx, key).Val() == 1 {
			rc.Del(ctx, key)
		}
	}
}

// @Tags 监控平台
// @Summary 报警通知暂停查询接口
// @Produce  json
// @Security ApiKeyAuth
// @Param trace_ids query array true "trace_id列表"
// @Success 200 {} json "{success:true,message:"ok",data:{}}"
// @Router /api/v1/monitor/alarm/pause [get]
func QueryAlarmPause(c *gin.Context) {
	//报警通知暂停接口
	var (
		JsonData monitor_conf.QueryPauseAlarm
		Response = common.Response{C: c}
		data     = map[string]bool{}
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
		JsonData.TraceIds = kits.FormListFormat(JsonData.TraceIds)
		for _, v := range JsonData.TraceIds {
			data[v] = false
			if rc.Exists(ctx, monitor_conf.PauseAlarmKey+"_"+v).Val() == 1 {
				data[v] = true
			}
		}
	}
	Response.Data = data
}
