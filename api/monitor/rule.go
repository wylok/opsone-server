package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/monitor_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 监控平台
// @Summary 创建报警规则
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body monitor_conf.CreateRule true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/rule [post]
func CreateRule(c *gin.Context) {
	//创建报警规则
	var (
		sqlErr         error
		JsonData       = monitor_conf.CreateRule{}
		MonitorRules   []databases.MonitorRules
		MonitorMetrics []databases.MonitorMetrics
		MonitorKeys    []databases.MonitorKeys
		MonitorGroups  []databases.MonitorGroups
		Response       = common.Response{C: c}
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
		//验证规则参数
		err = db.Where("monitor_key=?", JsonData.MonitorKey).First(&MonitorMetrics).Error
		if len(MonitorMetrics) > 0 {
			RuleMd5 := kits.MD5(JsonData.MonitorResource + JsonData.MonitorItem + JsonData.MonitorKey +
				JsonData.AlarmLevel + JsonData.DiffRule)
			//验证规则是否重复
			db.Where("rule_md5=?", RuleMd5).First(&MonitorRules)
			db.Where("monitor_key=?", JsonData.MonitorKey).First(&MonitorKeys)
			if len(MonitorRules) == 0 && len(MonitorKeys) > 0 {
				RuleId := kits.RandString(8)
				err = db.Transaction(func(tx *gorm.DB) error {
					//新增报警通知规则
					var Stages []byte
					if JsonData.Stages == nil {
						Stages, _ = json.Marshal([]map[string]int{{"interval": 3, "stage": 0}})
					} else {
						if FormatStages(JsonData.Stages) {
							Stages, _ = json.Marshal(JsonData.Stages)
						} else {
							err = errors.New("stages数据格式验证失败")
						}
					}
					if Stages != nil {
						as := databases.AlarmStages{RuleId: RuleId, Stages: string(Stages)}
						if err = tx.Create(&as).Error; err != nil {
							sqlErr = err
						}
					}
					//新增报警规则
					ac := MonitorKeys[0].MonitorKeyCn + JsonData.DiffRule + cast.ToString(JsonData.RuleValue)
					if MonitorKeys[0].MonitorKeyUnit != "" {
						ac = ac + MonitorKeys[0].MonitorKeyUnit
					}
					mr := databases.MonitorRules{
						RuleId:          RuleId,
						RuleName:        JsonData.RuleName,
						RuleType:        "custom_rule",
						MonitorResource: JsonData.MonitorResource,
						MonitorItem:     JsonData.MonitorItem,
						AlarmLevel:      JsonData.AlarmLevel,
						MonitorKey:      JsonData.MonitorKey,
						RuleValue:       JsonData.RuleValue,
						DiffRule:        JsonData.DiffRule,
						RuleT:           JsonData.RuleT,
						AlarmContent:    ac,
						Status:          "active",
						CreateUser:      c.GetString("user_id"),
						CreateTime:      time.Now(),
						UpdateUser:      c.GetString("user_id"),
						UpdateTime:      time.Now(),
						RuleMd5:         RuleMd5,
					}
					if err = tx.Create(&mr).Error; err != nil {
						sqlErr = err
					}
					if JsonData.GroupIds != nil {
						//新增资源组与报警规则关联
						for _, groupId := range JsonData.GroupIds {
							db.Where("asset_group_id=? and rule_id=?", groupId, RuleId).First(&MonitorGroups)
							if len(MonitorGroups) == 0 {
								mg := databases.MonitorGroups{AssetGroupId: groupId, RuleId: RuleId,
									AutoRelation: cast.ToInt(JsonData.Relation), CreateAt: time.Now(), Status: "active"}
								if err = tx.Create(&mg).Error; err != nil {
									sqlErr = err
								}
							}
						}
					}
					//新增报警渠道及地址
					if JsonData.Channels != nil {
						for _, val := range JsonData.Channels {
							startTime := "09:00:00"
							endTime := "23:00:00"
							s, ok := val["start_time"]
							if ok {
								startTime = s.(string)
							}
							e, ok := val["end_time"]
							if ok {
								endTime = e.(string)
							}
							channel, ok := val["channel"]
							if ok {
								contacts, ok := val["contacts"]
								if ok {
									acl := databases.AlarmChannel{RuleId: RuleId,
										Channel:   channel.(string),
										Address:   contacts.(string),
										StartTime: startTime,
										EndTime:   endTime,
										Status:    "active"}
									if err = tx.Create(&acl).Error; err != nil {
										sqlErr = err
									}
								}
							}
						}
					}
					return sqlErr
				})
			} else {
				err = errors.New("相同报警规则已存在,请勿重复创建")
			}
		}
	}
}

// @Tags 监控平台
// @Summary 修改报警规则
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.ModifyRule true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/rule [put]
func ModifyRule(c *gin.Context) {
	//修改报警规则
	var (
		sqlErr        error
		JsonData      = monitor_conf.ModifyRule{}
		MonitorRules  []databases.MonitorRules
		MonitorGroups []databases.MonitorGroups
		MonitorKeys   []databases.MonitorKeys
		AlarmHistory  []databases.AlarmHistory
		AlarmChannel  []databases.AlarmChannel
		Response      = common.Response{C: c}
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
		db.Where("rule_id=?", JsonData.RuleId).First(&MonitorRules)
		if len(MonitorRules) > 0 {
			data := MonitorRules[0]
			if JsonData.RuleName != "" {
				data.RuleName = JsonData.RuleName
			}
			if JsonData.AlarmLevel != "" {
				data.AlarmLevel = JsonData.AlarmLevel
			}
			if JsonData.DiffRule != "" {
				data.DiffRule = JsonData.DiffRule
			}
			if JsonData.Status != "" {
				data.Status = JsonData.Status
			}
			if JsonData.RuleValue != 0 {
				data.RuleValue = JsonData.RuleValue
			}
			if JsonData.RuleT != 0 {
				data.RuleT = JsonData.RuleT
			}
			//修改报警内容
			err = db.Where("monitor_key=?", data.MonitorKey).First(&MonitorKeys).Error
			if err == nil {
				ac := MonitorKeys[0].MonitorKeyCn + data.DiffRule + cast.ToString(data.RuleValue)
				if MonitorKeys[0].MonitorKeyUnit != "" {
					ac = ac + MonitorKeys[0].MonitorKeyUnit
				}
				data.AlarmContent = ac
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				//修改资源组与规则绑定
				if JsonData.GroupIds != nil {
					if err = tx.Where("rule_id=?", JsonData.RuleId).Delete(&MonitorGroups).Error; err != nil {
						sqlErr = err
					}
					for _, GroupId := range JsonData.GroupIds {
						mg := databases.MonitorGroups{AssetGroupId: GroupId, RuleId: JsonData.RuleId,
							AutoRelation: 1, CreateAt: time.Now(), Status: "active"}
						if err = tx.Create(&mg).Error; err != nil {
							sqlErr = err
						}
					}
					if err = tx.Model(&AlarmHistory).Where("rule_id in ? && status=?",
						[]string{JsonData.RuleId, "BuiltInRule"}, "fault").Updates(
						databases.AlarmHistory{Status: "unknown"}).Error; err != nil {
						sqlErr = err
					}
				}
				//修改报警通知渠道
				if JsonData.Channels != nil {
					if err = tx.Where("rule_id=?", JsonData.RuleId).Delete(&AlarmChannel).Error; err != nil {
						sqlErr = err
					}
					for _, v := range JsonData.Channels {
						startTime := "09:00:00"
						endTime := "23:00:00"
						s, ok := v["start_time"]
						if ok {
							startTime = s.(string)
						}
						e, ok := v["end_time"]
						if ok {
							endTime = e.(string)
						}
						channel, ok := v["channel"]
						if ok {
							contacts, ok := v["contacts"]
							if ok {
								acl := databases.AlarmChannel{RuleId: JsonData.RuleId,
									Channel: channel.(string), Address: contacts.(string),
									StartTime: startTime,
									EndTime:   endTime,
									Status:    "active"}
								if err = tx.Create(&acl).Error; err != nil {
									sqlErr = err
								}
							}
						}
					}
				}
				RuleMd5 := kits.MD5(data.MonitorResource + data.MonitorItem + data.MonitorKey + data.AlarmLevel + data.DiffRule)
				if data.RuleMd5 != RuleMd5 {
					data.RuleMd5 = RuleMd5
				}
				if JsonData.Status == "close" {
					if err = tx.Model(&AlarmHistory).Where("rule_id=? && status=?", JsonData.RuleId, "fault").Updates(
						databases.AlarmHistory{Status: "unknown"}).Error; err != nil {
						sqlErr = err
					}
				}
				data.UpdateTime = time.Now()
				data.UpdateUser = c.GetString("user_id")
				if err = tx.Model(&MonitorRules).Updates(data).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		}
	}
}

// @Tags 监控平台
// @Summary 删除报警规则
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteRule true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/monitor/rule [delete]
func DeleteRule(c *gin.Context) {
	//删除报警规则
	var (
		sqlErr        error
		MonitorRules  []databases.MonitorRules
		AlarmStages   []databases.AlarmStages
		MonitorGroups []databases.MonitorGroups
		AlarmChannel  []databases.AlarmChannel
		MonitorJobs   []databases.MonitorJobs
		AlarmHistory  []databases.AlarmHistory
		AlarmSend     []databases.AlarmSend
		JsonData      = monitor_conf.DeleteRule{}
		Response      = common.Response{C: c}
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
			//删除报警历史记录
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmHistory).Error; err != nil {
				sqlErr = err
			}
			//删除报警通知记录
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmSend).Error; err != nil {
				sqlErr = err
			}
			//删除资源组与规则绑定
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&MonitorGroups).Error; err != nil {
				sqlErr = err
			}
			//删除报警发送规则
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmStages).Error; err != nil {
				sqlErr = err
			}
			//删除报警发送渠道
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&AlarmChannel).Error; err != nil {
				sqlErr = err
			}
			//删除监控规则
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&MonitorRules).Error; err != nil {
				sqlErr = err
			}
			//删除故障自愈
			if err = tx.Where("rule_id in ?", JsonData.RuleIds).Delete(&MonitorJobs).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
}

// @Tags 监控平台
// @Summary 查询报警规则
// @Produce  json
// @Security ApiKeyAuth
// @Param user_id query string false "用户ID"
// @Param rule_ids query array false "规则ID"
// @Param rule_name query string false "规则名称"
// @Param monitor_resource query string false "监控资源"
// @Param monitor_item query string false "监控项"
// @Param monitor_key query string false "监控指标"
// @Param alarm_level query string false "报警等级" Enums["Info","Warning","Critical","Error"]
// @Param status query string false "规则状态" Enums["active","close"]
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/rule [get]
func QueryRule(c *gin.Context) {
	//报警规则查询接口
	var (
		MonitorRules []databases.MonitorRules
		JsonData     monitor_conf.QueryRule
		Response     = common.Response{C: c}
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
		tx := db.Order("monitor_rules.create_time desc")
		// 部分参数匹配
		if JsonData.UserId != "" {
			tx = tx.Where("monitor_rules.create_user = ? or monitor_rules.update_user=?",
				JsonData.UserId, JsonData.UserId)
		}
		if JsonData.RuleIds != nil {
			tx = tx.Where("monitor_rules.rule_id in ?", JsonData.RuleIds)
		}
		if JsonData.RuleName != "" {
			tx = tx.Where("monitor_rules.rule_name like ?", "%"+JsonData.RuleName+"%")
		}
		if JsonData.MonitorResource != "" {
			tx = tx.Where("monitor_rules.monitor_resource = ?", JsonData.MonitorResource)
		}
		if JsonData.MonitorItem != "" {
			tx = tx.Where("monitor_rules.monitor_item = ?", JsonData.MonitorItem)
		}
		if JsonData.MonitorKey != "" {
			tx = tx.Where("monitor_rules.monitor_key = ?", JsonData.MonitorKey)
		}
		if JsonData.AlarmLevel != "" {
			tx = tx.Where("monitor_rules.alarm_level = ?", JsonData.AlarmLevel)
		}
		if JsonData.Status != "" {
			tx = tx.Where("monitor_rules.status = ?", JsonData.Status)
		}
		d := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = d.Paging(&MonitorRules)
	}
}

// @Tags 监控平台
// @Summary 查询报警通知规则
// @Produce  json
// @Security ApiKeyAuth
// @Param rule_id query string true "规则ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/stages [get]
func QueryStages(c *gin.Context) {
	//报警通知规则查询接口
	var (
		AlarmStages []databases.AlarmStages
		JsonData    monitor_conf.QueryStages
		Stages      []map[string]interface{}
		Response    = common.Response{C: c}
		data        []interface{}
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
		db.Where("rule_id=?", JsonData.RuleId).First(&AlarmStages)
		if len(AlarmStages) > 0 {
			err = json.Unmarshal([]byte(AlarmStages[0].Stages), &Stages)
		} else {
			Stages = []map[string]interface{}{{"interval": 3, "stage": 0}}
		}
		//解析报警发送规则
		for _, v := range Stages {
			content := "报警连续触发" + cast.ToString(v["interval"].(float64)) + "次"
			if v["stage"].(float64) == 0 {
				content = content + "发送报警通知并周期性一直发送"
			} else {
				content = content + "发送报警并周期性发送" + cast.ToString(v["stage"].(float64)) + "次"
			}
			data = append(data, map[string]interface{}{"data": v, "content": content})
		}
		Response.Data = data
	}
}

// @Tags 监控平台
// @Summary 查询报警规则关联资源组
// @Produce  json
// @Security ApiKeyAuth
// @Param rule_ids query array true "规则ID列表"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/rule/groups [get]
func QueryRuleGroups(c *gin.Context) {
	//查询报警规则关联资源组接口
	var (
		MonitorGroups []databases.MonitorGroups
		JsonData      monitor_conf.RuleGroups
		Response      = common.Response{C: c}
		data          []string
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
		db.Distinct("asset_group_id").Select("asset_group_id").Where("rule_id in ?",
			JsonData.RuleIds).Find(&MonitorGroups)
		if len(MonitorGroups) > 0 {
			for _, v := range MonitorGroups {
				data = append(data, v.AssetGroupId)
			}
		}
		Response.Data = data
	}
}

// @Tags 监控平台
// @Summary 报警规则关联资源组
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.RelationRuleGroups true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:{}}"
// @Router /api/v1/monitor/rule/groups [post]
func RelationRuleGroups(c *gin.Context) {
	//报警规则关联资源组
	var (
		sqlErr        error
		MonitorGroups []databases.MonitorGroups
		AlarmHistory  []databases.AlarmHistory
		JsonData      monitor_conf.RelationRuleGroups
		Response      = common.Response{C: c}
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
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, ruleId := range JsonData.RuleIds {
				gIds := map[string]struct{}{}
				jIds := map[string]struct{}{}
				db.Select("asset_group_id").Where("rule_id=?", ruleId).Find(&MonitorGroups)
				if len(MonitorGroups) > 0 {
					for _, v := range MonitorGroups {
						gIds[v.AssetGroupId] = struct{}{}
					}
				}
				if len(JsonData.GroupIds) > 0 {
					for _, groupId := range JsonData.GroupIds {
						jIds[groupId] = struct{}{}
						_, ok := gIds[groupId]
						if !ok {
							//新增关联的资源组id
							dmg := databases.MonitorGroups{AssetGroupId: groupId, RuleId: ruleId, AutoRelation: 1,
								CreateAt: time.Now(), Status: "active"}
							if err = tx.Create(&dmg).Error; err != nil {
								sqlErr = err
							}
						}
					}
					if len(gIds) > 0 {
						for g := range gIds {
							_, ok := jIds[g]
							if !ok {
								//取消关联的资源组id
								if err = tx.Where("rule_id=? and asset_group_id = ?", ruleId,
									g).Delete(&MonitorGroups).Error; err != nil {
									sqlErr = err
								}
							}
						}
					}
				} else {
					//取消关联的资源组id
					if err = tx.Where("rule_id=?", ruleId).Delete(&MonitorGroups).Error; err != nil {
						sqlErr = err
					}
				}
				if err = tx.Model(&AlarmHistory).Where("rule_id=? && status=?", ruleId, "fault").Updates(
					databases.AlarmHistory{Status: "unknown"}).Error; err != nil {
					sqlErr = err
				}
			}
			return sqlErr
		})
	}
}

// @Tags 监控平台
// @Summary 创建/修改报警通知规则
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.CreateStages true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/stages [post]
func CreateStages(c *gin.Context) {
	//创建/修改报警通知规则接口
	var (
		sqlErr      error
		AlarmStages []databases.AlarmStages
		JsonData    monitor_conf.CreateStages
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
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
		if FormatStages(JsonData.Stages) {
			Stages, _ := json.Marshal(JsonData.Stages)
			db.Where("rule_id=?", JsonData.RuleId).First(&AlarmStages)
			if len(AlarmStages) == 0 {
				//新增报警通知规则
				as := databases.AlarmStages{RuleId: JsonData.RuleId, Stages: string(Stages)}
				err = db.Create(&as).Error
			} else {
				//修改报警通知规则
				db.Model(&AlarmStages).Updates(databases.AlarmStages{Stages: string(Stages)})
			}
		} else {
			err = errors.New("stages数据格式验证失败")
		}
	}
}

// @Tags 监控平台
// @Summary 查询报警规则联系人
// @Produce  json
// @Security ApiKeyAuth
// @Param rule_id query string true "规则ID"
// @Param channel query string true "报警渠道"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/rule/contacts [get]
func QueryRuleContacts(c *gin.Context) {
	//查询报警规则联系人
	var (
		AlarmChannel []databases.AlarmChannel
		JsonData     monitor_conf.RuleContacts
		Response     = common.Response{C: c}
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
		db.Where("rule_id=? and channel=?", JsonData.RuleId, JsonData.Channel).Find(&AlarmChannel)
		if len(AlarmChannel) > 0 {
			Response.Data = AlarmChannel[0]
		}
	}
}
