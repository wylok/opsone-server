package monitor

import (
	"fmt"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"reflect"
)

var (
	Log     kits.Log
	db      = databases.DB
	rc, ctx = common.RedisConnect()
	Cli     = common.ConnInflux()
)

func FormatStages(Stages []map[string]int) bool {
	for _, v := range Stages {
		for _, s := range []string{"stage", "interval"} {
			_, ok := v[s]
			if ok {
				if fmt.Sprint(reflect.TypeOf(v[s])) != "int" {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}
