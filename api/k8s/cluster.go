package k8s

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/k8s_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags k8s集群
// @Summary k8s集群查询
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_name query string false "k8s集群名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/cluster [get]
func QueryK8sCluster(c *gin.Context) {
	//k8s集群查询
	var (
		JsonData   k8s_conf.K8sCluster
		K8sCluster []databases.K8sCluster
		Response   = common.Response{C: c}
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
		tx := db.Select("k8s_id", "k8s_name", "alarm_channel", "alarm_contacts", "create_time")
		if JsonData.K8sName != "" {
			tx = tx.Where("k8s_name = ?", JsonData.K8sName)
		}
		tx.Find(&K8sCluster)
		Response.Data = K8sCluster
	}
}

// @Tags k8s集群
// @Summary 上传config文件
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.UploadK8sCluster true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/cluster [post]
func UploadK8sCluster(c *gin.Context) {
	//上传config文件
	var (
		JsonData   k8s_conf.UploadK8sCluster
		Response   = common.Response{C: c}
		K8sCluster []databases.K8sCluster
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
		db.Where("k8s_name=?", JsonData.K8sName).Find(&K8sCluster)
		if len(K8sCluster) == 0 {
			sk := databases.K8sCluster{K8sId: kits.RandString(8), K8sName: JsonData.K8sName,
				K8sConfig: JsonData.K8sConfig, AlarmChannel: JsonData.AlarmChannel,
				AlarmContacts: JsonData.AlarmContacts, CreateTime: time.Now()}
			err = db.Create(&sk).Error
		} else {
			err = errors.New(JsonData.K8sName + "k8s集群名称已存在")
		}
	}
}

// @Tags k8s集群
// @Summary 修改k8s集群配置
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.ModifyK8sCluster true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/cluster [put]
func ModifyK8sCluster(c *gin.Context) {
	//修改k8s集群配置
	var (
		sqlErr     error
		JsonData   k8s_conf.ModifyK8sCluster
		Response   = common.Response{C: c}
		K8sCluster []databases.K8sCluster
		K8sAlarm   []databases.K8sAlarm
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
		db.Where("k8s_id=?", JsonData.K8sId).Find(&K8sCluster)
		if len(K8sCluster) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Model(&K8sCluster).Where("k8s_id=?", JsonData.K8sId).Updates(
					databases.K8sCluster{K8sName: JsonData.K8sName, AlarmChannel: JsonData.AlarmChannel,
						AlarmContacts: JsonData.AlarmContacts}).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Model(&K8sAlarm).Where("k8s_id=?", JsonData.K8sId).Updates(
					databases.K8sAlarm{K8sName: JsonData.K8sName}).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New(JsonData.K8sId + "k8s集群不存在")
		}
	}
}

// @Tags k8s集群
// @Summary 删除k8s集群
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.DelK8sCluster true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/cluster [delete]
func DelK8sCluster(c *gin.Context) {
	//删除k8s集群
	var (
		JsonData   k8s_conf.DelK8sCluster
		K8sCluster []databases.K8sCluster
		K8sAlarm   []databases.K8sAlarm
		Response   = common.Response{C: c}
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
		db.Where("k8s_id = ? and k8s_name = ?", JsonData.K8sID, JsonData.K8sName).Delete(&K8sCluster)
		db.Where("k8s_id = ? and k8s_name = ?", JsonData.K8sID, JsonData.K8sName).Delete(&K8sAlarm)
	}
}
