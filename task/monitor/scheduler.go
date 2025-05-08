package monitor

import (
	"github.com/jakecoffman/cron"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	err     error
	Log     kits.Log
	db      = databases.DB
	rc, ctx = common.RedisConnect()
	Cli     = common.ConnInflux()
)

func init() {
	go AlarmEngine()
	go SendEngine()
	go MonitorOverView()
	go CleanServer()
	go MonitorAlive()
	go MonitorHandle()
	go Scheduler()
}

func Scheduler() {
	log := kits.Log{}
	c := cron.New()
	for k, v := range map[string]string{"5m": "0 */5 * * * *", "1h": "0 0 * * * *", "1d": "0 0 0 * * *"} {
		mt := MonitorTrend{k}
		c.AddFunc(v, mt.SyncMonitorTrend, "SyncMonitorTrend_"+k)
	}
	c.AddFunc("0 * * * * *", ServerHealth, "ServerHealth")
	//c.AddFunc("0 */5 * * * *", SyncZabbix, "SyncZabbix")
	c.Start()
	log.Info("monitor scheduler start working ......")
}
