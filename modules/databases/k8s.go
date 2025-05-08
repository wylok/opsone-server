package databases

import (
	"time"
)

type K8sCluster struct {
	Id            uint64    `gorm:"primary_key" json:"id"`
	K8sId         string    `gorm:"column:k8s_id;type:varchar(100);uniqueIndex" json:"k8s_id"`
	K8sName       string    `gorm:"column:k8s_name;type:varchar(100);uniqueIndex" json:"k8s_name"`
	K8sConfig     string    `gorm:"column:k8s_config;type:text" json:"k8s_config"`
	AlarmChannel  string    `gorm:"column:alarm_channel;type:varchar(100)" json:"alarm_channel"`
	AlarmContacts string    `gorm:"column:alarm_contacts;type:varchar(100)" json:"alarm_contacts"`
	CreateTime    time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (K8sCluster) TableName() string {
	return "k8s_cluster"
}

type K8sAlarm struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	K8sId        string    `gorm:"column:k8s_id;type:varchar(100);index" json:"k8s_id"`
	K8sName      string    `gorm:"column:k8s_name;type:varchar(100);index" json:"k8s_name"`
	NameSpace    string    `gorm:"column:namespace;type:varchar(100);index" json:"namespace"`
	PodName      string    `gorm:"column:pod_name;type:varchar(100)" json:"pod_name"`
	AlarmMessage string    `gorm:"column:alarm_message;type:varchar(300)" json:"alarm_message"`
	CreateTime   time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (K8sAlarm) TableName() string {
	return "k8s_alarm"
}
