package platform

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
	go LocalWscSend()
	go PoolsWscSend()
	go HeartBeatHandle()
	go RsyncAgentConf()
	go Scheduler()
}
func Scheduler() {
	c := cron.New()
	c.AddFunc("*/5 * * * * *", ModifyRemoteAddr, "ModifyRemoteAddr")
	c.Start()
	Log.Info("platform scheduler start working ......")
}
