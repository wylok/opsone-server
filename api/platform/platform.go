package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags 平台管理
// @Summary 主机离线时间
// @Produce  json
// @Security ApiKeyAuth
// @Param host_id query []string true "主机id列表"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/platform/offline_time [get]
func OfflineTime(c *gin.Context) {
	//主机离线时间
	var (
		JsonData   = platform_conf.OfflineTime{}
		Response   = common.Response{C: c}
		AgentAlive []databases.AgentAlive
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("host_id in ?", JsonData.HostIds).Find(&AgentAlive)
		if len(AgentAlive) > 0 {
			Response.Data = AgentAlive
		}
	}
}

// @Tags 平台管理
// @Summary 配置管理
// @Produce  json
// @Security ApiKeyAuth
// @Param name query string true "配置名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /v1/platform/config [get]
func Config(c *gin.Context) {
	//配置管理
	var (
		JsonData platform_conf.PlatformConfig
		Response = common.Response{C: c}
		cf       = platform_conf.Setting()
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Name != "" {
			d, _ := json.Marshal(cf)
			v := map[string]interface{}{}
			_ = json.Unmarshal(d, &v)
			Response.Data = map[string]interface{}{JsonData.Name: v[JsonData.Name]}
		}
		Response.Data = cf
	}
}
