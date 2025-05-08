package monitor

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/monitor_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 监控平台
// @Summary 创建报警故障自愈
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.CreateRuleJobs true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/job [post]
func CreateMonitorJob(c *gin.Context) {
	//创建报警规则
	var (
		JsonData    = monitor_conf.CreateRuleJobs{}
		MonitorJobs []databases.MonitorJobs
		Response    = common.Response{C: c}
		userId      = c.GetString("user_id")
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
		//验证参数
		if JsonData.Exec == "" && JsonData.ScriptId == "" {
			err = errors.New("执行命令或者监控脚本必须二选一")
		} else {
			if JsonData.Exec == "" {
				JsonData.Exec = "None"
			}
			if JsonData.ScriptId == "" {
				JsonData.ScriptId = "None"
			}
			db.Where("rule_id=?", JsonData.RuleId).First(&MonitorJobs)
			if len(MonitorJobs) > 0 {
				db.Model(&MonitorJobs).Where("rule_id=?", JsonData.RuleId).Updates(
					databases.MonitorJobs{Exec: JsonData.Exec, ScriptId: JsonData.ScriptId,
						UserId: userId, UpdateTime: time.Now()})
			} else {
				rj := databases.MonitorJobs{RuleId: JsonData.RuleId, Exec: JsonData.Exec, ScriptId: JsonData.ScriptId,
					UserId: userId, CreateTime: time.Now(), UpdateTime: time.Now()}
				db.Create(&rj)
			}
		}
	}
}

// @Tags 监控平台
// @Summary 删除报警故障自愈
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteRuleJobs true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/job [delete]
func DeleteMonitorJobs(c *gin.Context) {
	//删除报警规则
	var (
		sqlErr      error
		MonitorJobs []databases.MonitorJobs
		JsonData    = monitor_conf.DeleteRuleJobs{}
		Response    = common.Response{C: c}
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
		err = db.Transaction(func(tx *gorm.DB) error {
			//删除故障自愈
			if err = tx.Where("rule_id = ?", JsonData.RuleId).Delete(&MonitorJobs).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
}

// @Tags 监控平台
// @Summary 查询报警故障自愈
// @Produce  json
// @Security ApiKeyAuth
// @Param rule_ids query array true "规则ID列表"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/job [get]
func QueryMonitorJobs(c *gin.Context) {
	//查询报警故障自愈
	var (
		MonitorJobs []databases.MonitorJobs
		JsonData    monitor_conf.QueryRuleJobs
		Response    = common.Response{C: c}
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
		if len(JsonData.RuleIds) == 0 {
			db.Find(&MonitorJobs)
		} else {
			db.Where("rule_id in ?", JsonData.RuleIds).Find(&MonitorJobs)
		}
		Response.Data = MonitorJobs
	}
}
