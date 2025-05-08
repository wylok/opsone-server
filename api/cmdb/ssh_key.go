package cmdb

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/cmdb_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 资源组
// @Summary ssh密钥查询
// @Produce  json
// @Security ApiKeyAuth
// @Param key_name query string false "ssh密钥名称"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/ssh_key [get]
func QuerySshKey(c *gin.Context) {
	//ssh密钥查询
	var (
		JsonData cmdb_conf.SshKey
		SshKey   []databases.SshKey
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
		if JsonData.KeyName != "" {
			tx = tx.Where("key_name=?", JsonData.KeyName)
		}
		if JsonData.SshUser != "" {
			tx = tx.Where("ssh_user=?", JsonData.SshUser)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&SshKey)
	}
}

// @Tags 资源组
// @Summary 上传ssh密钥
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.UploadSshKey true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/ssh_key [post]
func SshKeyUpload(c *gin.Context) {
	//上传ssh密钥
	var (
		JsonData cmdb_conf.UploadSshKey
		Response = common.Response{C: c}
		SshKey   []databases.SshKey
		Encrypt  = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
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
		db.Where("key_name=?", JsonData.KeyName).Find(&SshKey)
		if len(SshKey) == 0 {
			pk := Encrypt.EncryptString(JsonData.KeyConfig, true)
			db.Where("ssh_key=?", pk).Find(&SshKey)
			if len(SshKey) == 0 {
				sk := databases.SshKey{KeyName: JsonData.KeyName, SshUser: "root",
					SshKey: pk, CreateTime: time.Now()}
				err = db.Create(&sk).Error
			} else {
				err = errors.New(JsonData.KeyName + "ssh密钥已存在")
			}
		} else {
			err = errors.New(JsonData.KeyName + "ssh密钥名称已存在")
		}
	}
}

// @Tags 资源组
// @Summary 删除ssh密钥
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.DeleteSshKey true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/ssh_key [delete]
func SshKeyDelete(c *gin.Context) {
	//删除ssh密钥
	var (
		JsonData        = cmdb_conf.DeleteSshKey{}
		AssetServerPool []databases.AssetServerPool
		SshKey          []databases.SshKey
		Response        = common.Response{C: c}
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
		db.Where("ssh_key_name = ?", JsonData.KeyName).Find(&AssetServerPool)
		if len(AssetServerPool) == 0 {
			db.Where("key_name = ?", JsonData.KeyName).Delete(&SshKey)
		} else {
			err = errors.New(JsonData.KeyName + "密钥已经关联服务器,无法删除操作")
		}
	}
}
