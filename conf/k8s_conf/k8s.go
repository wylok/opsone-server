package k8s_conf

type ListNodes struct {
	K8sID string `form:"k8s_id" binding:"required"`
}
type NodeDetail struct {
	K8sID    string `form:"k8s_id" binding:"required"`
	NodeName string `form:"node_name" binding:"required"`
}
type ListNameSpace struct {
	K8sID string `form:"k8s_id" binding:"required"`
}
type ListApplication struct {
	K8sID     string `form:"k8s_id" binding:"required"`
	NameSpace string `form:"namespace" binding:"required"`
}
type ListRoles struct {
	K8sID     string `form:"k8s_id" binding:"required"`
	NameSpace string `form:"namespace"`
}
type ListReplicaSet struct {
	K8sID     string `form:"k8s_id" binding:"required"`
	NameSpace string `form:"namespace" binding:"required"`
	Name      string `form:"name" binding:"required"`
}
type Application struct {
	K8sID     string `form:"k8s_id" binding:"required"`
	NameSpace string `form:"namespace" binding:"required"`
	Name      string `form:"name" binding:"required"`
}
type K8sMetric struct {
	K8sID     string `form:"k8s_id" binding:"required"`
	NameSpace string `form:"namespace"`
	Resource  string `form:"resource"`
	Name      string `form:"name"`
}
type NameSpace struct {
	K8sID     string `json:"k8s_id" binding:"required"`
	NameSpace string `json:"namespace" binding:"required"`
}
type App struct {
	K8sID     string `json:"k8s_id" binding:"required"`
	NameSpace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}
type ListAutoscalers struct {
	K8sID      string `form:"k8s_id" binding:"required"`
	NameSpace  string `form:"namespace" binding:"required"`
	Deployment string `form:"deployment"`
}
type ListPods struct {
	K8sID       string `form:"k8s_id" binding:"required"`
	NameSpace   string `form:"namespace" binding:"required"`
	Deployment  string `form:"deployment"`
	DaemonSet   string `form:"daemonSet"`
	StatefulSet string `form:"statefulSet"`
}
type K8sCluster struct {
	K8sName string `form:"k8s_name"`
}
type K8sAlarm struct {
	K8sName   string `form:"k8s_name"`
	NameSpace string `form:"namespace"`
	PodName   string `form:"pod_name"`
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type UploadK8sCluster struct {
	K8sName       string `json:"k8s_name" binding:"required"`
	K8sConfig     string `json:"k8s_config" binding:"required"`
	AlarmChannel  string `json:"alarm_channel"`
	AlarmContacts string `json:"alarm_contacts"`
}
type ModifyK8sCluster struct {
	K8sId         string `json:"k8s_id" binding:"required"`
	K8sName       string `json:"k8s_name" binding:"required"`
	AlarmChannel  string `json:"alarm_channel" binding:"required"`
	AlarmContacts string `json:"alarm_contacts" binding:"required"`
}
type DelK8sCluster struct {
	K8sID   string `json:"k8s_id" binding:"required"`
	K8sName string `json:"k8s_name" binding:"required"`
}
type DelK8sAlarm struct {
	Ids []uint64 `json:"ids" binding:"required"`
}
type TaintNode struct {
	K8sID    string `json:"k8s_id" binding:"required"`
	NodeName string `json:"node_name" binding:"required"`
	Effect   string `json:"effect"`
}
type DeleteNode struct {
	K8sID    string `json:"k8s_id" binding:"required"`
	NodeName string `json:"node_name" binding:"required"`
}
type UpdateNode struct {
	K8sID    string            `json:"k8s_id" binding:"required"`
	NodeName string            `json:"node_name" binding:"required"`
	Labels   map[string]string `json:"labels"`
}
type Deployment struct {
	K8sID     string              `json:"k8s_id" binding:"required"`
	NameSpace string              `json:"namespace" binding:"required"`
	Name      string              `json:"name" binding:"required"`
	Images    map[string][]string `json:"images"`
}
type DaemonSet struct {
	K8sID     string              `json:"k8s_id" binding:"required"`
	NameSpace string              `json:"namespace" binding:"required"`
	Name      string              `json:"name" binding:"required"`
	Images    map[string][]string `json:"images"`
}
type UpdateAutoscaler struct {
	K8sID     string            `json:"k8s_id" binding:"required"`
	NameSpace string            `json:"namespace" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Scaler    map[string]*int32 `json:"scaler" binding:"required"`
}
type Secrets struct {
	K8sID     string            `json:"k8s_id" binding:"required"`
	NameSpace string            `json:"namespace" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Type      string            `json:"type" binding:"required"`
	Data      map[string]string `json:"data" binding:"required"`
}
type ConfigMap struct {
	K8sID     string              `json:"k8s_id" binding:"required"`
	NameSpace string              `json:"namespace" binding:"required"`
	Name      string              `json:"name" binding:"required"`
	Data      []map[string]string `json:"data" binding:"required"`
}
