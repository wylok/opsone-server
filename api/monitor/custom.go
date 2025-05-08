package monitor

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/monitor_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"time"
)

// @Tags 监控平台
// @Summary 配置自定义指标
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.CustomMetric true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/custom/metric [post]
func ConfigCustom(c *gin.Context) {
	//配置自定义指标
	var (
		sqlErr        error
		JsonData      = monitor_conf.CustomMetric{}
		MonitorKeys   []databases.MonitorKeys
		GroupServer   []databases.GroupServer
		CustomMetrics []databases.CustomMetrics
		Response      = common.Response{C: c}
		cf            = platform_conf.Setting()
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		//验证key是否新增
		db.Where("monitor_key=?", JsonData.MonitorKey).First(&MonitorKeys)
		if len(MonitorKeys) == 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				mk := databases.MonitorKeys{MonitorKey: JsonData.MonitorKey, MonitorKeyCn: JsonData.KeyCn,
					MonitorKeyUnit: JsonData.KeyUnit}
				if err = tx.Create(&mk).Error; err != nil {
					sqlErr = err
				}
				mm := databases.MonitorMetrics{MonitorResource: "server", MonitorItem: "custom", MonitorKey: JsonData.MonitorKey}
				if err = tx.Create(&mm).Error; err != nil {
					sqlErr = err
				}
				gm := databases.GroupMetrics{GroupId: JsonData.GroupId, MonitorKey: JsonData.MonitorKey, CreateTime: time.Now()}
				if err = tx.Create(&gm).Error; err != nil {
					sqlErr = err
				}
				cm := databases.CustomMetrics{MonitorKey: JsonData.MonitorKey, ScriptId: JsonData.ScriptId}
				if err = tx.Create(&cm).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			//修改key配置
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Model(&CustomMetrics).Where("monitor_key=?", JsonData.MonitorKey).Updates(
					databases.CustomMetrics{ScriptId: JsonData.ScriptId}).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Model(&MonitorKeys).Where("monitor_key=?", JsonData.MonitorKey).Updates(
					databases.MonitorKeys{MonitorKeyCn: JsonData.KeyCn, MonitorKeyUnit: JsonData.KeyUnit}).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		}
		if err == nil {
			db.Where("group_id=?", JsonData.GroupId).Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, v := range GroupServer {
					_, err = common.RequestApiPost(cf.ApiUrlConfig+cf.JobApiConfig["script_run"].(string),
						platform_conf.PublicToken, map[string]interface{}{"host_id": v.HostId,
							"script_id": JsonData.ScriptId, "dst_path": platform_conf.AgentRoot + "/script/", "cron": false})
				}
			}
		}
	}
}

// @Tags 监控平台
// @Summary 更新指标收集器
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.CustomRefresh true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/custom/script [put]
func RefreshScript(c *gin.Context) {
	//更新指标收集器
	var (
		JsonData      = monitor_conf.CustomRefresh{}
		CustomMetrics []databases.CustomMetrics
		GroupServer   []databases.GroupServer
		Response      = common.Response{C: c}
		cf            = platform_conf.Setting()
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("monitor_key=?", JsonData.MonitorKey).First(&CustomMetrics)
		if len(CustomMetrics) > 0 {
			sql := "join group_metrics on group_metrics.group_id=group_server.group_id and group_metrics.monitor_key=?"
			db.Joins(sql, JsonData.MonitorKey).Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, v := range GroupServer {
					go func(HostId, ScriptId string) {
						defer func() {
							if r := recover(); r != nil {
								err = errors.New(fmt.Sprint(r))
							}
							if err != nil {
								Log.Error(err)
							}
						}()
						_, err = common.RequestApiPost(cf.ApiUrlConfig+cf.JobApiConfig["script_run"].(string),
							platform_conf.PublicToken, map[string]interface{}{"host_id": HostId,
								"script_id": ScriptId, "dst_path": platform_conf.AgentRoot + "/script/", "not_run": true})
					}(v.HostId, CustomMetrics[0].ScriptId)
				}
			}
		} else {
			err = errors.New(JsonData.MonitorKey + "无效的监控指标")
		}
	}
}

// @Tags 监控平台
// @Summary 删除自定义指标
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DelCustomMetric true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/custom/metric [delete]
func DeleteCustom(c *gin.Context) {
	//删除自定义指标
	var (
		sqlErr         error
		MonitorKeys    []databases.MonitorKeys
		MonitorMetrics []databases.MonitorMetrics
		CustomMetrics  []databases.CustomMetrics
		GroupMetrics   []databases.GroupMetrics
		JsonData       = monitor_conf.DelCustomMetric{}
		Response       = common.Response{C: c}
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("group_id=? and monitor_key=?", JsonData.GroupId, JsonData.MonitorKey).First(&GroupMetrics)
		if len(GroupMetrics) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("monitor_key=?", JsonData.MonitorKey).Delete(&MonitorKeys).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("monitor_key=?", JsonData.MonitorKey).Delete(&MonitorMetrics).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("monitor_key=?", JsonData.MonitorKey).Delete(&CustomMetrics).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("monitor_key=?", JsonData.MonitorKey).Delete(&GroupMetrics).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		}
	}
}

// @Tags 监控平台
// @Summary 查询自定义指标
// @Produce  json
// @Security ApiKeyAuth
// @Param group_id query string true "资源组ID"
// @Param monitor_key query string true "监控指标"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/custom/metric [get]
func QueryCustom(c *gin.Context) {
	//报警规则查询接口
	var (
		sqlErr        error
		MonitorKeys   []databases.MonitorKeys
		CustomMetrics []databases.CustomMetrics
		GroupMetrics  []databases.GroupMetrics
		JsonData      monitor_conf.QueryCustomMetric
		Response      = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)

	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("group_id=? and monitor_key=?", JsonData.GroupId, JsonData.MonitorKey).First(&GroupMetrics)
		if len(GroupMetrics) > 0 {
			db.Where("monitor_key=?", JsonData.MonitorKey).First(&MonitorKeys)
			db.Where("monitor_key=?", JsonData.MonitorKey).First(&CustomMetrics)
			if len(MonitorKeys) > 0 && len(CustomMetrics) > 0 {
				Response.Data = map[string]interface{}{
					"monitor_key_cn":   MonitorKeys[0].MonitorKeyCn,
					"monitor_key_unit": MonitorKeys[0].MonitorKeyUnit,
					"script_id":        CustomMetrics[0].ScriptId}
			}
		}
	}
}

// @Tags 监控平台
// @Summary 查询资源组指标
// @Produce  json
// @Security ApiKeyAuth
// @Param group_id query string true "资源组ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/custom/group/metric [get]
func QueryGroupCustom(c *gin.Context) {
	//查询资源组指标
	var (
		sqlErr       error
		GroupMetrics []databases.GroupMetrics
		JsonData     monitor_conf.QueryGroupCustom
		Response     = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)

	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("group_id=?", JsonData.GroupId).Find(&GroupMetrics)
		Response.Data = GroupMetrics
	}
}
