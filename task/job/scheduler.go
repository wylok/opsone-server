package job

import (
	"github.com/jakecoffman/cron"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log     kits.Log
	db      = databases.DB
	rc, ctx = common.RedisConnect()
)

func init() {
	go JobExecResults()
	go JobFileResults()
	go Scheduler()
}
func Scheduler() {
	log := kits.Log{}
	c := cron.New()
	c.AddFunc("*/30 * * * * *", JobCron, "JobCron")
	c.AddFunc("0 * 9-23 * * *", CleanAsset, "CleanAsset")
	c.AddFunc("0 */5 * * * *", CleanJobs, "CleanJobs")
	c.Start()
	log.Info("job scheduler start working ......")
}
