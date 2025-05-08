package common

import (
	"gorm.io/gorm"
	"inner/conf/platform_conf"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

var (
	err  error
	Log  kits.Log
	conn = ConnInflux()
	db   = databases.DB
	cf   = platform_conf.Setting()
)

func init() {
	var (
		err            error
		Users          []databases.Users
		Roles          []databases.Roles
		RoleGroup      []databases.RoleGroup
		AgentConf      []databases.AgentConf
		MonitorKeys    []databases.MonitorKeys
		MonitorMetrics []databases.MonitorMetrics
		Rules          []databases.Rules
		e              = kits.CasBin()
		userId         = kits.RandString(0)
		Encrypt        = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		policy         = map[string]string{"opsone_monitor": "90d", "opsone_k8s": "1h"}
	)
	influx := InfluxDb{Cli: conn, Database: ""}
	for _, d := range []string{"opsone_monitor", "opsone_k8s"} {
		err = influx.CreateDatabase(d)
		if err == nil {
			err = influx.CreatePolicy(d, policy[d])
		}
	}
	db.Where("user_name=?", "admin").First(&Users)
	if len(Users) == 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			//写入用户信息
			u := databases.Users{UserId: userId, UserName: "admin",
				NickName: "管理员", Email: "", Phone: "",
				Password: Encrypt.EncryptString("Opsone1234", true), LastLoginAt: time.Now(),
				IsRoot: 1, DepartmentId: "", CreateAt: time.Now(), UpdateAt: time.Now(), Status: "active"}
			err = tx.Create(&u).Error
			if err == nil {
				ok, _ := e.AddGroupingPolicy(userId, "root", platform_conf.TenantId)
				if ok {
					db.Where("ptype=? and v0=? and v1=? and v2=? and v3=?", "p", "root",
						platform_conf.TenantId, "*", "*").First(&Rules)
					if len(Rules) == 0 {
						ok, _ = e.AddPolicy("root", platform_conf.TenantId, "*", "*")
						if ok {
							err = e.SavePolicy()
						}
					}
				}
			}
			return err
		})
	}
	db.First(&Roles)
	if len(Roles) == 0 {
		for _, v := range []interface{}{[]string{"admin", "admin", "default_type", "平台管理员"},
			[]string{"operator", "operator", "default_type", "运维人员"},
			[]string{"user", "user", "default_type", "普通用户"}} {
			rl := databases.Roles{RoleId: v.([]string)[0], RoleName: v.([]string)[1],
				RoleType: v.([]string)[2], RoleDesc: v.([]string)[3]}
			err = db.Create(&rl).Error
		}
	}
	db.First(&RoleGroup)
	if len(RoleGroup) == 0 {
		for _, v := range []interface{}{[]string{"admin", userId}} {
			rg := databases.RoleGroup{RoleId: v.([]string)[0], UserId: v.([]string)[1]}
			err = db.Create(&rg).Error
		}
	}
	db.Find(&AgentConf)
	if len(AgentConf) == 0 {
		ac := databases.AgentConf{AgentVersion: platform_conf.AgentVersion,
			AssetAgentRun: platform_conf.AssetAgentRun, MonitorAgentRun: platform_conf.MonitorAgentRun,
			HeartBeatInterval: int64(platform_conf.HeartBeatInterval), AssetInterval: int64(platform_conf.AssetInterval),
			MonitorInterval: int64(platform_conf.MonitorInterval), Status: 1}
		err = db.Create(&ac).Error
	} else {
		platform_conf.HeartBeatInterval = int(AgentConf[0].HeartBeatInterval)
		platform_conf.AssetInterval = int(AgentConf[0].AssetInterval)
		platform_conf.MonitorInterval = int(AgentConf[0].MonitorInterval)
		platform_conf.AssetAgentRun = AgentConf[0].AssetAgentRun
		platform_conf.MonitorAgentRun = AgentConf[0].MonitorAgentRun
		db.Model(&AgentConf).Where("agent_version=?", AgentConf[0].AgentVersion).Updates(
			databases.AgentConf{AgentVersion: platform_conf.AgentVersion})
	}
	db.First(&MonitorKeys)
	if len(MonitorKeys) == 0 {
		for _, v := range []interface{}{[]string{"cpu_usage", "CPU使用率", "%"}, []string{"cpu_loadavg", "系统load负载", ""},
			[]string{"mem_used", "内存使用量", "GB"}, []string{"mem_pused", "内存使用率", "%"},
			[]string{"disk_usage", "磁盘使用率", "%"}, []string{"disk_read_traffic", "磁盘读流量", "MB/s"},
			[]string{"disk_write_traffic", "磁盘写流量", "MB/s"}, []string{"lan_outtraffic", "内网出带宽", "Mbit/s"},
			[]string{"lan_intraffic", "内网入带宽", "Mbit/s"}, []string{"alive", "Agent运行异常,主机不可达", ""},
			[]string{"tcp_estab", "tcp活动连接数", ""}, []string{"wan_outtraffic", "公网出带宽", "Mbit/s"},
			[]string{"wan_intraffic", "公网入带宽", "Mbit/s"}} {
			mk := databases.MonitorKeys{MonitorKey: v.([]string)[0], MonitorKeyCn: v.([]string)[1],
				MonitorKeyUnit: v.([]string)[2]}
			err = db.Create(&mk).Error
		}
	}
	db.First(&MonitorMetrics)
	if len(MonitorMetrics) == 0 {
		for _, v := range []interface{}{[]string{"server", "system", "cpu_usage"}, []string{"server", "system", "cpu_loadavg"},
			[]string{"server", "system", "mem_used"}, []string{"server", "system", "mem_pused"},
			[]string{"server", "system", "disk_usage"}, []string{"server", "system", "disk_read_traffic"},
			[]string{"server", "system", "disk_write_traffic"}, []string{"server", "system", "lan_outtraffic"},
			[]string{"server", "system", "lan_intraffic"}, []string{"server", "system", "tcp_estab"},
			[]string{"server", "system", "wan_outtraffic"}, []string{"server", "system", "wan_intraffic"},
			[]string{"server", "process", "cpu_usage"}, []string{"server", "process", "mem_pused"},
			[]string{"server", "process", "disk_read_traffic"}, []string{"server", "process", "disk_write_traffic"},
			[]string{"server", "process", "lan_intraffic"}, []string{"server", "process", "lan_outtraffic"}} {
			mm := databases.MonitorMetrics{MonitorResource: v.([]string)[0], MonitorItem: v.([]string)[1],
				MonitorKey: v.([]string)[2]}
			err = db.Create(&mm).Error
		}
	}
}
