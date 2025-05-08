package databases

import (
	"time"
)

type CmdbPartition struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	ObjectId     string `gorm:"column:object_id;type:varchar(100);uniqueIndex" json:"object_id"`
	ObjectType   string `gorm:"column:object_type;type:varchar(100);index" json:"object_type"`
	DepartmentId string `gorm:"column:department_id;type:varchar(100);index" json:"department_id"`
}

func (CmdbPartition) TableName() string {
	return "cmdb_partition"
}

type AssetIdc struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	IdcId      string    `gorm:"column:idc_id;type:varchar(100);uniqueIndex" json:"idc_id"`
	Idc        string    `gorm:"column:idc;type:varchar(100)" json:"idc"`
	IdcCn      string    `gorm:"column:idc_cn;type:varchar(100)" json:"idc_cn"`
	Region     string    `gorm:"column:region;type:varchar(100)" json:"region"`
	RegionCn   string    `gorm:"column:region_cn;type:varchar(100)" json:"region_cn"`
	DataCenter string    `gorm:"column:data_center;type:varchar(100)" json:"data_center"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime" json:"update_time"`
}

func (AssetIdc) TableName() string {
	return "asset_idc"
}

type AssetServer struct {
	Id              uint64    `gorm:"primary_key" json:"id"`
	HostId          string    `gorm:"column:host_id;type:varchar(100);uniqueIndex" json:"host_id"`
	NickName        string    `gorm:"column:nick_name;type:varchar(50);index" json:"nick_name"`
	Hostname        string    `gorm:"column:host_name;type:varchar(100);index" json:"host_name"`
	Sn              string    `gorm:"column:sn;type:varchar(100)" json:"sn"`
	ProductName     string    `gorm:"column:product_name;type:varchar(50)" json:"product_name"`
	Manufacturer    string    `gorm:"column:manufacturer;type:varchar(100)" json:"manufacturer"`
	HostType        string    `gorm:"column:host_type;type:varchar(100);index" json:"host_type"`
	HostTypeCn      string    `gorm:"column:host_type_cn;type:varchar(50)" json:"host_type_cn"`
	Cpu             int       `gorm:"column:cpu;type:int(8)" json:"cpu"`
	CpuInfo         string    `gorm:"column:cpu_info;type:varchar(50)" json:"cpu_info"`
	Memory          uint64    `gorm:"column:memory;type:bigint(50)" json:"memory"`
	Disk            uint64    `gorm:"column:disk;type:bigint(50)" json:"disk"`
	Os              string    `gorm:"column:os;type:varchar(100)" json:"os"`
	Platform        string    `gorm:"column:platform;type:varchar(50);index" json:"platform"`
	PlatformVersion string    `gorm:"column:platform_version;type:varchar(100)" json:"platform_version"`
	KernelVersion   string    `gorm:"column:kernel_version;type:varchar(50)" json:"kernel_version"`
	InternetIp      string    `gorm:"column:internet_ip;type:varchar(100)" json:"internet_ip"`
	PoolId          int       `gorm:"column:pool_id;type:int(8)" json:"pool_id"`
	CreateTime      time.Time `gorm:"column:create_time;type:datetime;index" json:"create_time"`
	UpdateTime      time.Time `gorm:"column:update_time;type:datetime" json:"update_time"`
	AssetTag        string    `gorm:"column:asset_tag;type:varchar(100);index" json:"asset_tag"`
	AssetStatus     string    `gorm:"column:asset_status;type:varchar(100)" json:"asset_status"`
}

func (AssetServer) TableName() string {
	return "asset_server"
}

type AssetNet struct {
	Id        uint64 `gorm:"primary_key" json:"id"`
	HostId    string `gorm:"column:host_id;type:varchar(100)" json:"host_id"`
	Name      string `gorm:"column:name;type:varchar(100)" json:"name"`
	Addr      string `gorm:"column:addr;type:varchar(100)" json:"addr"`
	Ip        string `gorm:"column:ip;type:varchar(50);index" json:"ip"`
	Netmask   string `gorm:"column:netmask;type:varchar(50)" json:"netmask"`
	Md5Verify string `gorm:"column:md5_verify;type:varchar(100);uniqueIndex" json:"md5_verify"`
}

func (AssetNet) TableName() string {
	return "asset_net"
}

type AssetDisk struct {
	Id         uint64 `gorm:"primary_key" json:"id"`
	HostId     string `gorm:"column:host_id;type:varchar(100)" json:"host_id"`
	DiskName   string `gorm:"column:disk_name;type:varchar(100);index" json:"disk_name"`
	MountPoint string `gorm:"column:mount_point;type:varchar(50)" json:"mount_point"`
	FsType     string `gorm:"column:fs_type;type:varchar(200)" json:"fs_type"`
	DiskSize   uint64 `gorm:"column:disk_size;type:varchar(200)" json:"disk_size"`
	Md5Verify  string `gorm:"column:md5_verify;type:varchar(200);uniqueIndex" json:"md5_verify"`
}

func (AssetDisk) TableName() string {
	return "asset_disk"
}

type AssetExtend struct {
	Id          uint64    `gorm:"primary_key" json:"id"`
	HostId      string    `gorm:"column:host_id;type:varchar(100)" json:"host_id"`
	IdcId       string    `gorm:"column:idc_id;type:varchar(100)" json:"idc_id"`
	Ipmi        string    `gorm:"column:ipmi_ip;type:varchar(50)" json:"ipmi"`
	Cabinet     string    `gorm:"column:cabinet;type:varchar(50)" json:"cabinet"`
	BuyTime     time.Time `gorm:"column:buy_time;type:datetime" json:"buy_time"`
	ExpiredTime time.Time `gorm:"column:expired_time;type:datetime" json:"expired_time"`
}

func (AssetExtend) TableName() string {
	return "asset_extend"
}

type AssetGroups struct {
	Id        uint64 `gorm:"primary_key" json:"id"`
	GroupId   string `gorm:"column:group_id;type:varchar(100);uniqueIndex" json:"group_id"`
	GroupName string `gorm:"column:group_name;type:varchar(100);uniqueIndex" json:"group_name"`
	Status    string `gorm:"column:status;type:enum('active','close')" json:"status"`
}

func (AssetGroups) TableName() string {
	return "asset_groups"
}

type GroupServer struct {
	Id      uint64 `gorm:"primary_key" json:"id"`
	GroupId string `gorm:"column:group_id;type:varchar(100);index" json:"group_id"`
	HostId  string `gorm:"column:host_id;type:varchar(100);index" json:"host_id"`
}

func (GroupServer) TableName() string {
	return "group_server"
}

type AssetUnder struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	AssetId      string `gorm:"column:asset_id;type:varchar(100);uniqueIndex" json:"asset_id"`
	AssetType    string `gorm:"column:asset_type;type:varchar(100)" json:"asset_type"`
	DepartmentId string `gorm:"column:department_id;type:varchar(100);index" json:"department_id"`
	BusinessId   string `gorm:"column:business_id;type:varchar(100);index" json:"business_id"`
}

func (AssetUnder) TableName() string {
	return "asset_under"
}

type AssetSwitch struct {
	Id            uint64    `gorm:"primary_key" json:"id"`
	SwitchPoolId  uint64    `gorm:"column:switch_pool_id;type:bigint;Index" json:"switch_pool_id"`
	SwitchIp      string    `gorm:"column:switch_ip;type:varchar(100);uniqueIndex:switch_ip_id" json:"switch_ip"`
	SwitchId      string    `gorm:"column:switch_id;type:varchar(100);uniqueIndex:switch_ip_id" json:"switch_id"`
	SwitchBrand   string    `gorm:"column:switch_brand;type:varchar(50)" json:"switch_brand"`
	SwitchName    string    `gorm:"column:switch_name;type:varchar(100);index" json:"switch_name"`
	SwitchVersion string    `gorm:"column:switch_version;type:varchar(50)" json:"switch_version"`
	IdcId         string    `gorm:"column:idc_id;type:varchar(100)" json:"idc_id"`
	Status        string    `gorm:"column:status;type:enum('online','offline')" json:"status"`
	SyncTime      time.Time `gorm:"column:sync_time;type:datetime;default:CURRENT_TIMESTAMP" json:"sync_time"`
}

func (AssetSwitch) TableName() string {
	return "asset_switch"
}

type AssetSwitchVlan struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	SwitchId   string    `gorm:"column:switch_id;type:varchar(100);uniqueIndex:switch_vlan_id" json:"switch_id"`
	SwitchVlan uint32    `gorm:"column:switch_vlan;type:int;uniqueIndex:switch_vlan_id" json:"switch_vlan"`
	LastTime   time.Time `gorm:"column:last_time;type:datetime;default:CURRENT_TIMESTAMP" json:"last_time"`
}

func (AssetSwitchVlan) TableName() string {
	return "asset_switch_vlan"
}

type AssetSwitchPort struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	SwitchId   string    `gorm:"column:switch_id;type:varchar(100);uniqueIndex:switch_vlan_port" json:"switch_id"`
	PortName   string    `gorm:"column:port_name;type:varchar(50);uniqueIndex:switch_vlan_port" json:"port_name"`
	PortType   string    `gorm:"column:port_type;type:enum('Access','Trunk')" json:"port_type"`
	SwitchVlan uint32    `gorm:"column:switch_vlan;type:int" json:"switch_vlan"`
	MacAddress string    `gorm:"column:mac_address;type:varchar(50);index" json:"mac_address"`
	PortStat   string    `gorm:"column:port_stat;type:enum('UP','DOWN')" json:"port_stat"`
	LastTime   time.Time `gorm:"column:last_time;type:datetime;default:CURRENT_TIMESTAMP" json:"last_time"`
}

func (AssetSwitchPort) TableName() string {
	return "asset_switch_port"
}

type AssetSwitchPool struct {
	Id             uint64    `gorm:"primary_key" json:"id"`
	StartIp        string    `gorm:"column:start_ip;type:varchar(100);uniqueIndex:start_end_ip" json:"start_ip"`
	EndIp          string    `gorm:"column:end_ip;type:varchar(100);uniqueIndex:start_end_ip" json:"end_ip"`
	SwitchPort     int       `gorm:"column:switch_port;type:int;default:22" json:"switch_port"`
	SwitchUser     string    `gorm:"column:switch_user;type:varchar(100)" json:"switch_user"`
	SwitchPassword string    `gorm:"column:switch_password;type:varchar(100)" json:"switch_password"`
	Discover       int       `gorm:"column:discover;type:int;default:22" json:"discover"`
	IdcId          string    `gorm:"column:idc_id;type:varchar(100)" json:"idc_id"`
	SwitchStatus   string    `gorm:"column:switch_status;type:enum('enable','disable')" json:"switch_status"`
	CreateTime     time.Time `gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP" json:"create_time"`
	ModifyTime     time.Time `gorm:"column:modify_time;type:datetime;default:CURRENT_TIMESTAMP" json:"modify_time"`
	SyncTime       time.Time `gorm:"column:sync_time;type:datetime;default:CURRENT_TIMESTAMP" json:"sync_time"`
}

func (AssetSwitchPool) TableName() string {
	return "asset_switch_pool"
}

type AssetServerPool struct {
	Id          uint64    `gorm:"primary_key" json:"id"`
	StartIp     string    `gorm:"column:start_ip;type:varchar(100);uniqueIndex:start_end_ip" json:"start_ip"`
	EndIp       string    `gorm:"column:end_ip;type:varchar(100);uniqueIndex:start_end_ip" json:"end_ip"`
	SshPort     int       `gorm:"column:ssh_port;type:int;default:22" json:"ssh_port"`
	SshUser     string    `gorm:"column:ssh_user;type:varchar(100)" json:"ssh_user"`
	SshPassword string    `gorm:"column:ssh_password;type:varchar(100)" json:"ssh_password"`
	SshKeyName  string    `gorm:"column:ssh_key_name;type:varchar(100)" json:"ssh_key_name"`
	Discover    int       `gorm:"column:discover;type:int;default:22" json:"discover"`
	IdcId       string    `gorm:"column:idc_id;type:varchar(100)" json:"idc_id"`
	Status      string    `gorm:"column:status;type:enum('enable','disable')" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP" json:"create_time"`
	ModifyTime  time.Time `gorm:"column:modify_time;type:datetime;default:CURRENT_TIMESTAMP" json:"modify_time"`
	SyncTime    time.Time `gorm:"column:sync_time;type:datetime;default:CURRENT_TIMESTAMP" json:"sync_time"`
}

func (AssetServerPool) TableName() string {
	return "asset_server_pool"
}

type SshKey struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	KeyName    string    `gorm:"column:key_name;type:varchar(100);uniqueIndex" json:"key_name"`
	SshUser    string    `gorm:"column:ssh_user;type:varchar(100);index" json:"ssh_user"`
	SshKey     string    `gorm:"column:ssh_key;type:text" json:"ssh_key"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP" json:"create_time"`
}

func (SshKey) TableName() string {
	return "ssh_key"
}

type AssetSwitchRelation struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	SwitchId     string    `gorm:"column:switch_id;type:varchar(100);uniqueIndex:switch_neighbor_id" json:"switch_id"`
	SwitchName   string    `gorm:"column:switch_name;type:varchar(100);index" json:"switch_name"`
	NeighborId   string    `gorm:"column:neighbor_id;type:varchar(100);uniqueIndex:switch_neighbor_id" json:"neighbor_id"`
	NeighborName string    `gorm:"column:neighbor_name;type:varchar(100);index" json:"neighbor_name"`
	UpdateTime   time.Time `gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (AssetSwitchRelation) TableName() string {
	return "asset_switch_relation"
}

type JumpServerKey struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	ServerUrl  string    `gorm:"column:server_url;type:varchar(100);Index" json:"server_url"`
	KeyId      string    `gorm:"column:key_id;type:varchar(100);uniqueIndex" json:"key_id"`
	SecretId   string    `gorm:"column:secret_id;type:varchar(100)" json:"secret_id"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP" json:"create_time"`
	SyncTime   time.Time `gorm:"column:sync_time;type:datetime;default:CURRENT_TIMESTAMP" json:"sync_time"`
}

func (JumpServerKey) TableName() string {
	return "jumpserver_key"
}
