package kits

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"inner/modules/databases"
)

func CasBin() *casbin.Enforcer {
	var Log Log
	db := databases.DB
	m := model.NewModel()
	m.AddDef("r", "r", "sub, dom, obj, act")
	m.AddDef("p", "p", "sub, dom, obj, act")
	m.AddDef("g", "g", "_, _, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch(r.obj, p.obj) &&"+
		" (r.act == p.act || p.act == \"*\" || p.obj==\"*\" || p.sub==\"root\")")
	a, err := gormadapter.NewAdapterByDBUseTableName(db, "", "rules")
	e, err := casbin.NewEnforcer(m, a)
	if e != nil {
		err = e.LoadPolicy()
		e.EnableAutoSave(true)
	}
	if err != nil {
		Log.Error(err)
	}
	return e
}
