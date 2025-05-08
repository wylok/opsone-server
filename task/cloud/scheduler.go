package cloud

import (
	"github.com/jakecoffman/cron"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log     kits.Log
	err     error
	db      = databases.DB
	rc, ctx = common.RedisConnect()
)

func init() {
	go Scheduler()
}
func Scheduler() {
	log := kits.Log{}
	c := cron.New()
	c.AddFunc("0 */5 9-23 * * *", SyncAliYunOss, "SyncAliYunOss")
	c.AddFunc("0 */5 9-23 * * *", SyncAliYunEcs, "SyncAliYunEcs")
	c.AddFunc("0 */5 9-23 * * *", SyncBaiduOss, "SyncBaiduOss")
	c.AddFunc("0 */5 9-23 * * *", SyncBaiduBcc, "SyncBaiduBcc")
	c.AddFunc("0 */5 9-23 * * *", SyncVolcengineEcs, "SyncVolcengineEcs")
	c.Start()
	log.Info("cloud scheduler start working ......")
}
