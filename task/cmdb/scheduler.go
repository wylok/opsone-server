package cmdb

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
	go CleanOwnership()
	go OverViewCmdb()
	go HandleCmdb()
	go Scheduler()
}
func Scheduler() {
	c := cron.New()
	c.AddFunc("@every 60s", DiscardAssets, "DiscardAssets")
	c.AddFunc("0 */5 9-23 * * *", SyncAssets, "SyncAssets")
	c.AddFunc("0 0 9-23/4 * * *", DiscoverSwitch, "DiscoverSwitch")
	c.AddFunc("0 */30 9-23 * * *", DiscoverServer, "DiscoverServer")
	c.AddFunc("@every 3m", CheckSwitch, "CheckSwitch")
	c.Start()
	Log.Info("cmdb scheduler start working ......")
}
