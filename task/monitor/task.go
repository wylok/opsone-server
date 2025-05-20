package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/golang-module/carbon"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/monitor_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strconv"
	"strings"
	"time"
)

type MonitorTrend struct {
	Duration string
}

func (mt *MonitorTrend) SyncMonitorTrend() {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	Duration := mt.Duration
	lock := common.SyncMutex{LockKey: "monitor_trend_lock_" + Duration}
	//加锁
	if lock.Lock() {
		defer lock.UnLock(true)
		Log.Info("cron: MonitorTrend_" + Duration + " start work")
		// 系统监控数据汇聚
		go SystemTrend(Duration)
		// 进程监控数据汇聚
		go ProcessTrend(Duration)
		// 自定义监控数据汇聚
		go CustomTrend(Duration)
		for {
			h := <-monitor_conf.Hch
			p := <-monitor_conf.Pch
			c := <-monitor_conf.Cch
			// 相同任务类型执行结束后退出
			if h == p && p == c {
				break
			}
		}
	}
}

func MonitorAlive() {
	lock := common.SyncMutex{LockKey: "monitor_alive_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				defer lock.UnLock(true)
				Log.Info("monitor_alive task start working ......")
				Alive()
			}()
		}
		time.Sleep(1 * time.Minute)
	}
}

func SystemTrend(Duration string) {
	var (
		tg     = monitor_conf.SystemTags{}
		influx = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		monitor_conf.Hch <- Duration
	}()
	// 监控数据趋势计算
	for {
		v := rc.SPop(ctx, monitor_conf.DataTrendKey+"_system_"+Duration).Val()
		if v != "" {
			err = json.Unmarshal([]byte(v), &tg)
			if err == nil {
				cmd := "select " + monitor_conf.DurationMeasurement[Duration] +
					" from system_" + monitor_conf.Measurements[Duration] +
					" where time > now() - " + Duration + " and host_id='" + tg.HostId + "'"
				res, err := influx.Query(cmd, true)
				if err == nil && len(res) > 0 {
					fields := map[string]interface{}{}
					for _, r := range res {
						for _, s := range r.Series {
							for i, d := range s.Columns {
								if strings.Contains(d, "max_") || strings.Contains(d, "mean_") {
									for _, p := range []string{"max_", "mean_"} {
										d = strings.Replace(d, p, "", 1)
									}
									if len(s.Values) > 0 && s.Values[0][i] != nil {
										val, _ := s.Values[0][i].(json.Number).Float64()
										fields[d] = val
									}
								}
							}
						}
					}
					tags := map[string]string{"host_id": tg.HostId, "source": tg.Source}
					err = influx.WritesPoints("system_"+Duration, tags, fields)
				}
				if err != nil {
					Log.Error(err)
				}
			}
		} else {
			break
		}
	}
}

func ProcessTrend(Duration string) {
	var (
		influx = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
		tg     = monitor_conf.ProcessTags{}
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		monitor_conf.Pch <- Duration
	}()
	// 监控数据趋势计算
	for {
		v := rc.SPop(ctx, monitor_conf.DataTrendKey+"_process_"+Duration).Val()
		if v != "" {
			err = json.Unmarshal([]byte(v), &tg)
			if err == nil {
				cmd := "select " + monitor_conf.DurationMeasurement[Duration] +
					" from process_" + monitor_conf.Measurements[Duration] + " where time > now() - " +
					Duration + " and host_id='" + tg.HostId + "' and process='" + tg.Process + "'"
				res, err := influx.Query(cmd, true)
				if err == nil && len(res) > 0 {
					fields := map[string]interface{}{}
					for _, r := range res {
						for _, s := range r.Series {
							for i, c := range s.Columns {
								if strings.Contains(c, "max_") || strings.Contains(c, "mean_") {
									for _, v := range []string{"max_", "mean_"} {
										c = strings.Replace(c, v, "", 1)
									}
									if len(s.Values) > 0 && s.Values[0][i] != nil {
										val, _ := s.Values[0][i].(json.Number).Float64()
										fields[c] = val
									}
								}
							}
						}
					}
					tags := map[string]string{"host_id": tg.HostId, "process": tg.Process}
					err = influx.WritesPoints("process_"+Duration, tags, fields)
				}
			}
		} else {
			break
		}
	}
}

func CustomTrend(Duration string) {
	var (
		tg     = monitor_conf.SystemTags{}
		influx = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		monitor_conf.Cch <- Duration
	}()
	// 监控数据趋势计算
	for {
		v := rc.SPop(ctx, monitor_conf.DataTrendKey+"_custom_"+Duration).Val()
		if v != "" {
			err = json.Unmarshal([]byte(v), &tg)
			if err == nil {
				cmd := "select " + monitor_conf.DurationMeasurement[Duration] +
					" from custom_" + monitor_conf.Measurements[Duration] +
					" where time > now() - " + Duration + " and host_id='" + tg.HostId + "'"
				res, err := influx.Query(cmd, true)
				if err == nil && len(res) > 0 {
					fields := map[string]interface{}{}
					for _, r := range res {
						for _, s := range r.Series {
							for i, d := range s.Columns {
								if strings.Contains(d, "max_") || strings.Contains(d, "mean_") {
									for _, p := range []string{"max_", "mean_"} {
										d = strings.Replace(d, p, "", 1)
									}
									if len(s.Values) > 0 && s.Values[0][i] != nil {
										val, _ := s.Values[0][i].(json.Number).Float64()
										fields[d] = val
									}
								}
							}
						}
					}
					tags := map[string]string{"host_id": tg.HostId, "source": tg.Source}
					err = influx.WritesPoints("custom_"+Duration, tags, fields)
				}
				if err != nil {
					Log.Error(err)
				}
			}
		} else {
			break
		}
	}
}

func Alive() {
	var (
		monitorAgentRun int
		AgentStatus     float64
		TraceId         string
		AgentConf       []databases.AgentConf
		MonitorProcess  []databases.MonitorProcess
		AgentAlive      []databases.AgentAlive
		AlarmHistory    []databases.AlarmHistory
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	//判断监控是否关闭
	if rc.Exists(ctx, platform_conf.AgentConfMonitor).Val() == 1 &&
		rc.Exists(ctx, platform_conf.AgentConfStatus).Val() == 1 {
		monitorAgentRun, _ = strconv.Atoi(rc.Get(ctx, platform_conf.AgentConfMonitor).Val())
		Status, _ := strconv.Atoi(rc.Get(ctx, platform_conf.AgentConfStatus).Val())
		AgentStatus = float64(Status)
	} else {
		db.Find(&AgentConf)
		if len(AgentConf) > 0 {
			rc.Set(ctx, platform_conf.AgentConfMonitor, AgentConf[0].AssetAgentRun, 5*time.Minute)
			rc.Set(ctx, platform_conf.AgentConfStatus, AgentConf[0].Status, 5*time.Minute)
		}
	}
	//处理异常报警记录
	go func() {
		db.Select("host_id").Where("status=?", "fault").Find(&AlarmHistory)
		if len(AlarmHistory) > 0 {
			for _, v := range AlarmHistory {
				if !rc.HExists(ctx, platform_conf.AgentAliveKey, v.HostId).Val() ||
					rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+v.HostId).Val() == 0 {
					db.Model(databases.AlarmHistory{}).Where("host_id=? and status=?",
						v.HostId, "fault").Updates(databases.AlarmHistory{Status: "unknown"})
				}
			}
		}
	}()
	//判断agent是否超时
	go func() {
		for hostId, onlineTime := range rc.HGetAll(ctx, platform_conf.AgentAliveKey).Val() {
			if rc.Exists(ctx, platform_conf.UpgradeKey+hostId).Val() == 0 {
				offlineTime := carbon.Parse(onlineTime).DiffInMinutes(carbon.Time2Carbon(time.Now()))
				if offlineTime > 1 {
					db.Model(&databases.AgentAlive{}).Where("host_id=?", hostId).Updates(
						databases.AgentAlive{OfflineTime: offlineTime})
					rc.HSet(ctx, platform_conf.OfflineAssetKey, hostId, offlineTime)
					platform_conf.WscPools.Delete(hostId)
				}
			}
		}
	}()
	for hostId := range rc.HGetAll(ctx, platform_conf.OfflineAssetKey).Val() {
		if monitorAgentRun == 1 && int(AgentStatus) == 1 {
			//心跳检测异常，主机不可达
			if rc.HExists(ctx, platform_conf.AgentAliveKey, hostId).Val() {
				db.Where("host_id=? and offline_time>?", hostId, 0).Find(&AgentAlive)
				if len(AgentAlive) > 0 {
					var ok bool
					if rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId).Val() == 0 {
						//新报警记录写入数据库
						TraceId = kits.RandString(12)
						da := databases.AlarmHistory{StartTime: time.Now(), EndTime: time.Now(),
							MonitorResource: "server", MonitorItem: "system",
							AlarmLevel: "Error", Content: "agent运行异常,主机不可达", Duration: 0,
							RuleId: "BuiltInRule", RuleName: "内置规则",
							RuleType: "BuiltInRule", HostId: hostId, Status: "fault",
							TraceId: TraceId}
						if err = db.Create(&da).Error; err == nil {
							ok = true
							rc.Set(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId, TraceId, 0)
							Log.Info("HostId:" + hostId + " TraceId:" + TraceId + "新报警记录写入数据库")
						}
					} else {
						TraceId = rc.Get(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId).Val()
						db.Where("trace_id=? and status=?", TraceId, "fault").Find(&AlarmHistory)
						if len(AlarmHistory) > 0 {
							db.Model(&AlarmHistory).Where("trace_id=? and status=?",
								TraceId, "fault").Updates(databases.AlarmHistory{EndTime: time.Now(),
								Duration: cast.ToInt64(time.Now().Sub(AlarmHistory[0].StartTime).Seconds())})
							ok = true
						}
					}
					if ok {
						monitor_conf.Sch <- monitor_conf.SendMsg{Status: "fault", HostId: hostId,
							MonitorResource: "server", RuleId: "BuiltInRule",
							MonitorItem: "system", TraceId: TraceId, MonitorValue: 0, MonitorInterval: 30,
							AlarmContent: "agent运行异常,主机不可达"}
						Log.Info("检测主机:" + hostId + " agent运行异常,主机不可达")
					}
				} else {
					if rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId).Val() == 1 {
						//心跳检测恢复，主机已可达
						TraceId := rc.Get(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId).Val()
						db.Where("host_id=? and offline_time=?", hostId, 0).Find(&AgentAlive)
						if len(AgentAlive) > 0 {
							db.Where("trace_id=?", TraceId).Find(&AlarmHistory)
							if len(AlarmHistory) > 0 {
								db.Model(&AlarmHistory).Where("trace_id=? and status=?", TraceId,
									"fault").Updates(databases.AlarmHistory{EndTime: time.Now(),
									Duration: cast.ToInt64(time.Now().Sub(
										AlarmHistory[0].StartTime).Seconds()),
									Status: "recovery"})
								monitor_conf.Sch <- monitor_conf.SendMsg{Status: "recovery", HostId: hostId,
									RuleId: "BuiltInRule", MonitorResource: "server", MonitorItem: "system",
									TraceId: TraceId, MonitorValue: 1, MonitorInterval: 30,
									AlarmContent: "agent运行恢复,主机已可达"}
								Log.Info("检测主机:" + hostId + " agent运行恢复,主机已可达")
							}
							rc.Del(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId)
							rc.HDel(ctx, platform_conf.OfflineAssetKey, hostId)
						}
					} else {
						rc.HDel(ctx, platform_conf.OfflineAssetKey, hostId)
					}
					// 监控进程是否存活
					db.Select("process").Where("host_id=? and status=?",
						hostId, "active").Find(&MonitorProcess)
					if len(MonitorProcess) > 0 {
						hostName := rc.HGet(ctx, platform_conf.ServerNameKey, hostId).Val()
						for _, pv := range MonitorProcess {
							if rc.Exists(ctx, platform_conf.ProcessAliveKey+"_"+hostId+"_"+pv.Process).Val() == 0 {
								//异常进程通知
								if rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process).Val() == 0 {
									//新报警记录写入数据库
									TraceId := kits.RandString(12)
									da := databases.AlarmHistory{StartTime: time.Now(), EndTime: time.Now(),
										MonitorResource: "process", MonitorItem: pv.Process,
										AlarmLevel: "Error", Duration: 0,
										Content: "检测主机:" + hostName + " 进程:" + pv.Process + "运行异常",
										RuleId:  "BuiltInRule", RuleName: "内置规则",
										RuleType: "BuiltInRule", HostId: hostId, Status: "fault",
										TraceId: TraceId}
									if err = db.Create(&da).Error; err == nil {
										rc.Set(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process, TraceId, 0)
										Log.Info("HostId:" + hostName + " TraceId:" + TraceId + "新报警记录写入数据库")
									}
								} else {
									TraceId := rc.Get(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process).Val()
									db.Model(&AlarmHistory).Where("trace_id=? and status=?",
										TraceId, "fault").Updates(databases.AlarmHistory{EndTime: time.Now(),
										Duration: cast.ToInt64(time.Now().Sub(AlarmHistory[0].StartTime).Seconds())})
								}
								TraceId := rc.Get(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process).Val()
								monitor_conf.Sch <- monitor_conf.SendMsg{Status: "fault", HostId: hostId,
									MonitorResource: "process", TraceId: TraceId,
									MonitorItem: pv.Process, MonitorValue: 0, MonitorInterval: 30,
									AlarmContent: "检测主机:" + hostName + " 进程:" + pv.Process + "运行异常"}
								Log.Info("检测主机:" + hostName + " 进程:" + pv.Process + "运行异常")
							} else {
								rc.Expire(ctx, platform_conf.ProcessAliveKey+"_"+hostId+"_"+pv.Process, 3*time.Minute)
								if rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process).Val() == 1 {
									//进程恢复通知
									TraceId := rc.Get(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process).Val()
									db.Model(&AlarmHistory).Where("trace_id=? and status=?",
										TraceId, "fault").Updates(databases.AlarmHistory{EndTime: time.Now(),
										Duration: cast.ToInt64(time.Now().Sub(
											AlarmHistory[0].StartTime).Seconds()),
										Status: "recovery"})
									monitor_conf.Sch <- monitor_conf.SendMsg{Status: "recovery", HostId: hostId,
										MonitorResource: "process", TraceId: TraceId,
										MonitorItem: pv.Process, MonitorValue: 1, MonitorInterval: 30,
										AlarmContent: "检测主机:" + hostName + " 进程:" + pv.Process + "运行已恢复"}
									rc.Del(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId+"_"+pv.Process)
									Log.Info("检测主机:" + hostName + " 进程:" + pv.Process + "运行已恢复")
								}
							}
						}
					}
				}
			}
		}
	}
}

func CleanServer() {
	lock := common.SyncMutex{LockKey: "monitor_clean_server_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				var (
					sqlErr         error
					AlarmHistory   []databases.AlarmHistory
					AlarmSend      []databases.AlarmSend
					GroupServer    []databases.GroupServer
					MonitorProcess []databases.MonitorProcess
					MonitorJobs    []databases.MonitorJobs
					hostIds        []string
				)
				Log.Info("Monitor_clean_server task start working ......")
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
					lock.UnLock(true)
				}()
				db.Select("host_id").Find(&GroupServer)
				if len(GroupServer) > 0 {
					for _, v := range GroupServer {
						if rc.HExists(ctx, platform_conf.DiscardAssetKey, v.HostId).Val() {
							hostIds = append(hostIds, v.HostId)
						}
					}
					if len(hostIds) > 0 {
						err = db.Transaction(func(tx *gorm.DB) error {
							if sqlErr = tx.Where("host_id in ?", hostIds).Delete(&MonitorProcess).Error; sqlErr != nil {
								err = sqlErr
							}
							if sqlErr = tx.Where("host_id in ?", hostIds).Delete(&AlarmHistory).Error; sqlErr != nil {
								err = sqlErr
							}
							if sqlErr = tx.Where("host_id in ?", hostIds).Delete(&AlarmSend).Error; sqlErr != nil {
								err = sqlErr
							}
							return sqlErr
						})
					}
				}
				//清除已删除脚本关联信息
				sc := rc.SMembers(ctx, platform_conf.DeleteScriptsKey).Val()
				if len(sc) > 0 {
					for _, v := range sc {
						db.Where("script_id=?", v).Delete(&MonitorJobs)
						rc.SRem(ctx, platform_conf.DeleteScriptsKey, v)
					}
				}
			}()
		}
		time.Sleep(5 * time.Minute)
	}
}

func MonitorOverView() {
	lock := common.SyncMutex{LockKey: "sync_monitor_overview_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				var (
					AlarmHosts int64
				)
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
					lock.UnLock(true)
				}()
				Log.Info("监控总览数据抓取任务开始执行......")
				db.Model(&databases.AlarmHistory{}).Distinct("host_id").Where("status=?", "fault").Count(&AlarmHosts)
				rc.HSet(ctx, platform_conf.OverViewKey, "alarm_hosts", AlarmHosts)
				Log.Info("计算当日报警主机数量" + cast.ToString(AlarmHosts) + "写入缓存......")
			}()
		}
		time.Sleep(1 * time.Minute)
	}
}

func ServerHealth() {
	lock := common.SyncMutex{LockKey: "monitor_server_health_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				AlarmHistory []databases.AlarmHistory
				GroupServer  []databases.GroupServer
			)
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
				lock.UnLock(true)
			}()
			Log.Info("服务器健康状态抓取任务开始执行......")
			db.Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, v := range GroupServer {
					db.Select("host_id").Where("host_id=? and status=?", v.HostId, "fault").Find(&AlarmHistory)
					if len(AlarmHistory) > 0 {
						rc.Set(ctx, platform_conf.ServerHealthKey+v.HostId, "fault", 3*time.Minute)
					} else {
						rc.Set(ctx, platform_conf.ServerHealthKey+v.HostId, "health", 3*time.Minute)
					}
				}
			}
		}()
	}
}

func MonitorHandle() {
	//监控数据上报
	var (
		Encrypt        = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		influx         = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
		ScriptContents []databases.ScriptContents
	)
	for {
		data := <-platform_conf.Mch
		if data != nil {
			jd := kits.StringToMap(data["monitor"].(string))
			if jd["tags"] != nil && jd["fields"] != nil {
				t := jd["tags"].(map[string]interface{})
				f := jd["fields"].(map[string]interface{})
				pt := jd["process_top"].(map[string]interface{})
				hostId := t["host_id"].(string)
				if len(t) > 0 {
					// 系统监控数据写入influxdb
					tags := map[string]string{"host_id": hostId, "source_type": "agent"}
					fields := map[string]interface{}{}
					for k, v := range f {
						fields[k] = v.(float64)
						// 监控数据写入报警引擎channel
						monitor_conf.Ach <- monitor_conf.AlarmData{HostId: hostId,
							MonitorResource: "server",
							MonitorItem:     "system",
							MonitorKey:      k,
							MonitorValue:    v.(float64),
							MonitorInterval: cast.ToInt32(jd["MonitorInterval"])}
					}
					if len(fields) > 0 {
						err = influx.WritesPoints("system_1m", tags, fields)
					}
					// 监控数据趋势计算
					for _, i := range monitor_conf.Duration {
						m, _ := json.Marshal(tags)
						rc.SAdd(ctx, monitor_conf.DataTrendKey+"_system_"+i, string(m))
					}
					// 初始化数据库连接
					var p []string
					key := "monitor_process_" + hostId
					// 缓存结果3分钟减轻数据库压力
					if rc.Exists(ctx, key).Val() == 1 {
						p = rc.LRange(ctx, key, 0, -1).Val()
					} else {
						var MonitorProcess []databases.MonitorProcess
						// 获取主机进程监控配置
						db.Select("process").Where("host_id=? and status=?",
							hostId, "active").Find(&MonitorProcess)
						if len(MonitorProcess) > 0 {
							for _, d := range MonitorProcess {
								p = append(p, d.Process)
								rc.LPush(ctx, key, d.Process)
							}
						}
					}
					rc.Expire(ctx, key, 3*time.Minute)
					Res := map[string]interface{}{}
					if len(p) > 0 {
						Res["process"] = p
						_, ok := jd["process"]
						if ok && jd["process"] != nil {
							// 进程监控数据写入influxdb
							for k, v := range jd["process"].(map[string]interface{}) {
								tags = map[string]string{"host_id": hostId, "process": k}
								fields = map[string]interface{}{}
								for pk, pv := range v.(map[string]interface{}) {
									fields[pk] = pv.(float64)
									// 监控数据写入报警引擎channel
									monitor_conf.Ach <- monitor_conf.AlarmData{HostId: hostId,
										MonitorResource: "process", MonitorItem: k, MonitorKey: pk,
										MonitorValue:    pv.(float64),
										MonitorInterval: cast.ToInt32(jd["MonitorInterval"])}
								}
								if len(fields) > 0 {
									err = influx.WritesPoints("process_1m", tags, fields)
								}
								// 监控数据趋势计算
								for _, i := range monitor_conf.Duration {
									m, _ := json.Marshal(tags)
									rc.SAdd(ctx, monitor_conf.DataTrendKey+"_process_"+i, string(m))
								}
								_, err = rc.Set(ctx, platform_conf.ProcessAliveKey+"_"+hostId+"_"+k, "alive",
									time.Duration(cast.ToInt32(jd["MonitorInterval"])*3)*time.Second).Result()
							}
						}
					}
					//自定义配置监控
					var (
						CustomMetrics []databases.CustomMetrics
						GroupMetrics  []databases.GroupMetrics
						GroupServer   []databases.GroupServer
					)
					db.Where("host_id=?", hostId).Find(&GroupServer)
					if len(GroupServer) > 0 {
						_, ok := jd["custom_metrics"]
						if ok && jd["custom_metrics"] != nil {
							// 自定义监控数据写入influxdb
							fields = map[string]interface{}{}
							tags = map[string]string{"host_id": hostId, "source_type": "agent"}
							for k, v := range jd["custom_metrics"].(map[string]interface{}) {
								fields[k] = v.(float64)
								// 监控数据写入报警引擎channel
								monitor_conf.Ach <- monitor_conf.AlarmData{HostId: hostId,
									MonitorResource: "server", MonitorItem: "custom", MonitorKey: k,
									MonitorValue:    v.(float64),
									MonitorInterval: cast.ToInt32(jd["MonitorInterval"])}
							}
							if len(fields) > 0 {
								err = influx.WritesPoints("custom_1m", tags, fields)
							}
							// 监控数据趋势计算
							for _, i := range monitor_conf.Duration {
								m, _ := json.Marshal(tags)
								rc.SAdd(ctx, monitor_conf.DataTrendKey+"_custom_"+i, string(m))
							}
						}
						var groupIds []string
						for _, v := range GroupServer {
							groupIds = append(groupIds, v.GroupId)
						}
						ck := "custom_metrics_" + hostId
						cm := map[string]string{}
						if rc.Exists(ctx, ck).Val() == 1 {
							cm = rc.HGetAll(ctx, ck).Val()
						} else {
							db.Where("group_id in ?", groupIds).Find(&GroupMetrics)
							if len(GroupMetrics) > 0 {
								for _, v := range GroupMetrics {
									var path string
									db.Where("monitor_key=?", v.MonitorKey).First(&CustomMetrics)
									if len(CustomMetrics) > 0 {
										db.Where("script_id=?", CustomMetrics[0].ScriptId).First(&ScriptContents)
										if len(ScriptContents) > 0 {
											if strings.Contains(ScriptContents[0].ScriptName, ".sh") {
												path = "sh " + platform_conf.AgentRoot + "/script/" + ScriptContents[0].ScriptName
											}
											if strings.Contains(ScriptContents[0].ScriptName, ".py") {
												path = "python3 " + platform_conf.AgentRoot + "/script/" + ScriptContents[0].ScriptName
											}
										}
									}
									if path != "" {
										cm[v.MonitorKey] = path
										rc.HSet(ctx, ck, v.MonitorKey, path)
										rc.Expire(ctx, ck, 3*time.Minute)
									}
								}
							}
						}
						if len(cm) > 0 {
							Res["custom_metrics"] = cm
						}
					}
					if len(Res) > 0 {
						Res["host_id"] = hostId
					}
					if len(pt["cpu"].(map[string]interface{})) > 0 {
						key := monitor_conf.ProcessTop + "_cpu_" + hostId
						rc.Expire(ctx, key, 5*time.Minute)
						rc.HSet(ctx, key, pt["cpu"].(map[string]interface{}))
					}
					if len(pt["mem"].(map[string]interface{})) > 0 {
						key := monitor_conf.ProcessTop + "_mem_" + hostId
						rc.Expire(ctx, key, 5*time.Minute)
						rc.HSet(ctx, key, pt["mem"].(map[string]interface{}))
					}
					//监控配置下发
					m := Encrypt.EncryptString(kits.MapToJson(Res), true)
					platform_conf.Wch <- map[string]interface{}{"monitor": kits.MapToJson(map[string]interface{}{"monitor": m}),
						"host_id": hostId, "msg_time": time.Now().Format("2006-01-02 15:04:05")}
				}
			}
		}
	}
}
