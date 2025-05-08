package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/k8s"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func K8sGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/k8s")
	{
		v1.Use(middleware.VerifyToken())
		v1.Use(middleware.VerifyPermission())
		v1.Use(middleware.Audit())
		v1.GET("/nodes", k8s.Nodes)
		v1.GET("/metric", k8s.MetricChart)
		v1.GET("/overview", k8s.OverView)
		v1.GET("/node/detail", k8s.NodeDetail)
		v1.GET("/node/yaml", k8s.NodeYaml)
		v1.POST("/node/taint", k8s.NodeTaint)
		v1.DELETE("/node", k8s.DeleteNode)
		v1.PUT("/node", k8s.UpdateNode)
		v1.GET("/namespaces", k8s.NameSpaces)
		v1.GET("/deployments", k8s.Deployments)
		v1.PUT("/deployment", k8s.UpdateDeployment)
		v1.DELETE("/deployment", k8s.DeleteDeployment)
		v1.POST("/deployment/restart", k8s.RestartDeployment)
		v1.GET("/deployment/yaml", k8s.DeploymentYaml)
		v1.GET("/replicaSets", k8s.ReplicaSets)
		v1.GET("/daemonSets", k8s.DaemonSets)
		v1.GET("/daemonSet/yaml", k8s.DaemonSetYaml)
		v1.POST("/daemonSet/restart", k8s.RestartDaemonSet)
		v1.PUT("/daemonSet", k8s.UpdateDaemonSet)
		v1.DELETE("/daemonSet", k8s.DeleteDaemonSet)
		v1.GET("/statefulSets", k8s.StatefulSets)
		v1.GET("/statefulSet/yaml", k8s.StatefulSetYaml)
		v1.GET("/services", k8s.Services)
		v1.GET("/service/yaml", k8s.ServiceYaml)
		v1.GET("/ingress", k8s.Ingress)
		v1.GET("/endpoints", k8s.Endpoints)
		v1.DELETE("/endpoint", k8s.DeleteEndpoint)
		v1.GET("/events", k8s.Events)
		v1.GET("/configMaps", k8s.ConfigMaps)
		v1.GET("/configMap/yaml", k8s.ConfigMapYaml)
		v1.POST("/configMap", k8s.CreatConfigMap)
		v1.PUT("/configMap", k8s.UpdateConfigMap)
		v1.GET("/secrets", k8s.Secrets)
		v1.POST("/secret", k8s.CreatSecret)
		v1.GET("/jobs", k8s.Jobs)
		v1.GET("/cronjobs", k8s.CronJobs)
		v1.GET("/limitRanges", k8s.LimitRanges)
		v1.GET("/serviceAccounts", k8s.ServiceAccounts)
		v1.GET("/autoScalers", k8s.Autoscalers)
		v1.PUT("/autoScaler", k8s.UpdateAutoscaler)
		v1.DELETE("/autoScaler", k8s.DeleteAutoscaler)
		v1.GET("/pods", k8s.Pods)
		v1.GET("/pod/event", k8s.GetPodEvent)
		v1.GET("/pod/yaml", k8s.PodYaml)
		v1.GET("/cluster", k8s.QueryK8sCluster)
		v1.GET("/role/cluster", k8s.ClusterRoles)
		v1.GET("/role/cluster/binding", k8s.ClusterRoleBindings)
		v1.POST("/cluster", k8s.UploadK8sCluster)
		v1.DELETE("/cluster", k8s.DelK8sCluster)
		v1.DELETE("/namespace", k8s.DeleteNamespace)
		v1.DELETE("/pod", k8s.DeletePod)
		v1.DELETE("/service", k8s.DeleteService)
		v1.DELETE("/statefulSet", k8s.DeleteStatefulSet)
		v1.DELETE("/configMap", k8s.DeleteConfigMap)
		v1.DELETE("/secret", k8s.DeleteSecret)
		v1.DELETE("/event", k8s.DeleteEvent)
		v1.DELETE("/job", k8s.DeleteJob)
		v1.DELETE("/cronjob", k8s.DeleteCronjob)
		v1.DELETE("/ingress", k8s.DeleteIngress)
		v1.DELETE("/limitRange", k8s.DeleteLimitRange)
		v1.DELETE("/serviceAccount", k8s.DeleteServiceAccount)
		v1.POST("/namespace", k8s.AddNamespace)
		v1.PUT("/cluster", k8s.ModifyK8sCluster)
	}
	v2 := r.Group("/api/v1/k8s")
	{
		v2.Use(middleware.VerifyToken())
		v2.GET("/connect", k8s.WsHandler)
	}
}

func init() {
	platform_conf.RouteNames["k8s.Nodes"] = "node列表"
	platform_conf.RouteNames["k8s.NodeDetail"] = "node详情"
	platform_conf.RouteNames["k8s.NodeYaml"] = "nodeYaml"
	platform_conf.RouteNames["k8s.NodeTaint"] = "node污点"
	platform_conf.RouteNames["k8s.DeleteNode"] = "删除node"
	platform_conf.RouteNames["k8s.UpdateNode"] = "变更node"
	platform_conf.RouteNames["k8s.NameSpaces"] = "NameSpace列表"
	platform_conf.RouteNames["k8s.AddNamespace"] = "创建NameSpace"
	platform_conf.RouteNames["k8s.DeleteNamespace"] = "删除NameSpace"
	platform_conf.RouteNames["k8s.Deployments"] = "Deployment列表"
	platform_conf.RouteNames["k8s.UpdateDeployment"] = "修改Deployment"
	platform_conf.RouteNames["k8s.DeleteDeployment"] = "删除Deployment"
	platform_conf.RouteNames["k8s.RestartDeployment"] = "重启Deployment"
	platform_conf.RouteNames["k8s.DeploymentYaml"] = "DeploymentYaml"
	platform_conf.RouteNames["k8s.ReplicaSets"] = "ReplicaSet列表"
	platform_conf.RouteNames["k8s.DaemonSets"] = "DaemonSet列表"
	platform_conf.RouteNames["k8s.DaemonSetYaml"] = "DaemonSetYaml"
	platform_conf.RouteNames["k8s.RestartDaemonSet"] = "重启DaemonSet"
	platform_conf.RouteNames["k8s.UpdateDaemonSet"] = "修改DaemonSet"
	platform_conf.RouteNames["k8s.DeleteDaemonSet"] = "删除DaemonSet"
	platform_conf.RouteNames["k8s.StatefulSets"] = "StatefulSet列表"
	platform_conf.RouteNames["k8s.StatefulSetYaml"] = "StatefulSetYaml"
	platform_conf.RouteNames["k8s.DeleteStatefulSet"] = "删除StatefulSet"
	platform_conf.RouteNames["k8s.Services"] = "service列表"
	platform_conf.RouteNames["k8s.ServiceYaml"] = "serviceYaml"
	platform_conf.RouteNames["k8s.DeleteService"] = "删除service"
	platform_conf.RouteNames["k8s.Ingress"] = "ingress列表"
	platform_conf.RouteNames["k8s.DeleteIngress"] = "删除ingress"
	platform_conf.RouteNames["k8s.Endpoints"] = "endpoint列表"
	platform_conf.RouteNames["k8s.DeleteEndpoint"] = "删除endpoint"
	platform_conf.RouteNames["k8s.Events"] = "event列表"
	platform_conf.RouteNames["k8s.DeleteEvent"] = "删除event"
	platform_conf.RouteNames["k8s.ConfigMaps"] = "configMap列表"
	platform_conf.RouteNames["k8s.ConfigMapYaml"] = "configMapYaml"
	platform_conf.RouteNames["k8s.CreatConfigMap"] = "新增configMap"
	platform_conf.RouteNames["k8s.UpdateConfigMap"] = "更新configMap"
	platform_conf.RouteNames["k8s.DeleteConfigMap"] = "删除configMap"
	platform_conf.RouteNames["k8s.Secrets"] = "secret列表"
	platform_conf.RouteNames["k8s.CreatSecret"] = "新增secret"
	platform_conf.RouteNames["k8s.DeleteSecret"] = "删除secret"
	platform_conf.RouteNames["k8s.Jobs"] = "job列表"
	platform_conf.RouteNames["k8s.DeleteJob"] = "删除job"
	platform_conf.RouteNames["k8s.CronJobs"] = "cronjob列表"
	platform_conf.RouteNames["k8s.DeleteCronjob"] = "删除cronjob"
	platform_conf.RouteNames["k8s.LimitRanges"] = "limitRange列表"
	platform_conf.RouteNames["k8s.DeleteLimitRange"] = "删除limitRange"
	platform_conf.RouteNames["k8s.ServiceAccounts"] = "serviceAccount列表"
	platform_conf.RouteNames["k8s.DeleteServiceAccount"] = "删除serviceAccount"
	platform_conf.RouteNames["k8s.Autoscalers"] = "autoScaler列表"
	platform_conf.RouteNames["k8s.UpdateAutoscaler"] = "配置autoScaler"
	platform_conf.RouteNames["k8s.DeleteAutoscaler"] = "删除autoScaler"
	platform_conf.RouteNames["k8s.Pods"] = "Pod列表"
	platform_conf.RouteNames["k8s.GetPodEvent"] = "Pod事件"
	platform_conf.RouteNames["k8s.PodYaml"] = "PodYaml"
	platform_conf.RouteNames["k8s.OverView"] = "K8S集群概览"
	platform_conf.RouteNames["k8s.QueryK8sCluster"] = "查询K8S集群"
	platform_conf.RouteNames["k8s.UploadK8sCluster"] = "上传配置文件"
	platform_conf.RouteNames["k8s.DelK8sCluster"] = "删除K8S集群"
	platform_conf.RouteNames["k8s.DeletePod"] = "删除集群pod"
	platform_conf.RouteNames["k8s.ClusterRoles"] = "ClusterRoles列表"
	platform_conf.RouteNames["k8s.ClusterRoleBindings"] = "ClusterRoleBindings列表"
	platform_conf.RouteNames["k8s.ModifyK8sCluster"] = "修改k8s集群配置"
}
