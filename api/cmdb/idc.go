package cmdb

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/cmdb_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags IDC池
// @Summary 查询IDC配置
// @Produce  json
// @Security ApiKeyAuth
// @Param idc query string false "idc名称"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/idc [get]
func QueryIdc(c *gin.Context) {
	//查询IDC配置
	var (
		JsonData cmdb_conf.Idc
		AssetIdc []databases.AssetIdc
		Response = common.Response{C: c}
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
		tx := db.Order("id desc")
		// 参数匹配
		if JsonData.Idc != "" {
			tx = tx.Where("idc like ? or idc_cn like ?", "%"+JsonData.Idc+"%", "%"+JsonData.Idc+"%")
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&AssetIdc)
	}
}

// @Tags IDC池
// @Summary 删除IDC配置
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.DeleteSshKey true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/idc [delete]
func DeleteIdc(c *gin.Context) {
	//删除IDC配置
	var (
		JsonData    cmdb_conf.DeleteIdc
		AssetIdc    []databases.AssetIdc
		AssetExtend []databases.AssetExtend
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("idc_id = ?", JsonData.IdcId).Find(&AssetExtend)
		if len(AssetExtend) == 0 {
			db.Where("idc_id = ?", JsonData.IdcId).Delete(&AssetIdc)
		} else {
			err = errors.New(JsonData.IdcId + "IDC已经关联服务器,无法删除操作")
		}
	}
}

// @Tags IDC池
// @Summary 新增IDC配置
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AddIdc true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/idc [post]
func AddIdc(c *gin.Context) {
	//新增IDC配置
	var (
		JsonData cmdb_conf.AddIdc
		AssetIdc []databases.AssetIdc
		Response = common.Response{C: c}
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
		db.Where("idc=? and region=? and data_center=?", JsonData.Idc, JsonData.Region, JsonData.DataCenter).Find(&AssetIdc)
		if len(AssetIdc) == 0 {
			ai := databases.AssetIdc{IdcId: kits.RandString(12), Idc: JsonData.Idc, IdcCn: JsonData.IdcCn,
				Region: JsonData.Region, RegionCn: JsonData.RegionCn, DataCenter: JsonData.DataCenter,
				CreateTime: time.Now(), UpdateTime: time.Now()}
			err = db.Create(&ai).Error
		} else {
			err = errors.New(JsonData.Idc + "已存在!")
		}
	}
}

// @Tags IDC池
// @Summary 修改IDC配置
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.ModifyIdc true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/idc [put]
func ModifyIdc(c *gin.Context) {
	//修改IDC配置
	var (
		JsonData cmdb_conf.ModifyIdc
		AssetIdc []databases.AssetIdc
		Response = common.Response{C: c}
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
		db.Where("idc_id=?", JsonData.IdcId).First(&AssetIdc)
		if len(AssetIdc) == 0 {
			err = errors.New(JsonData.IdcId + "不存在")
		} else {
			if JsonData.Idc == "" {
				JsonData.Idc = AssetIdc[0].Idc
			}
			if JsonData.IdcCn == "" {
				JsonData.IdcCn = AssetIdc[0].IdcCn
			}
			if JsonData.Region == "" {
				JsonData.Region = AssetIdc[0].Region
			}
			if JsonData.RegionCn == "" {
				JsonData.RegionCn = AssetIdc[0].RegionCn
			}
			if JsonData.DataCenter == "" {
				JsonData.DataCenter = AssetIdc[0].DataCenter
			}
			err = db.Model(&AssetIdc).Where("idc_id=?", JsonData.IdcId).Updates(databases.AssetIdc{Idc: JsonData.Idc,
				IdcCn: JsonData.IdcCn, Region: JsonData.Region, RegionCn: JsonData.RegionCn,
				DataCenter: JsonData.DataCenter, UpdateTime: time.Now()}).Error
		}
	}
}
