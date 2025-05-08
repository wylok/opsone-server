package cloud

import (
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log        kits.Log
	db, sqlErr = databases.MysqlConnect()
)

func init() {
	for _, e := range []error{sqlErr} {
		if e != nil {
			Log.Error(e)
		}
	}
}
