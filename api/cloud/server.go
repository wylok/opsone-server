package cloud

import (
	"errors"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	common2 "github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/gin-gonic/gin"
	volcEcs "github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
	"inner/conf/cloud_conf"
	"inner/conf/cmdb_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
)

// @Tags 资产主机
// @Summary 云主机查询
// @Produce  json
// @Security ApiKeyAuth
// @Param cloud query string true "公有云"
// @Param instance_id query string false "实例ID"
// @Param instance_name query string false "实例名称"
// @Param host_name query string false "主机名称"
// @Param sn query string false "主机SN"
// @Param status query string false "主机状态"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cloud/servers [get]
func QueryCloudServer(c *gin.Context) {
	//主机信息查询
	var (
		JsonData    = cmdb_conf.CloudServer{}
		CloudServer []databases.CloudServer
		Response    = common.Response{C: c}
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
		tx := db.Order("CreationDate desc").Where("cloud=?", JsonData.Cloud)
		// 部分参数匹配
		if JsonData.InstanceId != "" {
			tx = tx.Where("instance_id = ?", JsonData.InstanceId)
		}
		if JsonData.InstanceName != "" {
			tx = tx.Where("instance_name = ?", JsonData.InstanceName)
		}
		if JsonData.HostName != "" {
			tx = tx.Where("host_name = ?", JsonData.HostName)
		}
		if JsonData.SN != "" {
			tx = tx.Where("sn = ?", JsonData.SN)
		}
		if JsonData.Status != "" {
			tx = tx.Where("status = ?", JsonData.Status)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&CloudServer)
	}
}

// @Tags 多云管理
// @Summary 操作云主机
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cloud_conf.OperateCloudServer true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cloud/server/operate [post]
func OperateCloudServer(c *gin.Context) {
	//操作云主机
	var (
		cloud       string
		client      interface{}
		CloudKeys   []databases.CloudKeys
		CloudServer []databases.CloudServer
		JsonData    cloud_conf.OperateCloudServer
		Response    = common.Response{C: c}
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
		db.Where("key_id=?", JsonData.KeyId).Find(&CloudKeys)
		if len(CloudKeys) > 0 {
			for _, v := range CloudKeys {
				err = errors.New("client is nil)")
				switch v.Cloud {
				case "aliyun":
					cloud = "aliyun"
					if v.KeyType == "ecs" {
						client = ecs.NewClient(JsonData.KeyId, v.KeySecret)
						regionId := strings.Split(v.EndPoint, ".")
						if len(regionId) >= 2 {
							client.(*ecs.Client).WithRegionID(common2.Region(regionId[1]))
						}
						client.(*ecs.Client).WithEndpoint("https://" + v.EndPoint)
					}
				case "volcengine":
					cloud = "volcengine"
					if v.KeyType == "ecs" {
						regionId := "cn-beijing"
						regionIds := strings.Split(v.EndPoint, ".")
						if len(regionIds) >= 2 {
							regionId = regionIds[1]
						}
						config := volcengine.NewConfig().
							WithRegion(regionId).
							WithCredentials(credentials.NewStaticCredentials(v.KeyId, v.KeySecret, ""))
						sess, err := session.NewSession(config)
						if err == nil {
							client = volcEcs.New(sess)
						}
					}
				case "baidu":
					cloud = "baidu"
					if v.KeyType == "bcc" {
						client, _ = bcc.NewClient(JsonData.KeyId, v.KeySecret, v.EndPoint)
					}
				}
			}
			if client != nil {
				switch JsonData.Operate {
				case "stop":
					if cloud == "aliyun" {
						err = client.(*ecs.Client).StopInstance(JsonData.InstanceId, true)
					}
					if cloud == "baidu" {
						err = client.(*bcc.Client).StopInstance(JsonData.InstanceId, true)
					}
					if cloud == "volcengine" {
						forceStop := true
						_, err = client.(*volcEcs.ECS).StopInstance(&volcEcs.StopInstanceInput{InstanceId: &JsonData.InstanceId, ForceStop: &forceStop})
					}
					if err == nil {
						db.Model(&CloudServer).Where("instance_id=?", JsonData.InstanceId).Updates(databases.CloudServer{Status: "Stopped"})
					}
				case "start":
					if cloud == "aliyun" {
						err = client.(*ecs.Client).StartInstance(JsonData.InstanceId)
					}
					if cloud == "baidu" {
						err = client.(*bcc.Client).StartInstance(JsonData.InstanceId)
					}
					if cloud == "volcengine" {
						_, err = client.(*volcEcs.ECS).StartInstance(&volcEcs.StartInstanceInput{InstanceId: &JsonData.InstanceId})
					}
					if err == nil {
						db.Model(&CloudServer).Where("instance_id=?", JsonData.InstanceId).Updates(databases.CloudServer{Status: "Running"})
					}
				case "reboot":
					if cloud == "aliyun" {
						err = client.(*ecs.Client).RebootInstance(JsonData.InstanceId, true)
					}
					if cloud == "baidu" {
						err = client.(*bcc.Client).RebootInstance(JsonData.InstanceId, true)
					}
					if cloud == "volcengine" {
						_, err = client.(*volcEcs.ECS).RebootInstance(&volcEcs.RebootInstanceInput{InstanceId: &JsonData.InstanceId})
					}
				case "delete":
					if cloud == "aliyun" {
						err = client.(*ecs.Client).DeleteInstance(JsonData.InstanceId)
					}
					if cloud == "baidu" {
						err = client.(*bcc.Client).DeleteInstance(JsonData.InstanceId)
					}
					if cloud == "volcengine" {
						_, err = client.(*volcEcs.ECS).DeleteInstance(&volcEcs.DeleteInstanceInput{InstanceId: &JsonData.InstanceId})
					}
					if err == nil {
						db.Where("instance_id=?", JsonData.InstanceId).Delete(&CloudServer)
					}
				default:
					err = errors.New(JsonData.Operate + "无效的操作")
				}
			}
		} else {
			err = errors.New(JsonData.KeyId + "不存在")
		}
	}
}
