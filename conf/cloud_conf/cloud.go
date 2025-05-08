package cloud_conf

type QOss struct {
	Cloud   string `form:"cloud"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}
type QKey struct {
	Cloud   string `form:"cloud"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}
type CloudKey struct {
	Cloud     string `json:"cloud" binding:"required"`      //公有云
	KeyId     string `json:"key_id" binding:"required"`     //密钥ID
	KeySecret string `json:"key_secret" binding:"required"` //密钥串
	KeyType   string `json:"key_type" binding:"required"`   //密钥类型
	EndPoint  string `json:"end_point" binding:"required"`  //end_point
}
type DelCloudKey struct {
	KeyId string `json:"key_id" binding:"required"` //密钥ID
}
type OperateCloudServer struct {
	InstanceId string `json:"instance_id" binding:"required"` //实列ID
	KeyId      string `json:"key_id" binding:"required"`      //密钥ID
	Operate    string `json:"operate" binding:"required"`     //操作
}
