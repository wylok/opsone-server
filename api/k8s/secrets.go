package k8s

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/k8s_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
)

// @Tags k8s集群
// @Summary secret列表
// @Produce  json
// @Security ApiKeyAuth
// @Param k8s_id query string true "集群ID"
// @Param namespace query string true "名称空间"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/k8s/secrets [get]
func Secrets(c *gin.Context) {
	//secret列表
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
			Response.Data = k8s.ListSecrets(JsonData.NameSpace)
		}
	}
}

// @Tags k8s集群
// @Summary 删除secret
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.App true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/secret [delete]
func DeleteSecret(c *gin.Context) {
	//删除secret
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
			err = k8s.DelSecret(JsonData.NameSpace, JsonData.Name)
		}
	}
}

// @Tags k8s集群
// @Summary 创建secret
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  k8s_conf.Secrets true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/k8s/secret [post]
func CreatSecret(c *gin.Context) {
	//创建secret
	var (
		K8sCluster []databases.K8sCluster
		JsonData   k8s_conf.Secrets
		Response   = common.Response{C: c}
		data       = map[string]string{}
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
			err = errors.New("暂时只支持docker类型")
			if JsonData.Type == "kubernetes.io/dockerconfigjson" {
				data["docker-server"] = strings.TrimSpace(JsonData.Data["dockerServer"])
				data["docker-username"] = strings.TrimSpace(JsonData.Data["dockerUserName"])
				data["docker-password"] = strings.TrimSpace(JsonData.Data["dockerPassword"])
				d, _ := json.Marshal(data)
				k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
				err = k8s.AddSecret(JsonData.NameSpace, strings.TrimSpace(JsonData.Name),
					JsonData.Type, map[string]string{".dockerconfigjson": string(d)})
			}
		}
	}
}
