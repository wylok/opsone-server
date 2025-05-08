package databases

import "time"

type AgentConf struct {
	Id                uint64 `gorm:"primary_key" json:"id"`
	AgentVersion      string `gorm:"column:agent_version;type:varchar(100)" json:"agent_version"`
	AssetAgentRun     int    `gorm:"column:asset_agent_run;type:tinyint(1)" json:"asset_agent_run"`
	MonitorAgentRun   int    `gorm:"column:monitor_agent_run;type:tinyint(1)" json:"monitor_agent_run"`
	HeartBeatInterval int64  `gorm:"column:heartbeat_interval;type:int(8)" json:"heartbeat_interval"`
	AssetInterval     int64  `gorm:"column:asset_interval;type:int(8)" json:"asset_interval"`
	MonitorInterval   int64  `gorm:"column:monitor_interval;type:int(8)" json:"monitor_interval"`
	Status            int    `gorm:"column:status;type:tinyint(1)" json:"status"`
}

func (AgentConf) TableName() string {
	return "agent_conf"
}

type AgentAlive struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	HostId       string `gorm:"column:host_id;type:varchar(100);uniqueIndex" json:"host_id"`
	AgentVersion string `gorm:"column:agent_version;type:varchar(50)" json:"agent_version"`
	ClamAv       string `gorm:"column:clamAv;type:varchar(50)" json:"clamAv"`
	ClamRun      string `gorm:"column:clamRun;type:varchar(50)" json:"clamRun"`
	OfflineTime  int64  `gorm:"column:offline_time;type:int(11)" json:"offline_time"`
}

func (AgentAlive) TableName() string {
	return "agent_alive"
}

type SshAudit struct {
	Id        uint64    `gorm:"primary_key" json:"id"`
	AuditId   string    `gorm:"column:audit_id;type:varchar(100);uniqueIndex" json:"audit_id"`
	AssetId   string    `gorm:"column:asset_id;type:varchar(100);index" json:"asset_id"`
	AssetType string    `gorm:"column:asset_type;type:varchar(100);index" json:"asset_type"`
	UserId    string    `gorm:"column:user_id;type:varchar(100);index" json:"user_id"`
	FileName  string    `gorm:"column:file_name;type:varchar(100)" json:"file_name"`
	StartTime time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
}

func (SshAudit) TableName() string {
	return "ssh_audit"
}

type SshContent struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	AuditId      string `gorm:"column:audit_id;type:varchar(100);uniqueIndex" json:"audit_id"`
	ShellContent string `gorm:"column:shell_content;type:text" json:"shell_content"`
}

func (SshContent) TableName() string {
	return "ssh_content"
}
