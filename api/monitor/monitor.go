package monitor

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"inner/conf/monitor_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strings"
)

// @Tags 监控平台
// @Summary 聚合数据查询
// @Produce  json
// @Security ApiKeyAuth
// @Param host_ids query array true "主机ID"
// @Param resource query string true "监控资源"
// @Param item query string true "监控项"
// @Param converge query string true "汇聚参数(max|min|mean)"
// @Param key query string false "监控指标"
// @Param duration query int false "最近时间段(分钟)"
// @Param start_time query string false "起始时间(000-00-00 00:00:00)"
// @Param end_time query string false "结束时间(000-00-00 00:00:00)"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/data/converge [get]
func ConvergeData(c *gin.Context) {
	//聚合数据查询接口
	var (
		JsonData monitor_conf.DataConverge
		Response = common.Response{C: c}
		cmd      string
		cmdExt   string
		influx   = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
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
		Data := map[string]interface{}{}
		if JsonData.Key == "" {
			JsonData.Key = "*"
		}
		if JsonData.Duration > 0 {
			if JsonData.Duration > 129600 {
				err = errors.New("只允许查询近90天聚合数据")
				panic(nil)
			}
			cmdExt = "  and time > now() - " + cast.ToString(JsonData.Duration) + "m"
		} else {
			if !JsonData.StartTime.IsZero() && !JsonData.EndTime.IsZero() {
				JsonData.Duration = carbon.Time2Carbon(JsonData.StartTime).DiffInMinutes(
					carbon.Time2Carbon(JsonData.EndTime))
				if JsonData.Duration <= 0 {
					err = errors.New("结束时间要大于开始时间")
					panic(nil)
				}
				if carbon.Time2Carbon(JsonData.EndTime).DiffInDays(carbon.Time2Carbon(JsonData.StartTime)) > 30 {
					err = errors.New("只允许查询近30天聚合数据")
					panic(nil)
				}
				cmdExt = " and time >= '" + carbon.Time2Carbon(JsonData.StartTime).ToDateTimeString() + "'" +
					" and time <= '" + carbon.Time2Carbon(JsonData.EndTime).ToDateTimeString() + "'"
			} else {
				cmdExt = " and time > now() - 5m"
			}
		}
		if JsonData.Converge != "" {
			JsonData.Key = JsonData.Converge + "(" + JsonData.Key + ")"
		}
		if JsonData.Resource == "server" {
			cmd = "select " + JsonData.Key + " from system_1m" + " where"
			if JsonData.Item == "custom" {
				cmd = "select " + JsonData.Key + " from custom_1m" + " where"
			}
		}
		if JsonData.Resource == "process" {
			cmd = "select " + JsonData.Key + " from process_1m" + " where" + " process=" + JsonData.Item + " and"
		}
		for _, v := range kits.FormListFormat(JsonData.HostIds) {
			// Log.Info(cmd + " host_id=" + "'" + v + "'" + cmdExt)
			res, err := influx.Query(cmd+" host_id="+"'"+v+"'"+cmdExt, true)
			if err == nil && len(res) > 0 {
				for _, r := range res {
					for _, s := range r.Series {
						fields := map[string]interface{}{}
						for i, d := range s.Columns {
							d = strings.Replace(d, JsonData.Converge+"_", "", 1)
							if d == "time" {
								fields[d] = carbon.Parse(s.Values[0][i].(string)).ToDateTimeString()
							} else {
								fields[d] = s.Values[0][i]
							}
						}
						Data[v] = fields
					}
				}
				Response.Data = Data
			}
		}
	}
}

// @Tags 监控平台
// @Summary 详情数据查询
// @Produce  json
// @Security ApiKeyAuth
// @Param host_id query string true "主机ID"
// @Param resource query string true "监控资源"
// @Param item query string true "监控项"
// @Param key query string false "监控指标"
// @Param duration query int false "最近时间段(分钟)"
// @Param start_time query string false "起始时间(000-00-00 00:00:00)"
// @Param end_time query string false "结束时间(000-00-00 00:00:00)"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/data/detail [get]
func DetailData(c *gin.Context) {
	//详情数据查询接口
	var (
		JsonData     monitor_conf.DataDetail
		GroupMetrics []databases.GroupMetrics
		Response     = common.Response{C: c}
		Duration     = "1m"
		cmd          string
		cmdExt       string
		influx       = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回结果
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
		if JsonData.Duration > 0 {
			if JsonData.Duration > 720 && JsonData.Duration <= 7200 {
				Duration = "5m"
			}
			if JsonData.Duration > 7200 {
				Duration = "1h"
			}
			cmdExt = "  and time > now() - " + cast.ToString(JsonData.Duration) + "m"
		} else {
			if JsonData.StartTime != "" && JsonData.EndTime != "" {
				JsonData.Duration = carbon.Parse(JsonData.StartTime).DiffInMinutes(
					carbon.Parse(JsonData.EndTime))
				if JsonData.Duration <= 0 {
					err = errors.New(cast.ToString(JsonData.Duration) + ",结束时间要大于开始时间")
					panic(err)
				}
				if JsonData.Duration > 720 && JsonData.Duration <= 7200 {
					Duration = "5m"
				}
				if JsonData.Duration > 7200 {
					Duration = "1h"
				}
				cmdExt = " and time >= '" + JsonData.StartTime + "'" +
					" and time <= '" + JsonData.EndTime + "'"
			} else {
				cmdExt = " and time > now() - 5m"
			}
		}
		if JsonData.Key == "" {
			JsonData.Key = "*"
		}
		if JsonData.Resource == "server" {
			cmd = "select " + JsonData.Key + " from system_" + Duration + " where host_id='" + JsonData.HostId + "'"
			if JsonData.Item == "custom" {
				cmd = "select " + JsonData.Key + " from custom_" + Duration + " where host_id='" + JsonData.HostId + "'"
			}
		}
		if JsonData.Resource == "process" {
			cmd = "select " + JsonData.Key + " from process_" + Duration + " where host_id='" + JsonData.HostId +
				"' and process='" + JsonData.Item + "'"
		}
		sql := "join group_server on group_server.group_id=group_metrics.group_id and group_server.host_id=?"
		db.Joins(sql, JsonData.HostId).Find(&GroupMetrics)
		metrics := map[string]struct{}{}
		if len(GroupMetrics) > 0 {
			for _, v := range GroupMetrics {
				metrics[v.MonitorKey] = struct{}{}
			}
		}
		Wan := rc.Exists(ctx, platform_conf.HostWanKey+JsonData.HostId).Val()
		//Log.Info(cmd + cmdExt)
		res, err := influx.Query(cmd+cmdExt, true)
		if err == nil && len(res) > 0 {
			var data []interface{}
			for _, r := range res {
				for _, s := range r.Series {
					for _, v := range s.Values {
						fields := map[string]interface{}{}
						for i, d := range v {
							if s.Columns[i] == "time" {
								if JsonData.Duration <= 720 {
									fields[s.Columns[i]] = carbon.Parse(d.(string)).ToTimeString()
								} else {
									fields[s.Columns[i]] = carbon.Parse(d.(string)).ToDateTimeString()
								}
							} else {
								if JsonData.Item == "custom" {
									_, ok := metrics[s.Columns[i]]
									if ok {
										fields[s.Columns[i]] = d
									}
								} else {
									if strings.HasPrefix(s.Columns[i], "wan_") {
										if Wan == 1 {
											fields[s.Columns[i]] = d
										}
									} else {
										fields[s.Columns[i]] = d
									}
								}
							}
						}
						data = append(data, fields)
					}
				}
			}
			Response.Data = data
		}
	}
}
