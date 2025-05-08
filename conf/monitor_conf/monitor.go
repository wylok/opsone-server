package monitor_conf

import "time"

var (
	Ach                 = make(chan AlarmData, 100)
	Sch                 = make(chan SendMsg, 50)
	Hch                 = make(chan string)
	Pch                 = make(chan string)
	Cch                 = make(chan string)
	AlarmLevels         = map[string]int{"Info": 1, "Warning": 2, "Critical": 3, "Error": 4}
	PauseAlarmKey       = "monitor_pause_alarm_key"
	DefaultRuleContents = map[string]string{"server": "主机检测不可达", "process": "进程运行异常"}
	Measurements        = map[string]string{"5m": "1m", "1h": "5m", "1d": "1h"}
	DurationMeasurement = map[string]string{"5m": "max(*)", "1h": "mean(*)", "1d": "mean(*)"}
	Duration            = []string{"5m", "1h", "1d"}
	DataTrendKey        = "monitor_data_trend_key"
	ProcessTop          = "monitor_process_top"
)

type AlarmData struct {
	HostId          string
	MonitorResource string
	MonitorItem     string
	MonitorKey      string
	MonitorValue    float64
	MonitorInterval int32
}

type SendMsg struct {
	Status          string
	HostId          string
	RuleId          string
	MonitorResource string
	MonitorItem     string
	MonitorValue    float64
	TraceId         string
	MonitorInterval int32
	AlarmContent    string
}

type SystemTags struct {
	HostId string `json:"host_id"`
	Source string `json:"source"`
}

type ProcessTags struct {
	HostId  string `json:"host_id"`
	Process string `json:"process"`
}

type CreateRule struct {
	RuleName        string                   `json:"rule_name" binding:"required"`        //规则名称
	MonitorResource string                   `json:"monitor_resource" binding:"required"` //监控对象
	MonitorItem     string                   `json:"monitor_item" binding:"required"`     //监控项
	MonitorKey      string                   `json:"monitor_key" binding:"required"`      //监控指标
	AlarmLevel      string                   `json:"alarm_level" binding:"required"`      //报警等级
	RuleValue       float64                  `json:"rule_value"`                          //监控阈值
	DiffRule        string                   `json:"diff_rule" binding:"required"`        //报警条件
	RuleT           int                      `json:"rule_t" binding:"required"`           //采集周期
	Stages          []map[string]int         `json:"stages"`                              //发送报警规则
	GroupIds        []string                 `json:"group_ids" binding:"required"`        //绑定资源组id
	Relation        bool                     `json:"relation"`                            //自动关联资源组新增主机
	Channels        []map[string]interface{} `json:"channels"`                            //报警渠道及地址
}

type ModifyRule struct {
	RuleId     string                   `json:"rule_id" binding:"required"` //规则ID
	RuleName   string                   `json:"rule_name"`                  //规则名称
	AlarmLevel string                   `json:"alarm_level"`                //报警等级
	RuleValue  float64                  `json:"rule_value"`                 //监控阈值
	DiffRule   string                   `json:"diff_rule"`                  //报警条件
	Status     string                   `json:"status"`                     //规则状态("active","close")
	RuleT      int                      `json:"rule_t"`                     //采集周期
	GroupIds   []string                 `json:"group_ids"`                  //绑定资源组id
	Channels   []map[string]interface{} `json:"channels"`                   //报警渠道及地址
}

type QueryRule struct {
	UserId          string   `form:"user_id"`          //用户ID
	RuleIds         []string `form:"rule_ids"`         //规则ID
	RuleName        string   `form:"rule_name"`        //规则名称
	MonitorResource string   `form:"monitor_resource"` //监控对象
	MonitorItem     string   `form:"monitor_item"`     //监控项
	MonitorKey      string   `form:"monitor_key"`      //监控指标
	AlarmLevel      string   `form:"alarm_level"`      //报警等级
	Status          string   `form:"status"`           //规则状态("active","close")
	Page            int      `form:"page"`
	PerPage         int      `form:"pre_page"`
}

type DeleteRule struct {
	RuleIds []string `json:"rule_ids" binding:"required"` //规则ID列表
}

type QueryStages struct {
	RuleId string `form:"rule_id" binding:"required"`
}

type RuleGroups struct {
	RuleIds []string `form:"rule_ids" binding:"required"`
}

type RelationRuleGroups struct {
	RuleIds  []string `json:"rule_ids" binding:"required"`  //规则ID列表
	GroupIds []string `json:"group_ids" binding:"required"` //资源组ID列表
}

type CreateStages struct {
	RuleId string           `json:"rule_id" binding:"required"` //规则ID
	Stages []map[string]int `json:"stages" binding:"required"`  //发送报警规则
}

type DataConverge struct {
	HostIds   []string  `form:"host_ids"  binding:"required"`
	Resource  string    `form:"resource"  binding:"required"`
	Item      string    `form:"item"  binding:"required"`
	Converge  string    `form:"converge"  binding:"required"`
	Key       string    `form:"key"`
	Duration  int64     `form:"duration"`
	StartTime time.Time `form:"start_time"`
	EndTime   time.Time `form:"end_time"`
}

type DataDetail struct {
	HostId    string `form:"host_id"  binding:"required"`
	Resource  string `form:"resource"  binding:"required"`
	Item      string `form:"item"  binding:"required"`
	Key       string `form:"key"`
	Duration  int64  `form:"duration"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

type Metric struct {
	Resource string   `form:"resource"  binding:"required"`
	Items    []string `form:"items"  binding:"required"`
}

type AlarmHistory struct {
	RuleId   string `form:"rule_id"`
	HostId   string `form:"host_id"`
	RuleName string `form:"rule_name"`
	Status   string `form:"status"`
	Resource string `form:"resource"`
	Item     string `form:"item"`
	Page     int    `form:"page"`
	PerPage  int    `form:"pre_page"`
}

type AlarmSend struct {
	RuleIds []string `form:"rule_ids"`
	HostIds []string `form:"host_ids"`
	Channel string   `form:"channel"`
	TraceId string   `form:"trace_id"`
	Page    int      `form:"page"`
	PerPage int      `form:"pre_page"`
}

type DeleteAlarm struct {
	Ids     []int64  `json:"ids"`      //ID列表
	RuleIds []string `json:"rule_ids"` //规则ID列表
	HostIds []string `json:"host_ids"` //主机ID列表
}

type PauseAlarm struct {
	TraceId  string `json:"trace_id" binding:"required"` //报警跟踪ID
	Action   string `json:"action"  binding:"required"`  //操作(pause|cancel)
	Duration int64  `json:"duration"`                    //暂停时间(单位:分钟)
}

type QueryPauseAlarm struct {
	TraceIds []string `form:"trace_ids" binding:"required"` //报警跟踪ID列表
}

type QueryProcess struct {
	HostId string `form:"host_id"`
}

type MetricTop struct {
	Item   string `form:"item"  binding:"required"`
	Metric string `form:"metric"  binding:"required"`
}

type DeleteProcess struct {
	HostIds []string `json:"host_ids" binding:"required"` //主机ID列表
	Process []string `json:"process"`                     //进程列表
}

type AddProcess struct {
	HostIds []string `json:"host_ids" binding:"required"` //主机ID列表
	Process []string `json:"process" binding:"required"`  //进程列表
}

type QueryGroupProcess struct {
	GroupId string `form:"group_id"  binding:"required"`
}

type AddGroupProcess struct {
	GroupId string   `json:"group_id" binding:"required"` //资源组ID
	Process []string `json:"process" binding:"required"`  //进程列表
}

type DeleteGroupProcess struct {
	GroupId string   `json:"group_id" binding:"required"` //资源组ID
	Process []string `json:"process"`                     //进程列表
}

type CreateRuleJobs struct {
	RuleId   string `json:"rule_id" binding:"required"` //规则ID
	Exec     string `json:"exec"`                       //执行命令
	ScriptId string `json:"script_id"`                  //监控脚本ID
}

type DeleteRuleJobs struct {
	RuleId string `json:"rule_id" binding:"required"` //规则ID
}

type QueryRuleJobs struct {
	RuleIds []string `form:"rule_ids"` //规则ID列表
}

type CustomMetric struct {
	MonitorKey string `json:"monitor_key" binding:"required"` //监控指标
	KeyCn      string `json:"key_cn" binding:"required"`      //监控指标名称
	KeyUnit    string `json:"key_unit"`                       //监控指标单位
	GroupId    string `json:"group_id" binding:"required"`    //资源组id
	ScriptId   string `json:"script_id" binding:"required"`   //监控脚本ID
}

type DelCustomMetric struct {
	MonitorKey string `json:"monitor_key" binding:"required"` //监控指标
	GroupId    string `json:"group_id" binding:"required"`    //资源组id
}

type QueryCustomMetric struct {
	MonitorKey string `form:"monitor_key" binding:"required"` //监控指标
	GroupId    string `form:"group_id" binding:"required"`    //资源组id
}

type QueryGroupCustom struct {
	GroupId string `form:"group_id" binding:"required"` //资源组id
}

type CustomRefresh struct {
	MonitorKey string `json:"monitor_key" binding:"required"` //监控指标
}

type QueryProcessTop struct {
	HostId string `form:"host_id"  binding:"required"`
}
type RuleContacts struct {
	RuleId  string `form:"rule_id" binding:"required"`
	Channel string `form:"channel" binding:"required"`
}
