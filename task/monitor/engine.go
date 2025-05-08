package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/netutil"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
	"github.com/spf13/cast"
	"inner/conf/monitor_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"net"
	"strconv"
	"time"
)

func AlarmEngine() {
	Log.Info("监控引擎开始工作......")
	for {
		data := <-monitor_conf.Ach
		go func(data monitor_conf.AlarmData) {
			var (
				MonitorRules  []databases.MonitorRules
				AlarmHistory  []databases.AlarmHistory
				MonitorGroups []databases.MonitorGroups
				TraceId       string
				MonitorKey    = "monitor_" + data.HostId + "_" + data.MonitorResource +
					"_" + data.MonitorItem + "_" + data.MonitorKey
				RuleIdKey           = MonitorKey + "_rule_id"
				TraceKey            = MonitorKey + "_trace_id"
				AlarmIntervalKey    = MonitorKey + "_alarm_interval"
				RecoveryIntervalKey = MonitorKey + "_recovery_interval"
				GroupServer         []databases.GroupServer
				MonitorKeys         []databases.MonitorKeys
			)
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
			}()
			hostName := rc.HGet(ctx, platform_conf.ServerNameKey, data.HostId).Val()
			// 缓存规则，减轻数据库压力
			if rc.Exists(ctx, RuleIdKey).Val() == 1 {
				db.Where("rule_id=? and status=?", rc.Get(ctx, RuleIdKey).Val(), "active").Find(&MonitorRules)
			} else {
				// 判断监控主机是否绑定报警规则
				GroupId := rc.HGet(ctx, platform_conf.GroupServersKey, data.HostId).Val()
				if GroupId == "" {
					db.Where("host_id=?", data.HostId).Find(&GroupServer)
					if len(GroupServer) > 0 {
						GroupId = GroupServer[0].GroupId
					}
				}
				if GroupId != "" {
					var (
						ruleIds             []string
						monitorGroupRuleKey = "monitor_group_rule_"
					)
					if rc.Exists(ctx, monitorGroupRuleKey+GroupId).Val() == 1 {
						ruleIds = rc.SMembers(ctx, monitorGroupRuleKey+GroupId).Val()
					} else {
						db.Select("rule_id").Where("asset_group_id=? and status=?", GroupId, "active").Find(&MonitorGroups)
						if len(MonitorGroups) > 0 {
							for _, v := range MonitorGroups {
								ruleIds = append(ruleIds, v.RuleId)
								rc.SAdd(ctx, monitorGroupRuleKey+GroupId, v.RuleId)
								rc.Expire(ctx, monitorGroupRuleKey+GroupId, 5*time.Minute)
							}
						}
					}
					if len(ruleIds) > 0 {
						db.Where("rule_id in ? and monitor_resource=? and monitor_item=? and monitor_key=?  and status=?",
							ruleIds, data.MonitorResource, data.MonitorItem, data.MonitorKey, "active").Find(&MonitorRules)
						if len(MonitorRules) > 0 {
							rc.Set(ctx, RuleIdKey, MonitorRules[0].RuleId, 5*time.Minute)
						} else {
							rc.Del(ctx, TraceKey)
						}
					} else {
						rc.Del(ctx, TraceKey)
					}
				} else {
					rc.Del(ctx, TraceKey)
				}
			}
			var i int
			var rule databases.MonitorRules
			// 获取监控项匹配的报警规则
			if len(MonitorRules) > 0 {
				for _, v := range MonitorRules {
					db.Where("monitor_key=?", v.MonitorKey).Find(&MonitorKeys)
					if len(MonitorKeys) > 0 {
						var diffBreak bool
						if MonitorKeys[0].MonitorKey == "cpu_loadavg" {
							var Value float64
							if rc.Exists(ctx, platform_conf.HostCpuCoreKey).Val() == 1 {
								Value, _ = strconv.ParseFloat(rc.HGet(ctx, platform_conf.HostCpuCoreKey, data.HostId).Val(), 64)
								if v.RuleValue <= 1 {
									v.RuleValue = Value
								} else {
									v.RuleValue = Value * v.RuleValue
								}
								v.AlarmContent = MonitorKeys[0].MonitorKeyCn + v.DiffRule + cast.ToString(v.RuleValue)
							} else {
								diffBreak = true
							}
						}
						if !diffBreak {
							// 监控数据是否触发规则
							data.MonitorValue = kits.FormatMonitorValue(data.MonitorValue, MonitorKeys[0].MonitorKeyUnit)
							if DiffRule(v.RuleValue, v.DiffRule, data.MonitorValue) {
								// 多个报警等级选取最高等级
								if monitor_conf.AlarmLevels[v.AlarmLevel] > i {
									i = monitor_conf.AlarmLevels[v.AlarmLevel]
									rule = v
								}
							}
						}
					}
				}
			}
			// 判断规则是否被触发
			if i > 0 {
				// 判断是否达到连续采集周期条件
				rc.Incr(ctx, AlarmIntervalKey)
				v := cast.ToInt(rc.Get(ctx, AlarmIntervalKey).Val())
				// 初次才设置key过期时间
				if v == 1 {
					rc.Expire(ctx, AlarmIntervalKey,
						time.Duration(int32(rule.RuleT)*data.MonitorInterval+30)*time.Second)
				}
				if v >= rule.RuleT {
					//判断agent是否在升级
					upgradeKey := platform_conf.UpgradeKey + data.HostId
					if rc.Exists(ctx, upgradeKey).Val() == 0 {
						//判断是否为初次报警
						if rc.Exists(ctx, TraceKey).Val() == 1 {
							//持续报警跟踪，支持热修改报警规则
							TraceId = rc.Get(ctx, TraceKey).Val()
							rc.Set(ctx, TraceId+"alarm_content", rule.AlarmContent, 0)
							db.Where("trace_id=?", TraceId).First(&AlarmHistory)
							if len(AlarmHistory) > 0 {
								db.Model(&AlarmHistory).Updates(databases.AlarmHistory{EndTime: time.Now(),
									Duration: cast.ToInt64(time.Now().Sub(
										AlarmHistory[0].StartTime).Seconds()),
									Content:    rule.AlarmContent,
									AlarmLevel: rule.AlarmLevel,
									RuleName:   rule.RuleName,
									RuleId:     rule.RuleId})
							} else {
								rc.Del(ctx, TraceKey)
							}
							Log.Info("服务器:" + hostName + " 规则id:" + rule.RuleId + " trace_id:" + TraceId + "持续触发报警")
						} else {
							//将同rule_id的异常报警记录状态置为unknown
							db.Where("host_id=? and rule_id=? and status=?",
								data.HostId, rule.RuleId, "fault").First(&AlarmHistory)
							if len(AlarmHistory) > 0 {
								db.Model(&AlarmHistory).Where("host_id=? and rule_id=? and "+
									"status=?", data.HostId, rule.RuleId, "fault").Updates(
									databases.AlarmHistory{Status: "unknown"})
							}
							//新报警记录写入数据库
							TraceId = kits.RandString(12)
							da := databases.AlarmHistory{StartTime: time.Now(), EndTime: time.Now(),
								MonitorResource: rule.MonitorResource, MonitorItem: rule.MonitorItem,
								AlarmLevel: rule.AlarmLevel, Content: rule.AlarmContent, Duration: 0,
								RuleId: rule.RuleId, RuleName: rule.RuleName,
								RuleType: rule.RuleType, HostId: data.HostId, Status: "fault",
								TraceId: TraceId}
							err = db.Create(&da).Error
							if err == nil {
								rc.Set(ctx, TraceKey, TraceId, 0)
								Log.Info("HostName:" + hostName + " RuleId:" + rule.RuleId + " TraceId:" + TraceId + "新报警记录写入数据库")
							}
						}
						rc.Del(ctx, AlarmIntervalKey)
						//发送报警通知写入channel
						rc.Set(ctx, TraceId+"alarm_content", rule.AlarmContent, 0)
						monitor_conf.Sch <- monitor_conf.SendMsg{Status: "fault", HostId: data.HostId,
							RuleId: rule.RuleId, MonitorResource: data.MonitorResource, MonitorItem: data.MonitorItem,
							MonitorValue: data.MonitorValue, TraceId: TraceId, MonitorInterval: data.MonitorInterval,
							AlarmContent: rule.AlarmContent}
					}
				}
			} else {
				//判断是否恢复
				if rc.Exists(ctx, TraceKey).Val() == 1 {
					TraceId = rc.Get(ctx, TraceKey).Val()
					rc.Incr(ctx, RecoveryIntervalKey)
					v := cast.ToInt(rc.Get(ctx, RecoveryIntervalKey).Val())
					//初次恢复设置过期时间
					if v == 1 {
						_, err = rc.Expire(ctx, RecoveryIntervalKey,
							time.Duration(3*data.MonitorInterval)*time.Second).Result()
					}
					//连续2次设定为恢复
					if v >= 2 {
						db.Where("trace_id=?", TraceId).First(&AlarmHistory)
						if len(AlarmHistory) > 0 {
							err = db.Model(&AlarmHistory).Updates(databases.AlarmHistory{EndTime: time.Now(),
								Duration: cast.ToInt64(time.Now().Sub(
									AlarmHistory[0].StartTime).Seconds()), Status: "recovery"}).Error
							if err == nil {
								alarmContent := rc.Get(ctx, TraceId+"alarm_content").Val()
								//发送恢复通知写入channel
								monitor_conf.Sch <- monitor_conf.SendMsg{Status: "recovery", HostId: data.HostId,
									RuleId: rule.RuleId, MonitorResource: data.MonitorResource,
									MonitorItem: data.MonitorItem, MonitorValue: data.MonitorValue,
									TraceId: TraceId, MonitorInterval: data.MonitorInterval,
									AlarmContent: alarmContent}
								rc.Del(ctx, RecoveryIntervalKey)
								rc.Del(ctx, TraceKey)
								rc.Del(ctx, TraceId+"alarm_content")
								rc.Del(ctx, RuleIdKey)
								Log.Info("服务器:" + hostName + " RuleId:" + rule.RuleId + " trace_id:" + TraceId + "报警已恢复!")
							}
						}
					}
				}
			}
		}(data)
	}
}

func SendEngine() {
	Log.Info("通知引擎开始工作......")
	for {
		data := <-monitor_conf.Sch
		go func(data monitor_conf.SendMsg) {
			var (
				GroupServer   []databases.GroupServer
				MonitorGroups []databases.MonitorGroups
				MonitorRules  []databases.MonitorRules
				AlarmHistory  []databases.AlarmHistory
				AlarmSend     []databases.AlarmSend
				AlarmChannel  []databases.AlarmChannel
				MonitorKeys   []databases.MonitorKeys
				MonitorJobs   []databases.MonitorJobs
				Msg           = map[string]interface{}{}
				SendMsg       bool
				cf            = platform_conf.Setting()
				Titles        = map[string]string{"fault": "监控报警通知", "recovery": "监控恢复通知"}
				Status        = map[string]string{"fault": "故障中", "recovery": "已恢复"}
				ItemCn        = map[string]string{"system": "系统", "custom": "自定义"}
				AlarmLevelCn  = map[string]string{"Info": "通知", "Warning": "警告", "Critical": "严重", "Error": "致命"}
			)
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
			}()
			//判断报警通知是否暂停
			if rc.Exists(ctx, monitor_conf.PauseAlarmKey+"_"+data.TraceId).Val() == 0 {
				db.Where("trace_id=?", data.TraceId).Find(&AlarmHistory)
				if len(AlarmHistory) > 0 {
					//验证数据有效性
					switch AlarmHistory[0].RuleType {
					case "BuiltInRule":
						MonitorRules = BuiltInRule(data.MonitorResource, data.MonitorItem)
					default:
						db.Where("rule_id=? and status=?", AlarmHistory[0].RuleId,
							"active").Find(&MonitorRules)
					}
					if len(MonitorRules) > 0 {
						key := "monitor_alarm_" + data.HostId + "aggregation"
						switch data.Status {
						case "fault":
							//不收敛聚合进行报警通知
							db.Where("host_id=? and rule_id=? and status=?", data.HostId, "BuiltInRule", "fault").Find(&AlarmHistory)
							if len(AlarmHistory) == 0 || data.RuleId == "BuiltInRule" {
								//首次报警通知不受通知发送周期限制
								Data := map[string]interface{}{}
								var uri string
								db.Where("rule_id=?", MonitorRules[0].RuleId).Find(&MonitorJobs)
								if len(MonitorJobs) > 0 {
									if MonitorJobs[0].Exec != "None" {
										Data = map[string]interface{}{"host_ids": []string{data.HostId},
											"exec": MonitorJobs[0].Exec, "cron": false}
										uri = cf.JobApiConfig["exec_run"].(string)
									}
									if MonitorJobs[0].ScriptId != "None" {
										Data = map[string]interface{}{"host_id": data.HostId,
											"script_id": MonitorJobs[0].ScriptId, "cron": false}
										uri = cf.JobApiConfig["script_run"].(string)
									}
									v, err := common.RequestApiPost(cf.ApiUrlConfig+uri,
										platform_conf.PublicToken, Data)
									hostName := rc.HGet(ctx, platform_conf.ServerNameKey, data.HostId).Val()
									if v != nil && err == nil && v.(map[string]interface{})["success"] != nil &&
										v.(map[string]interface{})["success"].(bool) {
										Log.Info("主机(" + hostName + ")故障自愈发送成功")
									} else {
										Log.Info("主机(" + hostName + ")故障自愈发送失败")
									}
								} else {
									SendMsg = SendStages(MonitorRules[0].RuleId, data.TraceId, data.MonitorInterval)
								}
							}
						case "recovery":
							rc.Del(ctx, key)
							//报警恢复通知不受发送周期限制
							db.Where("trace_id=?", data.TraceId).Find(&AlarmSend)
							if len(AlarmSend) > 0 {
								SendMsg = true
							}
						}
						//匹配发送周期后发送报警通知
						if SendMsg {
							now := time.Now()
							switch AlarmHistory[0].RuleType {
							case "BuiltInRule":
								db.Where("host_id=?", data.HostId).Find(&GroupServer)
								if len(GroupServer) > 0 {
									db.Where("asset_group_id=?", GroupServer[0].GroupId).Find(&MonitorGroups)
									if len(MonitorGroups) > 0 {
										db.Where("rule_id = ? and status=?", MonitorGroups[0].RuleId, "active").Find(&AlarmChannel)
									}
								}
							default:
								db.Where("rule_id=? and status=?", MonitorRules[0].RuleId,
									"active").Find(&AlarmChannel)
							}
							if len(AlarmChannel) > 0 {
								var (
									ip          string
									idc         = "未知"
									hostType    string
									AssetServer []databases.AssetServer
									AssetNet    []databases.AssetNet
									AssetIdc    []databases.AssetIdc
								)
								sql := "join asset_extend on asset_extend.idc_id=asset_idc.idc_id " +
									"and asset_extend.host_id=?"
								db.Joins(sql, data.HostId).Find(&AssetIdc)
								db.Where("host_id=?", data.HostId).Find(&AssetNet)
								db.Where("host_id=?", data.HostId).Find(&AssetServer)
								if len(AssetServer) > 0 && len(AssetNet) > 0 {
									for _, v := range AssetNet {
										if netutil.IsInternalIP(net.ParseIP(v.Ip)) {
											ip = v.Ip
											break
										}
									}
									if len(AssetIdc) > 0 {
										idc = AssetIdc[0].IdcCn
									}
									hostType = AssetServer[0].HostTypeCn
								}
								for _, v := range AlarmChannel {
									nt := time.Now().Format("15:04:05")
									if nt > v.StartTime && v.EndTime > nt {
										var Result string
										switch v.Channel {
										case "DingDing":
											data.MonitorValue, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", data.MonitorValue),
												64)
											db.Where("monitor_key=?", MonitorRules[0].MonitorKey).Find(&MonitorKeys)
											if len(MonitorKeys) > 0 {
												item := MonitorRules[0].MonitorItem
												_, ok := ItemCn[MonitorRules[0].MonitorItem]
												if ok {
													item = ItemCn[MonitorRules[0].MonitorItem]
												}
												level := MonitorRules[0].AlarmLevel
												_, ok = AlarmLevelCn[MonitorRules[0].AlarmLevel]
												if ok {
													level = AlarmLevelCn[MonitorRules[0].AlarmLevel]
												}
												hostName := rc.HGet(ctx, platform_conf.ServerNameKey, data.HostId).Val()
												Msg = map[string]interface{}{
													"msgtype": "markdown",
													"markdown": map[string]string{
														"title": Titles[data.Status],
														"text": "### 服务器" + hostName + "报警信息:" + "\n " +
															"> - 当前状态: " + Status[data.Status] + "\n" +
															"> - 主机类型: " + hostType + "\n" +
															"> - 主机IP: " + ip + "\n" +
															"> - 所属IDC: " + idc + "\n" +
															"> - 规则名称: " + MonitorRules[0].RuleName + "\n" +
															"> - 监控项: " + item + "\n" +
															"> - 监控指标: " + MonitorRules[0].MonitorKey + "\n" +
															"> - 触发条件: " + data.AlarmContent + "\n" +
															"> - 当前值: " + cast.ToString(data.MonitorValue) + MonitorKeys[0].MonitorKeyUnit + "\n" +
															"> - 持续时间: " + kits.TimeFormat(float64(AlarmHistory[0].Duration)) + "\n" +
															"> - 报警等级: " + level + "\n" +
															"> - 报警时间: " + now.Format("2006-01-02 15:04:05"),
													},
												}
												if common.SendDingDingMsg(v.Address, Msg) {
													Result = "success"
												}
											}
										default:
											Result = "fault"
										}
										if Result != "" && len(Msg) > 0 {
											hostName := rc.HGet(ctx, platform_conf.ServerNameKey, AlarmHistory[0].HostId).Val()
											Log.Info("服务器:" + hostName + " TraceId:" + AlarmHistory[0].TraceId + "发送报警通知")
											html := blackfriday.MarkdownBasic([]byte(Msg["markdown"].(map[string]string)["text"]))
											as := databases.AlarmSend{SendTime: now, HostId: data.HostId,
												RuleId: MonitorRules[0].RuleId, Channel: v.Channel,
												Content: string(html), Result: Result,
												TraceId: data.TraceId}
											err = db.Create(&as).Error
										}
									} else {
										hostName := rc.HGet(ctx, platform_conf.ServerNameKey, data.HostId).Val()
										Log.Info("服务器:" + hostName + " TraceId:" + data.TraceId + "不在报警通知时间范围")
									}
								}
							}
						}
					}
				} else {
					if data.RuleId == "BuiltInRule" {
						da := databases.AlarmHistory{StartTime: time.Now(), EndTime: time.Now(),
							MonitorResource: "server", MonitorItem: "system",
							AlarmLevel: "Error", Content: "agent运行异常,主机不可达", Duration: 0,
							RuleId: "BuiltInRule", RuleName: "内置规则",
							RuleType: "BuiltInRule", HostId: data.HostId, Status: "fault",
							TraceId: data.TraceId}
						db.Create(&da)
					}
				}
			}
		}(data)
	}
}

func SendStages(RuleId, TraceId string, MonitorInterval int32) bool {
	var (
		AlarmStages   []databases.AlarmStages
		Stages        = []map[string]int{{"stage": 1, "interval": 1}}
		StageKey      = "monitor_alarm_stage_" + TraceId
		StageIndexKey = "monitor_alarm_stage_index_" + TraceId
		IntervalKey   = "monitor_alarm_interval_" + TraceId
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	db.Where("rule_id=?", RuleId).Find(&AlarmStages)
	if len(AlarmStages) > 0 {
		err = json.Unmarshal([]byte(AlarmStages[0].Stages), &Stages)
	}
	if err == nil {
		//验证发送步骤
		stage := map[string]int{}
		if rc.Exists(ctx, StageIndexKey).Val() == 0 {
			stage = Stages[0]
			if stage["stage"] > 0 {
				rc.Set(ctx, StageIndexKey, 0, time.Duration(MonitorInterval)*time.Minute)
			}
		} else {
			s, _ := strconv.Atoi(rc.Get(ctx, StageIndexKey).Val())
			if s < len(Stages) {
				stage = Stages[s]
				rc.Expire(ctx, StageIndexKey, time.Duration(MonitorInterval)*time.Minute)
				if s+1 < len(Stages) {
					stage = Stages[s+1]
				}
				if stage["stage"] > 0 {
					if rc.Exists(ctx, StageKey).Val() == 1 {
						t, _ := strconv.Atoi(rc.Get(ctx, StageKey).Val())
						if t >= stage["stage"] {
							rc.Set(ctx, StageIndexKey, s+1, time.Duration(MonitorInterval)*time.Minute)
						}
					}
				}
			}
		}
		//验证发送周期
		if len(stage) > 0 {
			rc.Incr(ctx, IntervalKey)
			rc.Expire(ctx, IntervalKey, time.Duration(MonitorInterval)*time.Minute)
			i, _ := strconv.Atoi(rc.Get(ctx, IntervalKey).Val())
			if i >= stage["interval"] || stage["interval"] == 0 {
				rc.Del(ctx, IntervalKey)
				if stage["stage"] > 0 {
					rc.Incr(ctx, StageKey)
				}
				Log.Info("trac_id:" + TraceId + "发送报警通知条件通过")
				return true
			}
		}
	}
	Log.Info("trac_id:" + TraceId + "发送报警通知条件未通过")
	return false
}

func DiffRule(alarmValue float64, diff string, monitorValue float64) bool {
	var (
		result bool
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	switch diff {
	case ">":
		if monitorValue > alarmValue {
			result = true
		}
	case "=":
		if monitorValue == alarmValue {
			result = true
		}
	case "<":
		if monitorValue < alarmValue {
			result = true
		}
	case "!=":
		if monitorValue != alarmValue {
			result = true
		}
	case ">=":
		if monitorValue >= alarmValue {
			result = true
		}
	case "<=":
		if monitorValue <= alarmValue {
			result = true
		}
	}
	return result
}

func BuiltInRule(MonitorResource, MonitorItem string) []databases.MonitorRules {
	var (
		MonitorRules []databases.MonitorRules
		AlarmContent string
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	switch MonitorResource {
	case "server":
		AlarmContent = monitor_conf.DefaultRuleContents[MonitorResource]
	case "process":
		AlarmContent = MonitorItem + monitor_conf.DefaultRuleContents[MonitorResource]
	}
	if AlarmContent != "" {
		MonitorRules = append(MonitorRules, databases.MonitorRules{Id: 0,
			RuleId:          "BuiltInRule",
			RuleName:        "内置规则",
			RuleType:        "BuiltInRule",
			MonitorResource: MonitorResource,
			MonitorItem:     MonitorItem,
			AlarmLevel:      "Error",
			MonitorKey:      "alive",
			RuleValue:       0,
			DiffRule:        "=",
			RuleT:           5,
			AlarmContent:    AlarmContent,
			Status:          "active",
			CreateUser:      "platform",
			CreateTime:      time.Now(),
			UpdateUser:      "None",
			UpdateTime:      time.Now(),
			RuleMd5:         kits.RandString(12),
		})
	}
	return MonitorRules
}
