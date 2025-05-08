package k8s

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"inner/conf/k8s_conf"
	"inner/modules/common"
	"inner/modules/kits"
)

// @Tags k8s集群
// @Summary MetricChart
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param resource query string true "资源"
// @Param name query string true "资源名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/metric [get]
func MetricChart(c *gin.Context) {
	//MetricChart
	var (
		JsonData k8s_conf.K8sMetric
		Response = common.Response{C: c}
		influx   = common.InfluxDb{Cli: Cli, Database: "opsone_k8s"}
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
	if err == nil && JsonData.Resource != "" {
		cmd := "SELECT cpu,memory FROM " + JsonData.Resource + "_1m WHERE time > now() - 30m and k8s_id='" + JsonData.K8sID + "'"
		switch JsonData.Resource {
		case "node":
			cmd = cmd + " and " + "node_name='" + JsonData.Name + "'"
		case "pod":
			cmd = cmd + " and " + "pod_name='" + JsonData.Name + "'"
			if JsonData.NameSpace != "" {
				cmd = cmd + " and " + "name_space='" + JsonData.NameSpace + "'"
			}
		}
		res, err := influx.Query(cmd, true)
		if err == nil && len(res) > 0 {
			var Data []map[string]any
			for _, r := range res {
				for _, s := range r.Series {
					for _, va := range s.Values {
						fields := map[string]any{}
						for i, v := range va {
							if s.Columns[i] == "time" {
								fields[s.Columns[i]] = carbon.Parse(v.(string)).ToTimeString()
							} else {
								fields[s.Columns[i]] = v
							}
						}
						Data = append(Data, fields)
					}
				}
			}
			Response.Data = Data
		}
	}
}

// @Tags k8s集群
// @Summary overView
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/overview [get]
func OverView(c *gin.Context) {
	//overView
	var (
		err      error
		Response = common.Response{C: c}
	)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if rc.Exists(ctx, "k8s_overview").Val() == 1 {
		Response.Data = kits.StringToMap(rc.Get(ctx, "k8s_overview").Val())
	}
}
