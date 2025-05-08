package platform

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"time"
)

// @Tags 平台管理
// @Summary 查询Agent配置
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:null}"
// @Router /api/v1/platform/agent/conf [get]
func AgentConfig(c *gin.Context) {
	//查询Agent配置
	var (
		sqlErr    error
		JsonData  = platform_conf.AgentConf{}
		Response  = common.Response{C: c}
		AgentConf []databases.AgentConf
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 10
		}
		p := databases.Pagination{DB: db, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&AgentConf)
	}
}

// @Tags 平台管理
// @Summary 修改Agent配置
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  platform_conf.UpdateAgentConf true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/platform/agent/conf [put]
func ModifyAgent(c *gin.Context) {
	//修改Agent配置
	var (
		sqlErr    error
		JsonData  = platform_conf.UpdateAgentConf{}
		Response  = common.Response{C: c}
		AgentConf []databases.AgentConf
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("id=?", JsonData.Id).Find(&AgentConf)
		if len(AgentConf) > 0 {
			if JsonData.AssetInterval == 0 {
				JsonData.AssetInterval = AgentConf[0].AssetInterval
			}
			if JsonData.HeartBeatInterval == 0 {
				JsonData.HeartBeatInterval = AgentConf[0].HeartBeatInterval
			}
			if JsonData.MonitorInterval == 0 {
				JsonData.MonitorInterval = AgentConf[0].MonitorInterval
			}
			updates := map[string]interface{}{
				"asset_agent_run":    cast.ToInt(JsonData.AssetAgentRun),
				"monitor_agent_run":  cast.ToInt(JsonData.MonitorAgentRun),
				"asset_interval":     JsonData.AssetInterval,
				"heartbeat_interval": JsonData.HeartBeatInterval,
				"monitor_interval":   JsonData.MonitorInterval,
				"status":             cast.ToInt(JsonData.Status),
			}
			err = db.Model(&databases.AgentConf{}).Where("id=?", JsonData.Id).Updates(updates).Error
		}
	}
}

// @Tags 平台管理
// @Summary Agent在线
// @Produce  json
// @Security ApiKeyAuth
// @Param host_name query string false "主机名称"
// @Param agent_version query string false "agent版本"
// @Param host_id query string false "主机ID"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:null}"
// @Router /api/v1/platform/agent/alive [get]
func AgentAlive(c *gin.Context) {
	//Agent在线
	var (
		sqlErr      error
		JsonData    = platform_conf.AgentAlive{}
		Response    = common.Response{C: c}
		AgentAlive  []databases.AgentAlive
		AssetServer []databases.AssetServer
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		if sqlErr != nil {
			err = sqlErr
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
		tx := db.Order("offline_time desc")
		if JsonData.HostName != "" {
			var hostIds []string
			db.Select("host_id").Where("host_name like ?", "%"+JsonData.HostName+"%").Find(&AssetServer)
			if len(AssetServer) > 0 {
				for _, v := range AssetServer {
					hostIds = append(hostIds, v.HostId)
				}
			}
			tx = tx.Where("host_id in ?", hostIds)
		}
		if JsonData.AgentVersion != "" {
			tx = tx.Where("agent_version like ?", "%"+JsonData.AgentVersion+"%")
		}
		if JsonData.HostId != "" {
			tx = tx.Where("host_id like ?", "%"+JsonData.HostId+"%")
		}
		if JsonData.ClamAv == "clamAv" {
			tx = tx.Where("clamAv = ?", "clamAv")
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AgentAlive)
		if len(AgentAlive) > 0 {
			var Data []map[string]interface{}
			for _, v := range AgentAlive {
				Data = append(Data, map[string]interface{}{"host_id": v.HostId,
					"host_name":     rc.HGet(ctx, platform_conf.ServerNameKey, v.HostId).Val(),
					"agent_version": v.AgentVersion, "offline_time": v.OfflineTime, "clamAv": v.ClamAv,
					"clamRun": v.ClamRun})
			}
			Response.Data = Data
		}
	}
}

// @Tags 平台管理
// @Summary 删除离线Agent
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  platform_conf.DeleteAgentAlive true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/platform/agent/alive [delete]
func DeleteAgentAlive(c *gin.Context) {
	//删除离线Agent
	var (
		sqlErr     error
		JsonData   = platform_conf.DeleteAgentAlive{}
		Response   = common.Response{C: c}
		AgentAlive []databases.AgentAlive
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		for _, hostId := range JsonData.HostIds {
			rc.HSet(ctx, platform_conf.DiscardAssetKey, hostId, "")
			rc.HDel(ctx, platform_conf.AgentAliveKey, hostId)
			rc.HDel(ctx, platform_conf.OfflineAssetKey, hostId)
		}
		rc.Expire(ctx, platform_conf.DiscardAssetKey, 30*time.Minute)
		db.Where("host_id in ?", JsonData.HostIds).Delete(&AgentAlive)
	}
}
