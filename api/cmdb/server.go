package cmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/cmdb_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"net"
)

// @Tags 资产主机
// @Summary 主机信息查询
// @Produce  json
// @Security ApiKeyAuth
// @Param department_id query string false "部门ID"
// @Param asset_group_id query string false "资源组ID"
// @Param host_ids query array false "主机ID列表"
// @Param host_name query string false "主机名称"
// @Param host_type query string false "主机类型"
// @Param sn query string false "主机sn"
// @Param ip query string false "主机IP"
// @Param asset_tag query string false "资产标签"
// @Param status query string false "主机状态"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/servers [get]
func QueryServer(c *gin.Context) {
	//主机信息查询
	var (
		JsonData    cmdb_conf.Server
		AssetServer []databases.AssetServer
		AssetNet    []databases.AssetNet
		AssetDisk   []databases.AssetDisk
		AssetExtend []databases.AssetExtend
		AssetUnder  []databases.AssetUnder
		AssetIdc    []databases.AssetIdc
		GroupServer []databases.GroupServer
		AgentAlive  []databases.AgentAlive
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
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
		JsonData.HostIds = kits.FormListFormat(JsonData.HostIds)
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
			if JsonData.HostIds != nil {
				JsonData.PerPage = len(JsonData.HostIds)
			}
		}
		tx := db.Where("asset_server.asset_status=?", "assigned")
		if JsonData.DepartmentId != "" {
			sql := "join cmdb_partition on cmdb_partition.object_id = asset_server.host_id " +
				"and cmdb_partition.object_type=? and cmdb_partition.department_id=?"
			tx = tx.Joins(sql, "server", JsonData.DepartmentId)
		}
		// 部分参数匹配
		if JsonData.AssetGroupId != "" {
			var hostIds []string
			db.Where("group_id=?", JsonData.AssetGroupId).Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, v := range GroupServer {
					hostIds = append(hostIds, v.HostId)
				}
			}
			tx = tx.Where("asset_server.host_id in ?", hostIds)
		}
		if JsonData.Ip != "" {
			var hostIds []string
			db.Where("ip=?", JsonData.Ip).Find(&AssetNet)
			if len(AssetNet) > 0 {
				for _, v := range AssetNet {
					hostIds = append(hostIds, v.HostId)
				}
			}
			tx = tx.Where("asset_server.host_id in ?", hostIds)
		}
		if JsonData.HostType != "" {
			if JsonData.HostType == "offline" {
				var HostIds []string
				db.Where("offline_time > ?", 0).Find(&AgentAlive)
				if len(AgentAlive) > 0 {
					for _, v := range AgentAlive {
						HostIds = append(HostIds, v.HostId)
					}
				}
				tx = tx.Where("asset_server.host_id in ?", HostIds)
			} else {
				tx = tx.Where("asset_server.host_type_cn like ? or asset_server.host_type like ?",
					"%"+JsonData.HostType+"%", "%"+JsonData.HostType)
			}
		}
		if len(JsonData.HostIds) > 0 {
			tx = tx.Where("asset_server.host_id in ?", JsonData.HostIds)
		}
		if JsonData.HostName != "" {
			tx = tx.Where("asset_server.host_name like ?", "%"+JsonData.HostName+"%")
		}
		if JsonData.AssetTag != "" {
			tx = tx.Where("asset_server.asset_tag like ?", "%"+JsonData.AssetTag+"%")
		}
		if JsonData.SN != "" {
			tx = tx.Where("asset_server.sn = ?", JsonData.SN)
		}
		if JsonData.Status != "" {
			tx = tx.Where("asset_server.status = ?", JsonData.Status)
		}
		tx = tx.Order("asset_server.id desc")
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AssetServer)
		//附加资产配置信息
		if len(AssetServer) > 0 {
			var Data []map[string]interface{}
			for _, sv := range AssetServer {
				var (
					health = "unknown"
					online bool
					data   = map[string]interface{}{}
					disk   []map[string]interface{}
					nets   []map[string]interface{}
					extend = map[string]interface{}{}
					idc    = map[string]interface{}{}
					under  = map[string]interface{}{}
				)
				db.Where("host_id=?", sv.HostId).Find(&AssetDisk)
				db.Where("host_id=?", sv.HostId).Find(&AssetNet)
				db.Where("host_id=?", sv.HostId).Find(&AssetExtend)
				db.Where("asset_id=? and asset_type=?", sv.HostId, "server").Find(&AssetUnder)
				if len(AssetDisk) > 0 {
					for _, v := range AssetDisk {
						disk = append(disk, map[string]interface{}{"disk_name": v.DiskName, "mount_point": v.MountPoint,
							"fs_type": v.FsType, "disk_size": v.DiskSize})
					}
				}
				if len(AssetNet) > 0 {
					for _, v := range AssetNet {
						if netutil.IsInternalIP(net.ParseIP(v.Ip)) {
							nets = append(nets, map[string]interface{}{"name": v.Name, "addr": v.Addr,
								"ip": v.Ip, "netmask": v.Netmask})
						}
					}
					for _, v := range AssetNet {
						if netutil.IsPublicIP(net.ParseIP(v.Ip)) {
							nets = append(nets, map[string]interface{}{"name": v.Name, "addr": v.Addr,
								"ip": v.Ip, "netmask": v.Netmask})
						}
					}
				}
				if len(AssetExtend) > 0 {
					extend = map[string]interface{}{"idc_id": AssetExtend[0].IdcId, "ipmi": AssetExtend[0].Ipmi,
						"cabinet": AssetExtend[0].Cabinet, "buy_time": AssetExtend[0].BuyTime,
						"expired_time": AssetExtend[0].ExpiredTime}
					db.Where("idc_id=?", AssetExtend[0].IdcId).First(&AssetIdc)
					if len(AssetIdc) > 0 {
						d, _ := json.Marshal(AssetIdc[0])
						idc = cast.ToStringMap(string(d))
					}
				}
				if len(AssetUnder) > 0 {
					under = map[string]interface{}{"department_id": AssetUnder[0].DepartmentId,
						"business_id": AssetUnder[0].BusinessId}
				}
				if rc.Exists(ctx, platform_conf.ServerHealthKey+sv.HostId).Val() == 1 {
					health = rc.Get(ctx, platform_conf.ServerHealthKey+sv.HostId).Val()
				}
				if rc.HExists(ctx, platform_conf.AgentAliveKey, sv.HostId).Val() {
					if !rc.HExists(ctx, platform_conf.OfflineAssetKey, sv.HostId).Val() {
						online = true
					}
				}
				data = map[string]interface{}{"id": sv.Id, "host": sv, "net": nets, "disk": disk,
					"extend": extend, "idc": idc, "under": under, "health": health, "online": online}
				Data = append(Data, data)
			}
			Response.Data = Data
		}
	}
}

// @Tags 资产主机
// @Summary 修改主机信息
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.UpdateServer true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/servers [put]
func UpdateServer(c *gin.Context) {
	//修改主机信息
	var (
		sqlErr      error
		JsonData    cmdb_conf.UpdateServer
		AssetExtend []databases.AssetExtend
		AssetServer []databases.AssetServer
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
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
		err = db.Transaction(func(tx *gorm.DB) error {
			updates := databases.AssetExtend{}
			if JsonData.IdcId != "" {
				updates.IdcId = JsonData.IdcId
			}
			if JsonData.IdcId != "" {
				updates.IdcId = JsonData.IdcId
			}
			if JsonData.Cabinet != "" {
				updates.Cabinet = JsonData.Cabinet
			}
			if JsonData.Ipmi != "" {
				updates.Ipmi = JsonData.Ipmi
			}
			if JsonData.BuyTime != "" {
				updates.BuyTime = carbon.Parse(JsonData.BuyTime).Carbon2Time()
			}
			if JsonData.ExpiredTime != "" {
				updates.ExpiredTime = carbon.Parse(JsonData.ExpiredTime).Carbon2Time()
			}
			update := databases.AssetServer{}
			if JsonData.HostType != "" {
				update.HostType = JsonData.HostType
			}
			if JsonData.AssetTag != "" {
				update.AssetTag = JsonData.AssetTag
			}
			if JsonData.NickName != "" {
				update.NickName = JsonData.NickName
			}
			db.Where("host_id = ?", JsonData.HostId).Find(&AssetServer)
			if len(AssetServer) > 0 {
				if err = tx.Model(&AssetServer).Where("host_id=?", JsonData.HostId).Updates(update).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Model(&AssetExtend).Where("host_id=?", JsonData.HostId).Updates(
					updates).Error; err != nil {
					sqlErr = err
				}
			}
			return sqlErr
		})
	}
}
