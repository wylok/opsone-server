package cmdb

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/cmdb_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strings"
	"time"
)

// @Tags 资源池
// @Summary 查询服务器资源池
// @Produce  json
// @Security ApiKeyAuth
// @Param host_name query string false "主机名称"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/pool/server [get]
func QueryPoolServer(c *gin.Context) {
	//查询服务器资源池
	var (
		JsonData    cmdb_conf.ServerPool
		AssetServer []databases.AssetServer
		AssetNet    []databases.AssetNet
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Where("asset_status = ?", "available")
		// 参数匹配
		if JsonData.HostName != "" {
			tx = tx.Where("host_name like ?", "%"+JsonData.HostName+"%")
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AssetServer)
		if len(AssetServer) > 0 {
			var Data []map[string]interface{}
			for _, sv := range AssetServer {
				var online bool
				if rc.HExists(ctx, platform_conf.AgentAliveKey, sv.HostId).Val() {
					if !rc.HExists(ctx, platform_conf.OfflineAssetKey, sv.HostId).Val() {
						online = true
					}
				}
				db.Where("host_id=?", sv.HostId).Find(&AssetNet)
				if len(AssetNet) > 0 {
					Data = append(Data, map[string]interface{}{"host_id": sv.HostId, "host_name": sv.Hostname,
						"host_type_cn": sv.HostTypeCn, "manufacturer": sv.Manufacturer, "cpu": sv.Cpu,
						"memory": sv.Memory, "disk": sv.Disk, "product_name": sv.ProductName, "sn": sv.Sn,
						"ip": AssetNet[0].Ip, "asset_status": sv.AssetStatus, "platform": sv.Platform,
						"platform_version": sv.PlatformVersion, "online": online})
				}
			}
			Response.Data = Data
		}
	}
}

// @Tags 资源池
// @Summary 服务器资源分配
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AssignAssetPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server [post]
func AssignServerPool(c *gin.Context) {
	//服务器资源分配
	var (
		sqlErr        error
		JsonData      cmdb_conf.AssignAssetPool
		AssetServer   []databases.AssetServer
		CmdbPartition []databases.CmdbPartition
		AssetUnder    []databases.AssetUnder
		Response      = common.Response{C: c}
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
	if err == nil && len(JsonData.AssetIds) > 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			if JsonData.AssetTag == "" {
				JsonData.AssetTag = "None"
			}
			if err = tx.Model(&AssetServer).Where("host_id in ?", JsonData.AssetIds).Updates(
				databases.AssetServer{AssetTag: JsonData.AssetTag, AssetStatus: "assigned"}).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Model(&CmdbPartition).Where("object_type=? and object_id in ?", "server",
				JsonData.AssetIds).Updates(
				databases.CmdbPartition{DepartmentId: JsonData.DepartmentId}).Error; err != nil {
				sqlErr = err
			}
			for _, hostId := range JsonData.AssetIds {
				db.Where("asset_id=? and asset_type=?", hostId, "server").First(&AssetUnder)
				if len(AssetUnder) > 0 {
					if err = tx.Model(&AssetUnder).Where("asset_id=? and asset_type=?", hostId, "server").Updates(
						databases.AssetUnder{DepartmentId: JsonData.DepartmentId,
							BusinessId: JsonData.BusinessId}).Error; err != nil {
						sqlErr = err
					}
				} else {
					au := databases.AssetUnder{AssetId: hostId, AssetType: "server",
						DepartmentId: JsonData.DepartmentId, BusinessId: JsonData.BusinessId}
					if err = tx.Create(&au).Error; err != nil {
						sqlErr = err
					}
				}
			}
			return sqlErr
		})
	}
}

// @Tags 资源池
// @Summary 回收服务器资源
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.Asset true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server [put]
func ReclaimServerPool(c *gin.Context) {
	//回收服务器资源
	var (
		sqlErr        error
		JsonData      cmdb_conf.Asset
		AssetServer   []databases.AssetServer
		AssetUnder    []databases.AssetUnder
		GroupServer   []databases.GroupServer
		CmdbPartition []databases.CmdbPartition
		Response      = common.Response{C: c}
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
	if err == nil && len(JsonData.AssetIds) > 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			if err = tx.Model(&AssetServer).Where("host_id in ?", JsonData.AssetIds).Updates(
				databases.AssetServer{AssetTag: "None", AssetStatus: "available"}).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Where("asset_id in ? and asset_type=?", JsonData.AssetIds,
				"server").Delete(&AssetUnder).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Model(&CmdbPartition).Where("object_id in ? and object_type=?", JsonData.AssetIds,
				"server").Updates(databases.CmdbPartition{DepartmentId: "None"}).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Where("host_id in ?", JsonData.AssetIds).Delete(&GroupServer).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
}

// @Tags 资源池
// @Summary 下架服务器资源
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.Asset true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server [delete]
func DiscardServerPool(c *gin.Context) {
	//下架服务器资源
	var (
		sqlErr        error
		JsonData      = cmdb_conf.Asset{}
		AssetServer   []databases.AssetServer
		AssetNet      []databases.AssetNet
		AssetDisk     []databases.AssetDisk
		AssetExtend   []databases.AssetExtend
		AssetUnder    []databases.AssetUnder
		GroupServer   []databases.GroupServer
		CmdbPartition []databases.CmdbPartition
		Response      = common.Response{C: c}
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
		db.Where("host_id in ?", JsonData.AssetIds).Find(&AssetServer)
		if len(AssetServer) > 0 {
			var hostIds []string
			err = db.Transaction(func(tx *gorm.DB) error {
				for _, v := range AssetServer {
					hostIds = append(hostIds, v.HostId)
					rc.HSet(ctx, platform_conf.DiscardAssetKey, v.HostId, "")
				}
				rc.Expire(ctx, platform_conf.DiscardAssetKey, 30*time.Minute)
				if err = tx.Where("host_id in ?", hostIds).Delete(&AssetServer).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", hostIds).Delete(&AssetNet).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", hostIds).Delete(&AssetDisk).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", hostIds).Delete(&AssetExtend).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("asset_id in ?", hostIds).Delete(&AssetUnder).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", hostIds).Delete(&GroupServer).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("object_id in ? and object_type=?", hostIds, "server").Delete(&CmdbPartition).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
			if err == nil && hostIds != nil {
				for _, v := range hostIds {
					rc.HDel(ctx, platform_conf.HostCpuCoreKey, v)
				}
			}
		}
	}
}

// @Tags 资源池
// @Summary 服务器业务组变更
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AssetBusiness true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server/business [put]
func ModifyAssetBusiness(c *gin.Context) {
	//服务器业务组变更
	var (
		sqlErr     error
		JsonData   = cmdb_conf.AssetBusiness{}
		AssetUnder []databases.AssetUnder
		Response   = common.Response{C: c}
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
	if err == nil && len(JsonData.AssetIds) > 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, hostId := range JsonData.AssetIds {
				db.Where("asset_id=? and asset_type=?", hostId, "server").First(&AssetUnder)
				if len(AssetUnder) > 0 {
					if err = tx.Model(&AssetUnder).Where("asset_id=? and asset_type=?", hostId, "server").Updates(
						databases.AssetUnder{DepartmentId: JsonData.DepartmentId,
							BusinessId: JsonData.BusinessId}).Error; err != nil {
						sqlErr = err
					}
				}
			}
			return sqlErr
		})
	}
}

// @Tags 资源池
// @Summary 查询交换机资源池
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/pool/switch [get]
func QuerySwitchPool(c *gin.Context) {
	//查询交换机资源池
	var (
		JsonData        = cmdb_conf.SwitchPool{}
		AssetSwitchPool []databases.AssetSwitchPool
		Response        = common.Response{C: c}
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		p := databases.Pagination{DB: db, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&AssetSwitchPool)
	}
}

// @Tags 资源池
// @Summary 新增交换机信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AddSwitchPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/switch [post]
func AddSwitchPool(c *gin.Context) {
	//新增交换机信息
	var (
		JsonData        = cmdb_conf.AddSwitchPool{}
		AssetSwitchPool []databases.AssetSwitchPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response        = common.Response{C: c}
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
		pwd := Encrypt.EncryptString(JsonData.SwitchPassword, true)
		if JsonData.StartIp != "" && JsonData.EndIp != "" {
			SIps := strings.Split(JsonData.StartIp, ".")
			EIps := strings.Split(JsonData.EndIp, ".")
			SIp := strings.Join(SIps[:len(SIps)-1], ".")
			EIp := strings.Join(EIps[:len(EIps)-1], ".")
			if SIp == EIp {
				if cast.ToInt(SIps[len(SIps)-1]) == 0 || cast.ToInt(EIps[len(EIps)-1]) == 0 {
					err = errors.New("无效的起始IP或结束IP")
				} else {
					if cast.ToInt(SIps[len(SIps)-1]) <= cast.ToInt(EIps[len(EIps)-1]) {
						if cast.ToInt(EIps[len(EIps)-1])-cast.ToInt(SIps[len(SIps)-1]) <= 20 {
							db.Where("start_ip=? and end_ip=?", JsonData.StartIp, JsonData.EndIp).First(&AssetSwitchPool)
							if len(AssetSwitchPool) == 0 {
								asd := databases.AssetSwitchPool{StartIp: JsonData.StartIp, EndIp: JsonData.EndIp,
									SwitchPort: JsonData.SwitchPort, SwitchUser: JsonData.SwitchUser, SwitchPassword: pwd,
									Discover: 0, IdcId: JsonData.IdcId, SwitchStatus: "enable", CreateTime: time.Now(),
									ModifyTime: time.Now(), SyncTime: time.Now()}
								err = db.Create(&asd).Error
							} else {
								err = errors.New(JsonData.StartIp + "-" + JsonData.EndIp + "已存在")
							}
						} else {
							err = errors.New("IP段区间不能大于20")
						}
					} else {
						err = errors.New("起始IP不能大于结束IP")
					}
				}
			} else {
				err = errors.New("起始IP与结束IP应为网段")
			}
		}
	}
}

// @Tags 资源池
// @Summary 变更交换机信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.ModifySwitchPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/switch [put]
func ModifySwitchPool(c *gin.Context) {
	//变更交换机信息
	var (
		JsonData        = cmdb_conf.ModifySwitchPool{}
		AssetSwitchPool []databases.AssetSwitchPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response        = common.Response{C: c}
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
		db.Where("id=?", JsonData.ID).First(&AssetSwitchPool)
		if len(AssetSwitchPool) > 0 {
			v := AssetSwitchPool[0]
			if JsonData.StartIp != "" {
				v.StartIp = JsonData.StartIp
			}
			if JsonData.EndIp != "" {
				v.EndIp = JsonData.EndIp
			}
			SIps := strings.Split(v.StartIp, ".")
			EIps := strings.Split(v.EndIp, ".")
			SIp := strings.Join(SIps[:len(SIps)-1], ".")
			EIp := strings.Join(EIps[:len(EIps)-1], ".")
			if SIp == EIp {
				if cast.ToInt(SIps[len(SIps)-1]) == 0 || cast.ToInt(EIps[len(EIps)-1]) == 0 {
					err = errors.New("无效的起始IP或结束IP")
				}
				if cast.ToInt(SIps[len(SIps)-1]) > cast.ToInt(EIps[len(EIps)-1]) {
					err = errors.New("起始IP不能大于结束IP")
				}
			} else {
				err = errors.New("起始IP与结束IP应为网段")
			}
			if err == nil {
				if JsonData.SwitchUser != "" {
					v.SwitchUser = JsonData.SwitchUser
				}
				if JsonData.SwitchPassword != "" && v.SwitchPassword != JsonData.SwitchPassword {
					v.SwitchPassword = Encrypt.EncryptString(JsonData.SwitchPassword, true)
				}
				if JsonData.SwitchPort != 0 {
					v.SwitchPort = JsonData.SwitchPort
				}
				if JsonData.IdcId != "" {
					v.IdcId = JsonData.IdcId
				}
				if JsonData.SwitchStatus != "" {
					if JsonData.SwitchStatus == "enable" || JsonData.SwitchStatus == "disable" {
						v.SwitchStatus = JsonData.SwitchStatus
					} else {
						err = errors.New("status无效的参数值")
					}
				}
				if err == nil {
					db.Model(&AssetSwitchPool).Where("id=?", JsonData.ID).Updates(
						databases.AssetSwitchPool{StartIp: v.StartIp, EndIp: v.EndIp, SwitchPort: v.SwitchPort,
							SwitchUser: v.SwitchUser, SwitchPassword: v.SwitchPassword, SwitchStatus: v.SwitchStatus,
							IdcId: v.IdcId, ModifyTime: time.Now()})
				}
			}
		}
	}
}

// @Tags 资源池
// @Summary 删除交换机信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.DeleteSwitchPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server [delete]
func DeleteSwitchPool(c *gin.Context) {
	//删除交换机信息
	var (
		JsonData        = cmdb_conf.DeleteSwitchPool{}
		AssetSwitchPool []databases.AssetSwitchPool
		Response        = common.Response{C: c}
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
		db.Where("id = ?", JsonData.ID).Delete(&AssetSwitchPool)
	}
}

// @Tags 资源池
// @Summary 新增服务器IP池信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AddServerIpPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server/ip [post]
func AddServerIpPool(c *gin.Context) {
	//新增服务器IP池信息
	var (
		JsonData        = cmdb_conf.AddServerIpPool{}
		AssetServerPool []databases.AssetServerPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response        = common.Response{C: c}
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
		pwd := "none"
		SshKeyName := "none"
		if JsonData.SshPassword != "" {
			pwd = Encrypt.EncryptString(JsonData.SshPassword, true)
		}
		if JsonData.SshKeyName != "" {
			SshKeyName = JsonData.SshKeyName
		}
		if JsonData.StartIp != "" && JsonData.EndIp != "" {
			SIps := strings.Split(JsonData.StartIp, ".")
			EIps := strings.Split(JsonData.EndIp, ".")
			SIp := strings.Join(SIps[:len(SIps)-1], ".")
			EIp := strings.Join(EIps[:len(EIps)-1], ".")
			if SIp == EIp {
				if cast.ToInt(SIps[len(SIps)-1]) == 0 || cast.ToInt(EIps[len(EIps)-1]) == 0 {
					err = errors.New("无效的起始IP或结束IP")
				} else {
					if cast.ToInt(SIps[len(SIps)-1]) <= cast.ToInt(EIps[len(EIps)-1]) {
						if cast.ToInt(EIps[len(EIps)-1])-cast.ToInt(SIps[len(SIps)-1]) <= 100 {
							db.Where("start_ip=? and end_ip=?", JsonData.StartIp, JsonData.EndIp).First(&AssetServerPool)
							if len(AssetServerPool) == 0 {
								asd := databases.AssetServerPool{StartIp: JsonData.StartIp, EndIp: JsonData.EndIp,
									SshPort: JsonData.SshPort, SshUser: JsonData.SshUser, SshPassword: pwd, IdcId: JsonData.IdcId,
									SshKeyName: SshKeyName, Discover: 1, Status: "enable", CreateTime: time.Now(),
									ModifyTime: time.Now(), SyncTime: time.Now()}
								err = db.Create(&asd).Error
							} else {
								err = errors.New(JsonData.StartIp + "-" + JsonData.EndIp + "已存在")
							}
						} else {
							err = errors.New("IP段区间不能大于100")
						}
					} else {
						err = errors.New("起始IP不能大于结束IP")
					}
				}
			} else {
				err = errors.New("起始IP与结束IP应为网段")
			}
		} else {
			err = errors.New("起始IP和结束IP不能为空")
		}
	}
}

// @Tags 资源池
// @Summary 查询服务器IP池
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/pool/server/ip [get]
func QueryServerIpPool(c *gin.Context) {
	//查询服务器IP池
	var (
		JsonData        = cmdb_conf.ServerIpPool{}
		AssetServerPool []databases.AssetServerPool
		Response        = common.Response{C: c}
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		p := databases.Pagination{DB: db, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&AssetServerPool)
	}
}

// @Tags 资源池
// @Summary 删除服务器IP池
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.DeleteServerIpPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server/ip [delete]
func DeleteServerIpPool(c *gin.Context) {
	//删除服务器IP池
	var (
		JsonData        = cmdb_conf.DeleteServerIpPool{}
		AssetServerPool []databases.AssetServerPool
		Response        = common.Response{C: c}
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
		db.Where("id = ?", JsonData.ID).Delete(&AssetServerPool)
	}
}

// @Tags 资源池
// @Summary 变更服务器IP池
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.ModifyServerIpPool true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/pool/server/ip [put]
func ModifyServerIpPool(c *gin.Context) {
	//变更服务器IP池
	var (
		JsonData        = cmdb_conf.ModifyServerIpPool{}
		AssetServerPool []databases.AssetServerPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response        = common.Response{C: c}
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
		db.Where("id=?", JsonData.ID).First(&AssetServerPool)
		if len(AssetServerPool) > 0 {
			v := AssetServerPool[0]
			SIps := strings.Split(v.StartIp, ".")
			EIps := strings.Split(v.EndIp, ".")
			SIp := strings.Join(SIps[:len(SIps)-1], ".")
			EIp := strings.Join(EIps[:len(EIps)-1], ".")
			if SIp == EIp {
				if cast.ToInt(SIps[len(SIps)-1]) == 0 || cast.ToInt(EIps[len(EIps)-1]) == 0 {
					err = errors.New("无效的起始IP或结束IP")
				}
				if cast.ToInt(SIps[len(SIps)-1]) > cast.ToInt(EIps[len(EIps)-1]) {
					err = errors.New("起始IP不能大于结束IP")
				}
			} else {
				err = errors.New("起始IP与结束IP应为网段")
			}
			if err == nil {
				if JsonData.StartIp != "" {
					v.StartIp = JsonData.StartIp
				}
				if JsonData.EndIp != "" {
					v.EndIp = JsonData.EndIp
				}
				if JsonData.SshUser != "" {
					v.SshUser = JsonData.SshUser
				}
				if JsonData.IdcId != "" {
					v.IdcId = JsonData.IdcId
				}
				if JsonData.SshPassword != "" && JsonData.SshPassword != "none" && v.SshPassword != JsonData.SshPassword {
					v.SshPassword = Encrypt.EncryptString(JsonData.SshPassword, true)
				}
				if JsonData.SshKeyName != "" && JsonData.SshKeyName != "none" {
					v.SshKeyName = JsonData.SshKeyName
				}
				if JsonData.SshPort != 0 {
					v.SshPort = JsonData.SshPort
				}
				if JsonData.Status != "" {
					if JsonData.Status == "enable" || JsonData.Status == "disable" {
						v.Status = JsonData.Status
					} else {
						err = errors.New("status无效的参数值")
					}
				}
				if err == nil {
					db.Model(&AssetServerPool).Where("id=?", JsonData.ID).Updates(
						databases.AssetServerPool{StartIp: v.StartIp, EndIp: v.EndIp, SshPort: v.SshPort,
							SshUser: v.SshUser, SshPassword: v.SshPassword, SshKeyName: v.SshKeyName,
							IdcId: v.IdcId, Status: v.Status, ModifyTime: time.Now()})
				}
			}
		}
	}
}
