package job_conf

var SendFileJobKey = "send_file_job_key"

type RExec struct {
	HostIds       []string `json:"host_ids"`                //主机ID列表
	AssetGroupIds []string `json:"asset_group_ids"`         //资源组ID
	Exec          string   `json:"exec" binding:"required"` //作业命令
	RunTime       string   `json:"run_time"`                //运行时间
	Cron          bool     `json:"cron"`                    //是否定时
}

type QExec struct {
	JobId   string `form:"job_id"`
	Runtime string `form:"run_time"`
	Status  string `form:"status"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}

type QOverview struct {
	Page    int `form:"page"`
	PerPage int `form:"pre_page"`
}

type JobResults struct {
	JobId   string `form:"job_id" binding:"required"` //作业ID
	HostId  string `form:"host_id"`                   //主机ID
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}

type OverviewDelete struct {
	JobIds []string `json:"job_ids" binding:"required"` //作业ID列表
}

type RFile struct {
	HostIds       []string `json:"host_ids"`                    //主机ID列表
	AssetGroupIds []string `json:"asset_group_ids"`             //资源组ID
	JobId         string   `json:"job_id" binding:"required"`   //作业ID
	DstPath       string   `json:"dst_path" binding:"required"` //目标目录
	SendTime      string   `json:"send_time"`                   //分发时间
	Cron          bool     `json:"cron"`                        //是否定时
}

type QFile struct {
	JobId   string `form:"job_id"`
	Status  string `form:"status"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}

type ScriptDetail struct {
	ScriptId string `form:"script_id" binding:"required"` //作业ID
}

type QScript struct {
	ScriptIds []string `form:"script_ids"`
	Purpose   string   `form:"purpose"`
	NotPage   bool     `form:"not_page"`
	Page      int      `form:"page"`
	PerPage   int      `form:"pre_page"`
}
type QRScript struct {
	JobId    string `form:"job_id"`
	ScriptId string `form:"script_id"`
	Page     int    `form:"page"`
	PerPage  int    `form:"pre_page"`
}
type ScriptDelete struct {
	ScriptIds []string `json:"script_ids" binding:"required"` //脚本ID列表
}
type ScriptModify struct {
	ScriptId   string `json:"script_id" binding:"required"`   //脚本ID
	ScriptDesc string `json:"script_desc" binding:"required"` //脚本描述
}
type RunScript struct {
	AssetGroupIds []string `json:"asset_group_ids"`              //资源组ID
	HostIds       []string `json:"host_ids"`                     //主机ID
	ScriptId      string   `json:"script_id" binding:"required"` //脚本ID
	DstPath       string   `json:"dst_path"`                     //目标路径
	RunTime       string   `json:"run_time"`                     //执行时间
	Cron          bool     `json:"cron"`                         //是否定时
}
