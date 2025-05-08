package auth

import (
	"inner/modules/databases"
	"inner/modules/kits"
)

var (
	Log kits.Log
	err error
	db  = databases.DB
)

func init() {
	go VerifyRolePrivileges()
	go VerifyAuthRules()
	go VerifyDefaultRoles()
}
