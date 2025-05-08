package monitor

import "C"
import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"inner/conf/monitor_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
	"time"
)

// @Tags 监控平台
// @Summary 报警次数统计查询
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/alarm/count [get]
func AlarmCount(c *gin.Context) {
	//报警次数统计查询接口
	var (
		Response     = common.Response{C: c}
		data         []map[string]int64
		ruleIds      []string
		err          error
		MonitorRules []databases.MonitorRules
	)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		for _, i := range []int{7, 6, 5, 4, 3, 2, 1} {
			var count int64
			db.Find(&MonitorRules)
			if len(MonitorRules) > 0 {
				for _, v := range MonitorRules {
					ruleIds = append(ruleIds, v.RuleId)
				}
				ruleIds = append(ruleIds, "BuiltInRule")
			}
			d := carbon.Time2Carbon(time.Now()).SubDays(i).ToDateString()
			db.Model(&databases.AlarmHistory{}).Where("rule_id in ?"+
				" and alarm_history.start_time like ?", ruleIds, d+" %").Count(&count)
			data = append(data, map[string]int64{strings.Join(strings.Split(d, "-")[1:], "-"): count})
		}
		Response.Data = data
	}
}

// @Tags 监控平台
// @Summary 监控TOP查询
// @Produce  json
// @Security ApiKeyAuth
// @Param item query string true "监控项"
// @Param metric query string true "监控指标"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/metric/top [get]
func MetricTop(c *gin.Context) {
	//监控TOP查询
	var (
		JsonData    monitor_conf.MetricTop
		Response    = common.Response{C: c}
		influx      = common.InfluxDb{Cli: Cli, Database: "opsone_monitor"}
		data        []map[string]interface{}
		AssetServer []databases.AssetServer
		hostIds     []string
		hostNames   = map[string]string{}
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
		cmd := "select top(" + JsonData.Metric + ",50),host_id from " + JsonData.Item + "_5m where time > now() - 5m"
		res, err := influx.Query(cmd, true)
		if err == nil && len(res) > 0 {
			hm := map[string]struct{}{}
			for _, r := range res {
				for _, s := range r.Series {
					for _, v := range s.Values {
						_, ok := hm[v[2].(string)]
						if !ok && len(hm) <= 7 {
							hm[v[2].(string)] = struct{}{}
							data = append(data, map[string]interface{}{s.Columns[1]: v[1], s.Columns[2]: v[2]})
							hostIds = append(hostIds, v[2].(string))
						}
					}
				}
			}
			db.Where("host_id in ?", hostIds).Find(&AssetServer)
			if len(AssetServer) > 0 {
				for _, v := range AssetServer {
					hostNames[v.HostId] = v.Hostname
				}
			}
			var Da []map[string]interface{}
			if len(hostNames) > 0 {
				for _, d := range data {
					d["host_name"] = hostNames[d["host_id"].(string)]
					Da = append(Da, d)
				}
			} else {
				for _, d := range data {
					d["host_name"] = d["host_id"]
					Da = append(Da, d)
				}
			}
			if len(Da) > 0 {
				Response.Data = Da
			}
		}
	}
}
