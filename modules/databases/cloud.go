package databases

import (
	"time"
)

type CloudKeys struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	Cloud      string    `gorm:"column:cloud;type:varchar(100);uniqueIndex:cloud_id_type_point" json:"cloud"`
	KeyId      string    `gorm:"column:key_id;type:varchar(100);uniqueIndex:cloud_id_type_point" json:"key_id"`
	KeySecret  string    `gorm:"column:key_secret;type:varchar(100)" json:"key_secret"`
	KeyType    string    `gorm:"column:key_type;type:varchar(50);uniqueIndex:cloud_id_type_point" json:"key_type"`
	EndPoint   string    `gorm:"column:end_point;type:varchar(100);uniqueIndex:cloud_id_type_point" json:"end_point"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (CloudKeys) TableName() string {
	return "cloud_keys"
}

type CloudOss struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	Cloud        string    `gorm:"column:cloud;type:varchar(100);uniqueIndex:cloud_key_id" json:"cloud"`
	KeyId        string    `gorm:"column:key_id;type:varchar(100);uniqueIndex:cloud_key_id" json:"key_id"`
	Bucket       string    `gorm:"column:bucket;type:varchar(100)" json:"bucket"`
	Location     string    `gorm:"column:Location;type:varchar(100)" json:"Location"`
	StorageClass string    `gorm:"column:StorageClass;type:varchar(100)" json:"StorageClass"`
	CreationDate time.Time `gorm:"column:CreationDate;type:datetime" json:"CreationDate"`
	Storage      int64     `gorm:"column:Storage;type:bigint(24)" json:"Storage"`
	ObjectCount  int64     `gorm:"column:ObjectCount;type:int(8)" json:"ObjectCount"`
	SyncTime     time.Time `gorm:"column:sync_time;type:datetime" json:"sync_time"`
}

func (CloudOss) TableName() string {
	return "cloud_oss"
}

type CloudServer struct {
	Id              uint64    `gorm:"primary_key" json:"id"`
	Cloud           string    `gorm:"column:cloud;type:varchar(100)" json:"cloud"`
	InstanceId      string    `gorm:"column:instance_id;type:varchar(100);uniqueIndex" json:"instance_id"`
	InstanceName    string    `gorm:"column:instance_name;type:varchar(100);index" json:"instance_name"`
	InstanceType    string    `gorm:"column:instance_type;type:varchar(100)" json:"instance_type"`
	Description     string    `gorm:"column:description;type:varchar(100)" json:"description"`
	HostName        string    `gorm:"column:host_name;type:varchar(100);index" json:"host_name"`
	Sn              string    `gorm:"column:sn;type:varchar(100)" json:"sn"`
	RegionId        string    `gorm:"column:region_id;type:varchar(100)" json:"region_id"`
	ZoneId          string    `gorm:"column:zone_id;type:varchar(100)" json:"zone_id"`
	Cpu             int       `gorm:"column:cpu;type:int" json:"cpu"`
	Memory          int       `gorm:"column:memory;type:int" json:"memory"`
	PublicIpAddress string    `gorm:"column:PublicIpAddress;type:varchar(100)" json:"PublicIpAddress"`
	InnerIpAddress  string    `gorm:"column:InnerIpAddress;type:varchar(100)" json:"InnerIpAddress"`
	Status          string    `gorm:"column:status;type:varchar(50)" json:"status"`
	CreationDate    time.Time `gorm:"column:CreationDate;type:datetime" json:"CreationDate"`
	ExpiredTime     time.Time `gorm:"column:ExpiredTime;type:datetime" json:"ExpiredTime"`
	KeyId           string    `gorm:"column:key_id;type:varchar(100)" json:"key_id"`
	SyncTime        time.Time `gorm:"column:sync_time;type:datetime" json:"sync_time"`
}

func (CloudServer) TableName() string {
	return "cloud_server"
}
