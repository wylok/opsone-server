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
// @Summary daemonSet列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/daemonSets [get]
func DaemonSets(c *gin.Context) {
	//daemonSet列表
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
			Response.Data = k8s.ListDaemonSets(JsonData.NameSpace)
		}
	}
}

// @Tags k8s集群
// @Summary DaemonSetYaml
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Param name query string true "daemonSet名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/daemonSet/yaml [get]
func DaemonSetYaml(c *gin.Context) {
	//DaemonSetYaml
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
			Response.Data = k8s.GetDaemonSetYaml(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 删除daemonSet
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.App true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/daemonSet [delete]
func DeleteDaemonSet(c *gin.Context) {
	//删除deployment
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
			err = k8s.DelDaemonSet(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 修改daemonSet
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.DaemonSet true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/daemonSet [put]
func UpdateDaemonSet(c *gin.Context) {
	//修改daemonSet
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.DaemonSet
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
			if JsonData.Images != nil {
				k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
				err = k8s.UpdateDaemonSet(JsonData.NameSpace, JsonData.Name, JsonData.Images)
			}
		}
	}
}

// @Tags k8s集群
// @Summary 重启daemonSet
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.DaemonSet true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/daemonSet/restart [post]
func RestartDaemonSet(c *gin.Context) {
	//重启daemonSet
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.DaemonSet
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
			err = k8s.RestartDaemonSet(JsonData.NameSpace, JsonData.Name)
		}
	}
}
