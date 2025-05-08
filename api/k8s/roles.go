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
// @Summary ClusterRoles列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/role/cluster [get]
func ClusterRoles(c *gin.Context) {
	//ClusterRoles列表
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ListRoles
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
			Response.Data = k8s.ListClusterRoles()
		}
	}
}

// @Tags k8s集群
// @Summary ClusterRoleBindings列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/role/cluster/binding [get]
func ClusterRoleBindings(c *gin.Context) {
	//ClusterRoleBindings列表
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ListRoles
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
			Response.Data = k8s.ListClusterRoleBindings()
		}
	}
}
