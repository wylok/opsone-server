package platform_conf

import (
	"github.com/duke-git/lancet/random"
	"github.com/gorilla/websocket"
	"sync"
)

const AgentVersion = "2025041401"
const TenantId = "NSJBKI2w3e4rVKTT4r5tyPYG"
const CryptKey = "4098879a2529ca11b8675505ahf88a2d"
const PublicToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MzUyNjQwOTQsImlhdCI6MTYzNTI1Njg5NCwidGVuYW50X2lk" +
	"IjoidGVzdCIsInVzZXJfaWQiOiJOU0pCS0lTVVFWVktUVFdYVVBZRyJ9.UE438SZq8_bOK4truM87yvBRq1vsTFQF7mVU0e0OaPM"
const WscSend = "wsc_send_pool"
const AgentRoot = "/opt/opsone"

var (
	HeartBeatInterval   = 10
	AssetInterval       = 5
	MonitorInterval     = 60
	AssetAgentRun       = 1
	MonitorAgentRun     = 1
	RootPath            string
	Uuid, _             = random.UUIdV4()
	WscPools            = sync.Map{}
	Mch                 = make(chan map[string]interface{}, 100)
	Ech                 = make(chan map[string]interface{}, 100)
	Fch                 = make(chan map[string]interface{}, 100)
	Cch                 = make(chan map[string]interface{}, 100)
	Wch                 = make(chan map[string]interface{}, 100)
	Qch                 = make(chan int, 1)
	Hch                 = make(chan HeartbeatData, 100)
	RemoteAddr          string
	AgentAliveKey       = "agent_alive_key"
	DiscardAssetKey     = "discard_asset_key"
	OfflineAssetKey     = "asset_offline_key"
	DepartmentDeleteKey = "auth_department_delete"
	BusinessDeleteKey   = "auth_business_delete"
	RouterKey           = "platform_route_key"
	RouteNames          = map[string]string{}
	OverViewKey         = "platform_overview_"
	ServerHealthKey     = "platform_server_health_"
	ProcessAliveKey     = "monitor_process_alive_key"
	ServerNameKey       = "cmdb_server_name_key"
	ServerIpKey         = "server_ip_"
	GroupNameKey        = "cmdb_group_name_key"
	ServerSnKey         = "cmdb_server_sn_key"
	DeleteScriptsKey    = "job_delete_script_key"
	HostCpuCoreKey      = "host_cpu_core_key"
	HostWanKey          = "host_wan_"
	UpgradeKey          = "agent_upgrade_lock_"
	IdcIdKey            = "cmdb_idc_id_key"
	IpHostIdKey         = "ip_host_id_key"
	AgentConfMonitor    = "agent_conf_monitor_key"
	AgentConfStatus     = "agent_conf_status_key"
	GroupServersKey     = "cmdb_group_server_key"
	AgentUpgradeKey     = "agent_upgrade_key"
	AgentAliveTraceKey  = "monitor_agent_alive_trace_key"
	AssetPoolIdsKey     = "asset_pool_ids_key"
)

type HeartbeatData struct {
	Ws           *websocket.Conn `json:"ws"`
	HostId       string          `json:"host_id"`
	HostName     string          `json:"host_name"`
	AgentVersion string          `json:"agent_version"`
	ClamAv       string          `json:"clamAv"`
	ClamRun      string          `json:"clamRun"`
}

type OfflineTime struct {
	HostIds []string `form:"host_ids"`
}

type DeepSeek struct {
	Content string `form:"content"`
}

type AgentConf struct {
	Page    int `form:"page"`
	PerPage int `form:"pre_page"`
}
type SshAudit struct {
	AuditId   string `form:"audit_id"`
	AssetType string `form:"asset_type"`
	AssetId   string `form:"asset_id"`
	UserId    string `form:"user_id"`
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type DelSshAudit struct {
	AuditId string `json:"audit_id" binding:"required"`
}
type SshContent struct {
	AuditId string `form:"audit_id" binding:"required"`
}
type AgentAlive struct {
	HostName     string `form:"host_name"`
	AgentVersion string `form:"agent_version"`
	HostId       string `form:"host_id"`
	ClamAv       string `form:"clamAv"`
	Page         int    `form:"page"`
	PerPage      int    `form:"pre_page"`
}
type DeleteAgentAlive struct {
	HostIds []string `json:"host_ids" binding:"required"` //HostIDs
}

type UpdateAgentConf struct {
	Id                int64 `json:"id" binding:"required"` //ID
	AssetAgentRun     bool  `json:"asset_agent_run"`       //是否开启配置采集
	MonitorAgentRun   bool  `json:"monitor_agent_run"`     //是否开启监控采集
	HeartBeatInterval int64 `json:"heartbeat_interval"`    //心跳检测周期
	AssetInterval     int64 `json:"asset_interval"`        //资产上报周期
	MonitorInterval   int64 `json:"monitor_interval"`      //监控上报周期
	Status            bool  `json:"status"`                //开启&关闭Agent
}

type PlatformConfig struct {
	Name    string `form:"name" binding:"required"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}

type Config struct {
	SqlConfig      string                 `yaml:"sql_config" json:"sql_config"`
	InfluxdbConfig map[string]interface{} `yaml:"influxdb_config" json:"influxdb_config"`
	RedisConfig    map[string]interface{} `yaml:"redis_config" json:"redis_config"`
	ApiUrlConfig   string                 `yaml:"api_url_config" json:"api_url_config"`
	JobApiConfig   map[string]interface{} `yaml:"job_api_config" json:"job_api_config"`
	LogPath        string                 `yaml:"log_path" json:"log_path"`
	InfoFile       string                 `yaml:"info_file" json:"info_file"`
	ErrorFile      string                 `yaml:"error_file" json:"error_file"`
	DebugFile      string                 `yaml:"debug_file" json:"debug_file"`
}

func Setting() Config {
	var cf Config
	cf.SqlConfig = "root:Opsone_2024@(opsone-mysql-svc:3306)/<db>?charset=utf8mb4&parseTime=True&loc=Local"
	cf.InfluxdbConfig = map[string]interface{}{"addr": "http://opsone-influxdb-svc:8086", "username": "", "password": ""}
	cf.RedisConfig = map[string]interface{}{"addr": "opsone-redis-svc:6379", "password": ""}
	cf.ApiUrlConfig = "http://opsone-server-svc:8888"
	cf.JobApiConfig = map[string]interface{}{"exec_run": "/api/v1/job/exec", "script_run": "/api/v1/job/script/run"}
	cf.LogPath = "/opt/logs/"
	cf.InfoFile = "opsone.log"
	cf.ErrorFile = "opsone.error"
	cf.DebugFile = "opsone.debug"
	return cf
}
