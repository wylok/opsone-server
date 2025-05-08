package cloud

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/cloud_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
	"time"
)

// @Tags 多云管理
// @Summary 密钥列表
// @Produce  json
// @Security ApiKeyAuth
// @Param cloud query string false "公有云"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cloud/key [get]
func QueryCloudKey(c *gin.Context) {
	//密钥列表
	var (
		CloudKeys []databases.CloudKeys
		JsonData  cloud_conf.QKey
		Response  = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
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
		tx := db.Select("cloud", "key_id", "key_type", "end_point", "create_time").Order("create_time desc")
		if JsonData.Cloud != "" {
			tx = tx.Where("cloud=?", JsonData.Cloud)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&CloudKeys)
	}
}

// @Tags 多云管理
// @Summary 新建密钥
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cloud_conf.CloudKey true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cloud/key [post]
func AddCloudKey(c *gin.Context) {
	//密钥列表
	var (
		CloudKeys []databases.CloudKeys
		JsonData  cloud_conf.CloudKey
		Response  = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("cloud=? and key_id=? and key_type=? and end_point=?", JsonData.Cloud, JsonData.KeyId,
			strings.Split(JsonData.KeyType, "-")[0], JsonData.EndPoint).Find(&CloudKeys)
		if len(CloudKeys) == 0 {
			ck := databases.CloudKeys{Cloud: JsonData.Cloud, KeyId: JsonData.KeyId,
				KeySecret: JsonData.KeySecret, KeyType: strings.Split(JsonData.KeyType, "-")[0],
				EndPoint: JsonData.EndPoint, CreateTime: time.Now()}
			err = db.Create(&ck).Error
		} else {
			err = errors.New(JsonData.KeyId + "已经存在")
		}
	}
}

// @Tags 多云管理
// @Summary 删除密钥
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cloud_conf.DelCloudKey true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cloud/key [delete]
func DelCloudKey(c *gin.Context) {
	//密钥列表
	var (
		CloudKeys []databases.CloudKeys
		JsonData  cloud_conf.DelCloudKey
		Response  = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("key_id=?", JsonData.KeyId).Delete(&CloudKeys)
	}
}
