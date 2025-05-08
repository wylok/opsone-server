package k8s

import (
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
