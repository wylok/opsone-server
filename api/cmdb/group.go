package cmdb

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/cmdb_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strconv"
)

// @Tags 资源组
// @Summary 资源组查询
// @Produce  json
// @Security ApiKeyAuth
// @Param group_id query string false "资产组ID"
// @Param group_name query string false "资产组名称"
// @Param not_page query boolean false "是否分页"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/group [get]
func QueryAssetGroup(c *gin.Context) {
	//资产组查询
	var (
		JsonData      cmdb_conf.AssetGroup
		AssetGroups   []databases.AssetGroups
		CmdbPartition []databases.CmdbPartition
		Response      = common.Response{C: c}
		data          []interface{}
		GroupIds      []string
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
		tx := db.Where("asset_groups.status=?", "active")
		if JsonData.GroupId != "" {
			tx = tx.Where("asset_groups.group_id=?", JsonData.GroupId)
		}
		if JsonData.GroupName != "" {
			tx = tx.Where("asset_groups.group_name=?", JsonData.GroupName)
		}
		if JsonData.NoTPage {
			tx.Find(&AssetGroups)
			Response.Data = AssetGroups
		} else {
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, _ = p.Paging(&AssetGroups)
			if len(AssetGroups) > 0 {
				for _, v := range AssetGroups {
					GroupIds = append(GroupIds, v.GroupId)
				}
				type result struct {
					GroupId string
					Count   int64
				}
				var GroupResult []result
				db.Where("object_id in ? and object_type=?", GroupIds, "asset_group").Find(&CmdbPartition)
				db.Model(&databases.GroupServer{}).Select("group_id, count(host_id) as count").Where(
					"group_id in ?", GroupIds).Group("group_id").Find(&GroupResult)
				for _, v := range AssetGroups {
					d := map[string]string{}
					d["group_id"] = v.GroupId
					d["group_name"] = v.GroupName
					for _, c := range CmdbPartition {
						if c.ObjectType == "asset_group" && c.ObjectId == v.GroupId {
							d["department_id"] = c.DepartmentId
						}
					}
					for _, g := range GroupResult {
						if g.GroupId == v.GroupId {
							d["hosts"] = strconv.FormatInt(g.Count, 10)
						}
					}
					data = append(data, d)
				}
				Response.Data = data
			}
		}
	}
}

// @Tags 资源组
// @Summary 关联主机查询
// @Produce  json
// @Security ApiKeyAuth
// @Param group_ids query array false "资产组ID"
// @Param host_ids query array false "主机ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/group/related/servers [get]
func RelatedGroupServer(c *gin.Context) {
	//关联主机查询
	var (
		JsonData    = cmdb_conf.RelatedGroupServer{}
		GroupServer []databases.GroupServer
		AssetServer []databases.AssetServer
		Response    = common.Response{C: c}
		Data        = map[string][]map[string]string{}
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
		if JsonData.GroupIds != nil {
			JsonData.GroupIds = kits.FormListFormat(JsonData.GroupIds)
			for _, groupId := range JsonData.GroupIds {
				Data[groupId] = []map[string]string{}
				var HostIds []string
				db.Select("host_id").Where("group_id=?", groupId).Find(&GroupServer)
				if len(GroupServer) > 0 {
					for _, v := range GroupServer {
						HostIds = append(HostIds, v.HostId)
					}
				}
				if HostIds != nil {
					db.Select("host_id", "host_name").Where("host_id in ?", HostIds).Find(&AssetServer)
					for _, v := range AssetServer {
						Data[groupId] = append(Data[groupId], map[string]string{"host_id": v.HostId, "host_name": v.Hostname})
					}
				}
			}
		}
		if JsonData.HostIds != nil {
			JsonData.HostIds = kits.FormListFormat(JsonData.HostIds)
			for _, hostId := range JsonData.HostIds {
				db.Select("group_id").Where("host_id=?", hostId).Find(&GroupServer)
				if len(GroupServer) > 0 {
					Data[hostId] = []map[string]string{}
					for _, v := range GroupServer {
						Data[hostId] = append(Data[hostId], map[string]string{"group_id": v.GroupId})
					}
				}
			}
		}
		Response.Data = Data
	}
}

// @Tags 资源组
// @Summary 资源组主机查询
// @Produce  json
// @Security ApiKeyAuth
// @Param group_ids query array true "资产组ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/group/servers [get]
func GroupServers(c *gin.Context) {
	//资源组主机查询
	var (
		JsonData    = cmdb_conf.GroupServers{}
		GroupServer []databases.GroupServer
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
		JsonData.GroupIds = kits.FormListFormat(JsonData.GroupIds)
		sql := "join cmdb_partition on cmdb_partition.object_id = group_server.host_id " +
			"and cmdb_partition.object_type=?"
		db.Joins(sql, "server").Where("group_id in ?", JsonData.GroupIds).Find(&GroupServer)
		if len(GroupServer) > 0 {
			Response.Data = GroupServer
		}
	}
}

// @Tags 资源组
// @Summary 资源组主机详情查询
// @Produce  json
// @Security ApiKeyAuth
// @Param group_id query string true "资产组ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/group/servers/detail [get]
func GroupServersDetail(c *gin.Context) {
	//资源组主机详情查询
	var (
		JsonData    = cmdb_conf.GroupNoneServers{}
		AssetServer []databases.AssetServer
		GroupServer []databases.GroupServer
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
		var (
			ghosts  = map[string]struct{}{}
			hosts   []string
			hostIds []string
		)
		db.Select("host_id").Where("group_id = ?", JsonData.GroupId).Find(&GroupServer)
		if len(GroupServer) > 0 {
			for _, v := range GroupServer {
				ghosts[v.HostId] = struct{}{}
			}
		}
		db.Select("host_id").Find(&GroupServer)
		if len(GroupServer) > 0 {
			for _, v := range GroupServer {
				hosts = append(hosts, v.HostId)
			}
		}
		if len(hosts) == 0 {
			//初次配置资产组
			db.Select("host_id", "host_name").Where(
				"asset_status=?", "assigned").Find(&AssetServer)
		} else {
			if len(ghosts) == 0 {
				db.Select("host_id", "host_name").Where("host_id not in ? and asset_status=?",
					hosts, "assigned").Find(&AssetServer)
			} else {
				for _, v := range hosts {
					_, ok := ghosts[v]
					if !ok {
						hostIds = append(hostIds, v)
					}
				}
				if len(hostIds) == 0 {
					db.Select("host_id", "host_name").Where("asset_status=?", "assigned").Find(&AssetServer)
				} else {
					db.Select("host_id", "host_name").Where("host_id not in ? and asset_status=?",
						hostIds, "assigned").Find(&AssetServer)
				}
			}
		}
		if len(AssetServer) > 0 {
			Response.Data = AssetServer
		}
	}
}

// @Tags 资源组
// @Summary 创建资源组
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.CreateAssetGroup true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/group [post]
func CreateAssetGroup(c *gin.Context) {
	//创建资源组
	var (
		sqlErr      error
		JsonData    = cmdb_conf.CreateAssetGroup{}
		GroupServer []databases.GroupServer
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
		GroupId := kits.RandString(0)
		err = db.Transaction(func(tx *gorm.DB) error {
			ag := databases.AssetGroups{GroupId: GroupId, GroupName: JsonData.GroupName, Status: "active"}
			if err = tx.Create(&ag).Error; err != nil {
				sqlErr = err
			}
			cp := databases.CmdbPartition{ObjectId: GroupId, ObjectType: "asset_group", DepartmentId: JsonData.DepartmentId}
			if err = tx.Create(&cp).Error; err != nil {
				sqlErr = err
			}
			if JsonData.AssetType != "" && JsonData.AssetIds != nil {
				if JsonData.AssetType == "server" {
					for _, hostId := range JsonData.AssetIds {
						db.Where("host_id=?", hostId).First(&GroupServer)
						if len(GroupServer) == 0 {
							gs := databases.GroupServer{GroupId: GroupId, HostId: hostId}
							if err = tx.Create(&gs).Error; err != nil {
								sqlErr = err
							}
						}
					}
				}
			}
			return sqlErr
		})
	}
}

// @Tags 资源组
// @Summary 变更资源组
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.ChangeAssetGroup true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/group [put]
func ChangeAssetGroup(c *gin.Context) {
	//变更资源组
	var (
		JsonData      = cmdb_conf.ChangeAssetGroup{}
		AssetGroups   []databases.AssetGroups
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
		db.Where("group_id=?", JsonData.GroupId).Find(&AssetGroups)
		if len(AssetGroups) > 0 {
			if JsonData.GroupName != "" {
				err = db.Model(&AssetGroups).Where("group_id=?", JsonData.GroupId).Updates(
					databases.AssetGroups{GroupName: JsonData.GroupName}).Error
			}
			if len(JsonData.AssetIds) == 0 {
				err = db.Where("group_id=?", JsonData.GroupId).Delete(&GroupServer).Error
			}
			if len(JsonData.AssetIds) > 0 {
				for _, assetId := range JsonData.AssetIds {
					db.Where("host_id=? and group_id=?", assetId, JsonData.GroupId).Find(&GroupServer)
					if len(GroupServer) == 0 {
						gs := databases.GroupServer{GroupId: JsonData.GroupId, HostId: assetId}
						err = db.Create(&gs).Error
					}
				}
				db.Select("host_id").Where("group_id=?", JsonData.GroupId).Find(&GroupServer)
				if len(GroupServer) > 0 {
					for _, v := range GroupServer {
						delHost := true
						for _, assetId := range JsonData.AssetIds {
							if v.HostId == assetId {
								delHost = false
							}
						}
						if delHost {
							err = db.Where("host_id=? and group_id=?", v.HostId,
								JsonData.GroupId).Delete(&GroupServer).Error
						}
					}
				}
			}
			if JsonData.DepartmentId != "" {
				db.Where("object_type=? and object_id=?", "asset_group", JsonData.GroupId).First(&CmdbPartition)
				if len(CmdbPartition) > 0 {
					err = db.Model(&CmdbPartition).Where("object_type=? and object_id=?",
						"asset_group", JsonData.GroupId).Updates(
						databases.CmdbPartition{DepartmentId: JsonData.DepartmentId}).Error
				} else {
					cp := databases.CmdbPartition{ObjectId: JsonData.GroupId, ObjectType: "asset_group",
						DepartmentId: JsonData.DepartmentId}
					err = db.Create(&cp).Error
				}
			}
		} else {
			err = errors.New(JsonData.GroupId + "无效的资源组ID")
		}
	}
}

// @Tags 资源组
// @Summary 删除资源组
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.DeleteAssetGroup true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/group [delete]
func DeleteAssetGroup(c *gin.Context) {
	//删除资源组
	var (
		sqlErr        error
		JsonData      = cmdb_conf.DeleteAssetGroup{}
		GroupServer   []databases.GroupServer
		AssetGroups   []databases.AssetGroups
		CmdbPartition []databases.CmdbPartition
		Count         int64
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
		db.Model(&databases.AssetGroups{}).Where("group_id in ?", JsonData.GroupIds).Count(&Count)
		if len(JsonData.GroupIds) == int(Count) {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("group_id in ?", JsonData.GroupIds).Delete(&GroupServer).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("object_id in ? and object_type=?", JsonData.GroupIds,
					"asset_group").Delete(&CmdbPartition).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("group_id in ?", JsonData.GroupIds).Delete(&AssetGroups).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New("资源组ID列表包含无效数据")
		}
	}
}
