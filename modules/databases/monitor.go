package databases

import (
	"time"
)

type MonitorKeys struct {
	Id             uint64 `gorm:"primary_key" json:"id"`
	MonitorKey     string `gorm:"column:monitor_key;type:varchar(100);uniqueIndex" json:"monitor_key"`
	MonitorKeyCn   string `gorm:"column:monitor_key_cn;type:varchar(100)" json:"monitor_key_cn"`
	MonitorKeyUnit string `gorm:"column:monitor_key_unit;type:varchar(100)" json:"monitor_key_unit"`
}

func (MonitorKeys) TableName() string {
	return "monitor_keys"
}

type MonitorMetrics struct {
	Id              uint64 `gorm:"primary_key" json:"id"`
	MonitorResource string `gorm:"column:monitor_resource;type:varchar(100);uniqueIndex:resource_item_key" json:"monitor_resource"`
	MonitorItem     string `gorm:"column:monitor_item;type:varchar(100);uniqueIndex:resource_item_key" json:"monitor_item"`
	MonitorKey      string `gorm:"column:monitor_key;type:varchar(100);uniqueIndex:resource_item_key" json:"monitor_key"`
}

func (MonitorMetrics) TableName() string {
	return "monitor_metrics"
}

type CustomMetrics struct {
	Id         uint64 `gorm:"primary_key" json:"id"`
	MonitorKey string `gorm:"column:monitor_key;type:varchar(100);uniqueIndex:key_item_id" json:"monitor_key"`
	ScriptId   string `gorm:"column:script_id;type:varchar(100);uniqueIndex:key_item_id" json:"script_id"`
}

func (CustomMetrics) TableName() string {
	return "custom_metrics"
}

type MonitorRules struct {
	Id              uint64    `gorm:"primary_key" json:"id"`
	RuleId          string    `gorm:"column:rule_id;type:varchar(100);uniqueIndex" json:"rule_id"`
	RuleName        string    `gorm:"column:rule_name;type:varchar(100);index" json:"rule_name"`
	RuleType        string    `gorm:"column:rule_type;type:varchar(100);index" json:"rule_type"`
	MonitorResource string    `gorm:"column:monitor_resource;type:varchar(100);index:resourceItemKey" json:"monitor_resource"`
	MonitorItem     string    `gorm:"column:monitor_item;type:varchar(100);index:resourceItemKey" json:"monitor_item"`
	AlarmLevel      string    `gorm:"column:alarm_level;type:enum('Info','Warning','Critical','Error')" json:"alarm_level"`
	MonitorKey      string    `gorm:"column:monitor_key;type:varchar(100);index:resourceItemKey" json:"monitor_key"`
	RuleValue       float64   `gorm:"column:rule_value;type:float" json:"rule_value"`
	DiffRule        string    `gorm:"column:diff_rule;type:enum('!=','=','>=','>','<=','<','<>')" json:"diff_rule"`
	RuleT           int       `gorm:"column:rule_t;type:int(8)" json:"rule_t"`
	AlarmContent    string    `gorm:"column:alarm_content;type:varchar(200)" json:"alarm_content"`
	Status          string    `gorm:"column:status;type:enum('active','close');index" json:"status"`
	CreateUser      string    `gorm:"column:create_user;type:varchar(100)" json:"create_user"`
	CreateTime      time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
	UpdateUser      string    `gorm:"column:update_user;type:varchar(100)" json:"update_user"`
	UpdateTime      time.Time `gorm:"column:update_time;type:datetime" json:"update_time"`
	TrendMetric     int       `gorm:"column:trend_metric;type:tinyint(4)" json:"trend_metric"`
	RuleMd5         string    `gorm:"column:rule_md5;type:varchar(100);uniqueIndex" json:"rule_md5"`
}

func (MonitorRules) TableName() string {
	return "monitor_rules"
}

type MonitorProcess struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	HostId     string    `gorm:"column:host_id;type:varchar(100);uniqueIndex:host_process" json:"host_id"`
	Process    string    `gorm:"column:process;type:varchar(100);uniqueIndex:host_process" json:"process"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
	Status     string    `gorm:"column:status;type:enum('active','close')" json:"status"`
}

func (MonitorProcess) TableName() string {
	return "monitor_process"
}

type GroupProcess struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	GroupId    string    `gorm:"column:group_id;type:varchar(100);uniqueIndex:group_process" json:"group_id"`
	Process    string    `gorm:"column:process;type:varchar(100);uniqueIndex:group_process" json:"process"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (GroupProcess) TableName() string {
	return "group_process"
}

type GroupMetrics struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	GroupId    string    `gorm:"column:group_id;type:varchar(100);uniqueIndex:group_key" json:"group_id"`
	MonitorKey string    `gorm:"column:monitor_key;type:varchar(100);uniqueIndex:group_key" json:"monitor_key"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (GroupMetrics) TableName() string {
	return "group_metrics"
}

type AlarmChannel struct {
	Id        uint64 `gorm:"primary_key" json:"id"`
	RuleId    string `gorm:"column:rule_id;type:varchar(100);uniqueIndex" json:"rule_id"`
	Channel   string `gorm:"column:channel;type:varchar(50);index" json:"channel"`
	Address   string `gorm:"column:address;type:varchar(500);index" json:"address"`
	StartTime string `gorm:"column:start_time;type:varchar(50)" json:"start_time"`
	EndTime   string `gorm:"column:end_time;type:varchar(50)" json:"end_time"`
	Status    string `gorm:"column:status;type:enum('active','close')" json:"status"`
}

func (AlarmChannel) TableName() string {
	return "alarm_channel"
}

type AlarmHistory struct {
	Id              uint64    `gorm:"primary_key" json:"id"`
	StartTime       time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
	EndTime         time.Time `gorm:"column:end_time;type:datetime" json:"end_time"`
	MonitorResource string    `gorm:"column:monitor_resource;type:varchar(100);index" json:"monitor_resource"`
	MonitorItem     string    `gorm:"column:monitor_item;type:varchar(100);index" json:"monitor_item"`
	AlarmLevel      string    `gorm:"column:alarm_level;type:enum('Info','Warning','Critical','Error')" json:"alarm_level"`
	Content         string    `gorm:"column:content;type:varchar(100)" json:"content"`
	Duration        int64     `gorm:"column:duration;type:varchar(100)" json:"duration"`
	RuleId          string    `gorm:"column:rule_id;type:varchar(100);index" json:"rule_id"`
	RuleName        string    `gorm:"column:rule_name;type:varchar(100);index" json:"rule_name"`
	RuleType        string    `gorm:"column:rule_type;type:varchar(100)" json:"rule_type"`
	HostId          string    `gorm:"column:host_id;type:varchar(100);index" json:"host_id"`
	Status          string    `gorm:"column:status;type:enum('fault','recovery','unknown')" json:"status"`
	TraceId         string    `gorm:"column:trace_id;type:varchar(100);uniqueIndex" json:"trace_id"`
}

func (AlarmHistory) TableName() string {
	return "alarm_history"
}

type AlarmSend struct {
	Id       uint64    `gorm:"primary_key" json:"id"`
	SendTime time.Time `gorm:"column:send_time;type:datetime" json:"send_time"`
	HostId   string    `gorm:"column:host_id;type:varchar(100);index" json:"host_id"`
	RuleId   string    `gorm:"column:rule_id;type:varchar(100);index" json:"rule_id"`
	Channel  string    `gorm:"column:channel;type:varchar(100);index" json:"channel"`
	Content  string    `gorm:"column:content;type:varchar(500)" json:"content"`
	TraceId  string    `gorm:"column:trace_id;type:varchar(100);index" json:"trace_id"`
	Result   string    `gorm:"column:result;type:enum('fail','success')" json:"result"`
}

func (AlarmSend) TableName() string {
	return "alarm_send"
}

type AlarmStages struct {
	Id     uint64 `gorm:"primary_key" json:"id"`
	RuleId string `gorm:"column:rule_id;type:varchar(100);uniqueIndex" json:"rule_id"`
	Stages string `gorm:"column:stages;type:varchar(500)" json:"stages"`
}

func (AlarmStages) TableName() string {
	return "alarm_stages"
}

type RuleTemplates struct {
	Id              uint64  `gorm:"primary_key" json:"id"`
	TempId          string  `gorm:"column:temp_id;type:varchar(100);uniqueIndex" json:"temp_id"`
	RuleName        string  `gorm:"column:rule_name;type:varchar(100);index" json:"rule_name"`
	MonitorResource string  `gorm:"column:monitor_resource;type:varchar(100);index" json:"monitor_resource"`
	MonitorItem     string  `gorm:"column:monitor_item;type:varchar(100);index" json:"monitor_item"`
	AlarmLevel      string  `gorm:"column:alarm_level;type:enum('Info','Warning','Critical','Error')" json:"alarm_level"`
	MonitorKey      string  `gorm:"column:monitor_key;type:varchar(100);index" json:"monitor_key"`
	RuleValue       float64 `gorm:"column:rule_value;type:float" json:"rule_value"`
	DiffRule        string  `gorm:"column:diff_rule;type:enum('=','>=','>','<=','<','<>')" json:"diff_rule"`
	RuleT           int     `gorm:"column:rule_t;type:int(8)" json:"rule_t"`
	Describe        string  `gorm:"column:describe;type:varchar(200)" json:"describe"`
}

func (RuleTemplates) TableName() string {
	return "rule_templates"
}

type MonitorGroups struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	AssetGroupId string    `gorm:"column:asset_group_id;type:varchar(100);uniqueIndex:group_rule" json:"asset_group_id"`
	RuleId       string    `gorm:"column:rule_id;type:varchar(100);uniqueIndex:group_rule" json:"rule_id"`
	AutoRelation int       `gorm:"column:auto_relation;type:tinyint(1)" json:"auto_relation"`
	CreateAt     time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	Status       string    `gorm:"column:status;type:enum('active','close')" json:"status"`
}

func (MonitorGroups) TableName() string {
	return "monitor_groups"
}

type MonitorJobs struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	RuleId     string    `gorm:"column:rule_id;type:varchar(100);uniqueIndex:rule_script_id" json:"rule_id"`
	Exec       string    `gorm:"column:exec;type:varchar(200)" json:"exec"`
	ScriptId   string    `gorm:"column:script_id;type:varchar(100);uniqueIndex:rule_script_id" json:"script_id"`
	UserId     string    `gorm:"column:user_id;type:varchar(100);index" json:"user_id"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime" json:"update_time"`
}

func (MonitorJobs) TableName() string {
	return "monitor_jobs"
}
