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
// @Summary configMap列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/configMaps [get]
func ConfigMaps(c *gin.Context) {
	//configMap列表
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ListApplication
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
			Response.Data = k8s.ListConfigMaps(JsonData.NameSpace)
		}
	}
}

// @Tags k8s集群
// @Summary ConfigMapYaml
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Param name query string true "configMap名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/configMap/yaml [get]
func ConfigMapYaml(c *gin.Context) {
	//ConfigMapYaml
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.Application
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
			Response.Data = k8s.GetConfigMapYaml(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 删除configMap
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.App true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/configMap [delete]
func DeleteConfigMap(c *gin.Context) {
	//删除configMap
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
			err = k8s.DelConfigMap(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 创建configMap
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.ConfigMap true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/configMap [post]
func CreatConfigMap(c *gin.Context) {
	//创建configMap
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ConfigMap
		Response   = common.Response{C: c}
		data       = make(map[string]string)
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
			for _, d := range JsonData.Data {
				data[d["key"]] = d["value"]
			}
			k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
			err = k8s.AddConfigMap(JsonData.NameSpace, JsonData.Name, data)
		}
	}
}

// @Tags k8s集群
// @Summary 更新configMap
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.ConfigMap true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/configMap [put]
func UpdateConfigMap(c *gin.Context) {
	//更新configMap
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ConfigMap
		Response   = common.Response{C: c}
		data       = make(map[string]string)
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
			for _, d := range JsonData.Data {
				data[d["key"]] = d["value"]
			}
			k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
			err = k8s.UpdateConfigMap(JsonData.NameSpace, JsonData.Name, data)
		}
	}
}
