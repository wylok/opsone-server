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
