package cmdb_conf

var (
	HostType = map[string]string{"Alibaba": "cvm", "Baidu": "cvm", "HVM": "vm", "KVM": "vm",
		"VMware": "vm", "Tencent": "cvm", "Huawei": "cvm", "ByteDance": "cvm", "OpenStack": "vm", "Xen": "vm"}
	HostTypeCn = map[string]string{"physical": "物理机", "vm": "虚拟机", "cvm": "云主机"}
)

type Server struct {
	DepartmentId string   `form:"department_id"`
	AssetGroupId string   `form:"asset_group_id"`
	HostIds      []string `form:"host_ids"`
	HostType     string   `form:"host_type"`
	HostName     string   `form:"host_name"`
	SN           string   `form:"sn"`
	Ip           string   `form:"ip"`
	AssetTag     string   `form:"asset_tag"`
	Status       string   `form:"status"`
	Page         int      `form:"page"`
	PerPage      int      `form:"pre_page"`
}
type AssetGroup struct {
	GroupId   string `form:"group_id"`
	GroupName string `form:"group_name"`
	NoTPage   bool   `form:"not_page"`
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type RelatedGroupServer struct {
	GroupIds []string `form:"group_ids"`
	HostIds  []string `form:"host_ids"`
}
type GroupServers struct {
	GroupIds []string `form:"group_ids" binding:"required"`
}
type GroupNoneServers struct {
	GroupId string `form:"group_id"`
}
type UpdateServer struct {
	HostId      string `json:"host_id" binding:"required"`
	NickName    string `json:"nick_name"`
	HostType    string `json:"host_type"`
	AssetTag    string `json:"asset_tag"`
	IdcId       string `json:"idc_id"`
	Ipmi        string `json:"ipmi"`
	Cabinet     string `json:"cabinet"`
	BuyTime     string `json:"buy_time"`
	ExpiredTime string `json:"expired_time"`
}
type ServerPool struct {
	HostName string `form:"host_name"`
	Page     int    `form:"page"`
	PerPage  int    `form:"pre_page"`
}
type AssignAssetPool struct {
	AssetIds     []string `json:"asset_ids" binding:"required"`     //资源ID列表
	AssetTag     string   `json:"asset_tag"`                        //资源标签
	DepartmentId string   `json:"department_id" binding:"required"` //部门ID
	BusinessId   string   `json:"business_id" binding:"required"`   //业务ID
}
type Asset struct {
	AssetIds []string `json:"asset_ids" binding:"required"` //资源ID列表
}
type AssetBusiness struct {
	AssetIds     []string `json:"asset_ids" binding:"required"`     //资源ID列表
	DepartmentId string   `json:"department_id" binding:"required"` //部门ID
	BusinessId   string   `json:"business_id" binding:"required"`   //业务ID
}
type CreateAssetGroup struct {
	GroupName    string   `json:"group_name" binding:"required"`    //资源组名称
	DepartmentId string   `json:"department_id" binding:"required"` //部门ID
	AssetType    string   `json:"asset_type"`                       //资源类型(server|switch)
	AssetIds     []string `json:"asset_ids"`                        //资源ID列表
}
type ChangeAssetGroup struct {
	GroupId      string   `json:"group_id" binding:"required"` //资源组ID
	GroupName    string   `json:"group_name"`                  //资源组名称
	AssetIds     []string `json:"asset_ids"`                   //资源ID列表
	DepartmentId string   `json:"department_id"`               //部门ID
}
type DeleteAssetGroup struct {
	GroupIds []string `json:"group_ids" binding:"required"` //资源组ID
}
type AddSwitchPool struct {
	StartIp        string `json:"start_ip" binding:"required"`        //交换机起始IP
	EndIp          string `json:"end_ip" binding:"required"`          //交换机结束IP
	SwitchUser     string `json:"switch_user" binding:"required"`     //ssh用户名
	SwitchPassword string `json:"switch_password" binding:"required"` //ssh密码
	SwitchPort     int    `json:"switch_port" binding:"required"`     //ssh端口
	IdcId          string `json:"idc_id" binding:"required"`          //IDC_ID
}
type ModifySwitchPool struct {
	ID             uint64 `json:"id" binding:"required"` //id
	StartIp        string `json:"start_ip"`              //交换机起始IP
	EndIp          string `json:"end_ip"`                //交换机结束IP
	SwitchUser     string `json:"switch_user"`           //ssh用户名
	SwitchPassword string `json:"switch_password"`       //ssh密码
	SwitchPort     int    `json:"switch_port"`           //ssh端口
	IdcId          string `json:"idc_id"`                //IDC_ID
	SwitchStatus   string `json:"switch_status"`         //状态(启用|禁用)
}
type DeleteSwitchPool struct {
	ID uint64 `json:"id" binding:"required"` //id
}
type SwitchPool struct {
	Page    int `form:"page"`
	PerPage int `form:"pre_page"`
}
type QuerySwitch struct {
	SwitchId   string `form:"switch_id"`
	SwitchIp   string `form:"switch_ip"`
	SwitchName string `form:"switch_name"`
	HostMac    string `form:"host_mac"`
	Page       int    `form:"page"`
	PerPage    int    `form:"pre_page"`
}
type QuerySwitchPort struct {
	SwitchId   string `form:"switch_id" binding:"required"`
	PortName   string `form:"port_name"`
	MacAddress string `form:"mac_address"`
	Page       int    `form:"page"`
	PerPage    int    `form:"pre_page"`
}
type QuerySwitchVlan struct {
	SwitchId string `form:"switch_id" binding:"required"`
}
type CloudServer struct {
	Cloud        string `form:"cloud" binding:"required"`
	InstanceId   string `form:"instance_id"`
	InstanceName string `form:"instance_name"`
	HostName     string `form:"host_name"`
	SN           string `form:"sn"`
	Status       string `form:"status"`
	Page         int    `form:"page"`
	PerPage      int    `form:"pre_page"`
}
type AddSwitchVlan struct {
	SwitchIps []string `json:"switch_ips" binding:"required"` //交换机IP
	Vlan      string   `json:"vlan" binding:"required"`       //vlan
}
type ChangeSwitchPortVlan struct {
	SwitchId string `json:"switch_id" binding:"required"` //交换机ID
	PortName string `json:"port_name" binding:"required"` //交换机端口名称
	NewVlan  string `json:"new_vlan" binding:"required"`  //新vlan
}
type SwitchPortOperate struct {
	SwitchId string `json:"switch_id" binding:"required"` //交换机ID
	PortName string `json:"port_name" binding:"required"` //交换机端口名称
	Operate  string `json:"operate" binding:"required"`   //操作(UP/DOWN)
}
type SwitchOperate struct {
	SwitchId string `json:"switch_id" binding:"required"` //交换机ID
	Commands string `json:"commands" binding:"required"`  //操作命令
}
type SwitchName struct {
	SwitchId string `json:"switch_id" binding:"required"` //交换机ID
	Name     string `json:"name" binding:"required"`      //交换机名称
}
type AddServerIpPool struct {
	StartIp     string `json:"start_ip"`                    //交换机起始IP
	EndIp       string `json:"end_ip"`                      //交换机结束IP
	SshUser     string `json:"ssh_user" binding:"required"` //ssh用户名
	SshPassword string `json:"ssh_password"`                //ssh密码
	SshKeyName  string `json:"ssh_key_name"`                //ssh密钥名称
	SshPort     int    `json:"ssh_port" binding:"required"` //ssh端口
	IdcId       string `json:"idc_id" binding:"required"`   //IDC_ID
}
type ServerIpPool struct {
	Page    int `form:"page"`
	PerPage int `form:"pre_page"`
}
type DeleteServerIpPool struct {
	ID uint64 `json:"id" binding:"required"` //id
}
type ModifyServerIpPool struct {
	ID          uint64 `json:"id" binding:"required"` //id
	StartIp     string `json:"start_ip"`              //服务器起始IP
	EndIp       string `json:"end_ip"`                //服务器结束IP
	SshUser     string `json:"ssh_user"`              //ssh用户名
	SshPassword string `json:"ssh_password"`          //ssh密码
	SshKeyName  string `json:"ssh_key_name"`          //ssh密钥名称
	SshPort     int    `json:"ssh_port"`              //ssh端口
	Status      string `json:"status"`                //状态(启用|禁用)
	IdcId       string `json:"idc_id"`                //IDC_ID
}
type SshKey struct {
	KeyName string `form:"key_name"`
	SshUser string `form:"ssh_user"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}
type DeleteSshKey struct {
	KeyName string `json:"key_name" binding:"required"` //密钥名称
}
type UploadSshKey struct {
	KeyName   string `json:"key_name" binding:"required"`   //密钥名称
	KeyConfig string `json:"key_config" binding:"required"` //密钥内容
}
type Idc struct {
	Idc     string `form:"idc"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}
type DeleteIdc struct {
	IdcId string `json:"idc_id" binding:"required"` //IDC_ID
}
type AddIdc struct {
	Idc        string `json:"idc" binding:"required"`         //idc名称
	IdcCn      string `json:"idc_cn" binding:"required"`      //idc中文
	Region     string `json:"region" binding:"required"`      //地区名称
	RegionCn   string `json:"region_cn" binding:"required"`   //地区中文
	DataCenter string `json:"data_center" binding:"required"` //数据中心名称
}
type ModifyIdc struct {
	IdcId      string `json:"idc_id" binding:"required"` //idc_id
	Idc        string `json:"idc"`                       //idc名称
	IdcCn      string `json:"idc_cn"`                    //idc中文
	Region     string `json:"region"`                    //地区名称
	RegionCn   string `json:"region_cn"`                 //地区中文
	DataCenter string `json:"data_center"`               //数据中心名称
}
type QuerySwitchRelation struct {
	SwitchId string `form:"switch_id"`
}
type JumpServerKey struct {
	ServerUrl string `form:"server_url"`
	KeyId     string `form:"key_id"`
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type UploadJumpServerKey struct {
	ServerUrl string `json:"server_url" binding:"required"` //堡垒机地址
	KeyId     string `json:"key_id" binding:"required"`     //堡垒机密钥ID
	SecretId  string `json:"secret_id" binding:"required"`  //堡垒机密钥
}
type DeleteJumpServerKey struct {
	KeyId string `json:"key_id" binding:"required"` //堡垒机密钥ID
}
