package k8s

import (
	"github.com/jakecoffman/cron"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log     kits.Log
	db      = databases.DB
	Cli     = common.ConnInflux()
	rc, ctx = common.RedisConnect()
)

func init() {
	go Scheduler()
}
func Scheduler() {
	c := cron.New()
	c.AddFunc("0 * * * * *", GetNodeMetric, "GetNodeMetric")
	c.AddFunc("0 * * * * *", GetPodMetric, "GetPodMetric")
	c.AddFunc("0 */5 * * * *", GetOverView, "GetOverView")
	c.Start()
	Log.Info("k8s task start working ......")
}
