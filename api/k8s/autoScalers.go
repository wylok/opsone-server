package k8s

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/k8s_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags k8s集群
// @Summary Autoscalers列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Param deployment query string false "deployment名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/autoscalers [get]
func Autoscalers(c *gin.Context) {
	//Autoscalers列表
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ListAutoscalers
		Response   = common.Response{C: c}
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
	if err == nil {
		db.Where("k8s_id=?", JsonData.K8sID).First(&K8sCluster)
		if len(K8sCluster) > 0 {
			k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
			Response.Data = k8s.ListAutoscalers(JsonData.NameSpace, JsonData.Deployment)
		}
	}
}

// @Tags k8s集群
// @Summary 设置Autoscaler
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.UpdateAutoscaler true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/autoscaler [put]
func UpdateAutoscaler(c *gin.Context) {
	//设置Autoscaler
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.UpdateAutoscaler
		Response   = common.Response{C: c}
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("k8s_id=?", JsonData.K8sID).First(&K8sCluster)
		if len(K8sCluster) > 0 {
			k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
			err = k8s.UpdateAutoscaler(JsonData.NameSpace, JsonData.Name, JsonData.Scaler)
		}
	}
}

// @Tags k8s集群
// @Summary 删除Autoscaler
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.App true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/autoscaler [delete]
func DeleteAutoscaler(c *gin.Context) {
	//删除Autoscaler
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.App
		Response   = common.Response{C: c}
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("k8s_id=?", JsonData.K8sID).First(&K8sCluster)
		if len(K8sCluster) > 0 {
			k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
			err = k8s.DeleteAutoscaler(JsonData.NameSpace, JsonData.Name)
		}
	}
}
