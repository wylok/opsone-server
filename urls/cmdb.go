package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/cmdb"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func CmdbGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/cmdb")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.GET("/servers", cmdb.QueryServer)
		v1.PUT("/servers", cmdb.UpdateServer)
		v1.GET("/pool/server", cmdb.QueryPoolServer)
		v1.POST("/pool/server", cmdb.AssignServerPool)
		v1.PUT("/pool/server", cmdb.ReclaimServerPool)
		v1.DELETE("/pool/server", cmdb.DiscardServerPool)
		v1.PUT("/pool/server/business", cmdb.ModifyAssetBusiness)
		v1.GET("/group", cmdb.QueryAssetGroup)
		v1.POST("/group", cmdb.CreateAssetGroup)
		v1.PUT("/group", cmdb.ChangeAssetGroup)
		v1.DELETE("/group", cmdb.DeleteAssetGroup)
		v1.GET("/group/servers", cmdb.GroupServers)
		v1.GET("/group/related/servers", cmdb.RelatedGroupServer)
		v1.GET("/group/servers/detail", cmdb.GroupServersDetail)
		v1.GET("/pool/switch", cmdb.QuerySwitchPool)
		v1.POST("/pool/switch", cmdb.AddSwitchPool)
		v1.PUT("/pool/switch", cmdb.ModifySwitchPool)
		v1.DELETE("/pool/switch", cmdb.DeleteSwitchPool)
		v1.GET("/switches", cmdb.QuerySwitches)
		v1.GET("/switch/port", cmdb.QuerySwitchPort)
		v1.GET("/switch/vlan", cmdb.QuerySwitchVlan)
		v1.POST("/switch/vlan", cmdb.AddSwitchVlan)
		v1.POST("/switch/port/vlan", cmdb.ChangeSwitchPortVlan)
		v1.POST("/switch/port/operate", cmdb.SwitchPortOperate)
		v1.POST("/switch/operate", cmdb.SwitchOperate)
		v1.POST("/switch/name", cmdb.SwitchName)
		v1.DELETE("/switch", cmdb.DeleteSwitch)
		v1.POST("/pool/server/ip", cmdb.AddServerIpPool)
		v1.GET("/pool/server/ip", cmdb.QueryServerIpPool)
		v1.DELETE("/pool/server/ip", cmdb.DeleteServerIpPool)
		v1.PUT("/pool/server/ip", cmdb.ModifyServerIpPool)
		v1.GET("/ssh_key", cmdb.QuerySshKey)
		v1.POST("/ssh_key", cmdb.SshKeyUpload)
		v1.DELETE("/ssh_key", cmdb.SshKeyDelete)
		v1.GET("/idc", cmdb.QueryIdc)
		v1.DELETE("/idc", cmdb.DeleteIdc)
		v1.POST("/idc", cmdb.AddIdc)
		v1.PUT("/idc", cmdb.ModifyIdc)
		v1.GET("/switch/relation", cmdb.QuerySwitchesRelation)
	}
	v2 := r.Group("/api/v1/cmdb")
	{
		v2.Use(middleware.VerifyToken())
		v2.GET("/connect", cmdb.WsHandler)
	}
}

func init() {
	platform_conf.RouteNames["cmdb.QueryServer"] = "查询主机"
	platform_conf.RouteNames["cmdb.UpdateServer"] = "配置主机"
	platform_conf.RouteNames["cmdb.QueryDiscardServer"] = "查询下架主机"
	platform_conf.RouteNames["cmdb.QueryPoolServer"] = "查询资源池"
	platform_conf.RouteNames["cmdb.AssignServerPool"] = "分配资源池"
	platform_conf.RouteNames["cmdb.ReclaimServerPool"] = "回收资源"
	platform_conf.RouteNames["cmdb.DiscardServerPool"] = "下架资源"
	platform_conf.RouteNames["cmdb.ModifyAssetBusiness"] = "变更资源业务组"
	platform_conf.RouteNames["cmdb.QueryAssetGroup"] = "查询资源组"
	platform_conf.RouteNames["cmdb.CreateAssetGroup"] = "新建资源组"
	platform_conf.RouteNames["cmdb.ChangeAssetGroup"] = "修改资源组"
	platform_conf.RouteNames["cmdb.DeleteAssetGroup"] = "删除资源组"
	platform_conf.RouteNames["cmdb.GroupServers"] = "查询资源组主机"
	platform_conf.RouteNames["cmdb.RelatedGroupServer"] = "查询主机关联资源组"
	platform_conf.RouteNames["cmdb.GroupServersDetail"] = "查询资源组主机详情"
	platform_conf.RouteNames["cmdb.AddSwitchPool"] = "配置交换机信息"
	platform_conf.RouteNames["cmdb.QuerySwitchPool"] = "查询交换机资源池"
	platform_conf.RouteNames["cmdb.ModifySwitchPool"] = "修改交换机资源池"
	platform_conf.RouteNames["cmdb.DeleteSwitchPool"] = "删除交换机资源池"
	platform_conf.RouteNames["cmdb.QuerySwitches"] = "查询交换机列表"
	platform_conf.RouteNames["cmdb.AddSwitchVlan"] = "新增交换机vlan"
	platform_conf.RouteNames["cmdb.DeleteSwitch"] = "删除交换机"
	platform_conf.RouteNames["cmdb.ChangeSwitchPortVlan"] = "变更交换机端口vlan"
	platform_conf.RouteNames["cmdb.SwitchPortOperate"] = "开启/关闭交换机端口"
	platform_conf.RouteNames["cmdb.SwitchOperate"] = "交换机执行命令"
	platform_conf.RouteNames["cmdb.SwitchName"] = "修改交换机名称"
	platform_conf.RouteNames["cmdb.AddSwitchPool"] = "配置服务器IP池"
	platform_conf.RouteNames["cmdb.QueryServerIpPool"] = "查询服务器IP池"
	platform_conf.RouteNames["cmdb.DeleteServerIpPool"] = "删除服务器IP池"
	platform_conf.RouteNames["cmdb.ModifyServerIpPool"] = "修改服务器IP池"
	platform_conf.RouteNames["cmdb.QuerySshKey"] = "查询SSH密钥"
	platform_conf.RouteNames["cmdb.SshKeyUpload"] = "上传SSH密钥"
	platform_conf.RouteNames["cmdb.SshKeyDelete"] = "删除SSH密钥"
	platform_conf.RouteNames["cmdb.QueryIdc"] = "查询IDC配置"
	platform_conf.RouteNames["cmdb.DeleteIdc"] = "删除IDC配置"
	platform_conf.RouteNames["cmdb.AddIdc"] = "新增IDC配置"
	platform_conf.RouteNames["cmdb.ModifyIdc"] = "修改IDC配置"
	platform_conf.RouteNames["cmdb.QuerySwitchesRelation"] = "查询交换机级联"
}
