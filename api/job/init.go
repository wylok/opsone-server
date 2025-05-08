package job

import (
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log     kits.Log
	db      = databases.DB
	rc, ctx = common.RedisConnect()
)
