package k8s

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"inner/conf/k8s_conf"
	"inner/modules/common"
	"inner/modules/databases"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// @Tags k8s集群
// @Summary deployment列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/deployments [get]
func Deployments(c *gin.Context) {
	//deployment列表
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
			Response.Data = k8s.ListDeployments(JsonData.NameSpace)
		}
	}
}

// @Tags k8s集群
// @Summary DeploymentYaml
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Param name query string true "Deployment名称"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/deployment/yaml [get]
func DeploymentYaml(c *gin.Context) {
	//DeploymentYaml
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
			Response.Data = k8s.GetDeploymentYaml(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary replicaSet列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/replicaSets [get]
func ReplicaSets(c *gin.Context) {
	//replicaSet列表
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.ListReplicaSet
		Response   = common.Response{C: c}
		RSets      []map[string]interface{}
		OSets      []map[string]interface{}
		ASets      []string
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
			for _, item := range k8s.ListReplicaSets(JsonData.NameSpace) {
				if strings.HasPrefix(item.ObjectMeta.Name, JsonData.Name+"-") {
					OSets = append(OSets, map[string]interface{}{"Name": item.ObjectMeta.Name,
						"Image":      item.Spec.Template.Spec.Containers[0].Image,
						"Generation": item.Status.ObservedGeneration, "CreationTimestamp": item.CreationTimestamp})
					ASets = append(ASets, item.CreationTimestamp.String())
				}
			}
			slice.Sort(ASets)
			for _, v := range ASets {
				for _, i := range OSets {
					if i["CreationTimestamp"].(v1.Time).String() == v {
						RSets = append(RSets, i)
					}
				}
			}
		}
	}
	Response.Data = RSets
}

// @Tags k8s集群
// @Summary 删除deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.Deployment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/deployment [delete]
func DeleteDeployment(c *gin.Context) {
	//删除deployment
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.Deployment
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
			err = k8s.DelDeployment(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 重启deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.Deployment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/deployment/restart [post]
func RestartDeployment(c *gin.Context) {
	//重启deployment
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.Deployment
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
			err = k8s.RestartDeployment(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 修改deployment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.Deployment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/deployment [put]
func UpdateDeployment(c *gin.Context) {
	//修改deployment
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.Deployment
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
				err = k8s.UpdateDeployment(JsonData.NameSpace, JsonData.Name, JsonData.Images)
			}
		}
	}
}
