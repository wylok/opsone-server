package auth

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strings"
	"time"
)

func VerifyRolePrivileges() {
	lock := common.SyncMutex{LockKey: "verify_role_privileges_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				var (
					sqlErr     error
					Privileges []databases.Privileges
					Permission []databases.Permission
					Rules      []databases.Rules
					Roles      []databases.Roles
					RuleData   = map[string]string{}
					pvs        = map[string]struct{}{}
					rc, ctx    = common.RedisConnect()
				)
				Log.Info("VerifyRolePrivileges start working ......")
				//新增接口权限
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
					lock.UnLock(true)
				}()
				db.Find(&Roles)
				if len(Roles) > 0 {
					for _, v := range Roles {
						auth_conf.RoleTypes.Store(v.RoleId, struct{}{})
					}
				}
				data := rc.HGetAll(ctx, platform_conf.RouterKey).Val()
				if err == nil {
					for k, v := range data {
						n := strings.Split(k, "/")
						PrivilegeId := n[len(n)-1]
						if PrivilegeId != "" {
							pvs[PrivilegeId] = struct{}{}
						}
						if len(strings.Split(v, ":")) >= 2 {
							path := strings.Split(v, ":")[0]
							method := strings.Split(v, ":")[1]
							_, ok := platform_conf.RouteNames[PrivilegeId]
							if ok {
								PrivilegeName := platform_conf.RouteNames[PrivilegeId]
								db.Where("privilege_id = ?", PrivilegeId).First(&Privileges)
								if len(Privileges) == 0 {
									pv := databases.Privileges{PrivilegeId: PrivilegeId, PrivilegeName: PrivilegeName,
										ApiUri: path, ApiMethod: method, VerifyAuth: 1, Admin: 1, Operator: 0, User: 0}
									db.Create(&pv)
									Log.Info("新增权限接口:" + PrivilegeId + " " + path + " " + method)
									if method == "GET" {
										for _, r := range []string{"operator", "user"} {
											jp := databases.Permission{RoleId: r, PrivilegeId: PrivilegeId}
											db.Create(&jp)
										}
									}
									jp := databases.Permission{RoleId: "admin", PrivilegeId: PrivilegeId}
									db.Create(&jp)
								} else {
									db.Model(&databases.Privileges{}).Where("privilege_id=?",
										PrivilegeId).Updates(databases.Privileges{PrivilegeName: PrivilegeName,
										ApiUri: path, ApiMethod: method})
								}
							}
						}
					}
				}
				//验证权限规则一致性
				db.Find(&Privileges)
				if len(Privileges) > 0 && len(pvs) > 0 {
					err = db.Transaction(func(tx *gorm.DB) error {
						for _, v := range Privileges {
							_, ok := pvs[v.PrivilegeId]
							if !ok {
								if err = tx.Where("privilege_id=?", v.PrivilegeId).Delete(&Permission).Error; err != nil {
									sqlErr = err
								}
								if err = tx.Where("v2=? and v3=?", v.ApiUri, v.ApiMethod).Delete(&Rules).Error; err != nil {
									sqlErr = err
								}
								if err = tx.Where("privilege_id=?", v.PrivilegeId).Delete(&Privileges).Error; err != nil {
									sqlErr = err
								}
							}
						}
						return sqlErr
					})
				}
				db.Select("v2", "v3").Distinct("v2", "v3").Where("ptype=?", "p").Find(&Rules)
				if len(Rules) > 0 {
					for _, v := range Rules {
						if v.V2 != "*" && v.V3 != "*" {
							RuleData[v.V2] = v.V3
						}
					}
					if len(RuleData) > 0 {
						for k, v := range RuleData {
							db.Where("api_uri=? and api_method=?", k, v).First(&Privileges)
							if len(Privileges) == 0 {
								db.Where("v2=? and v3=?", k, v).Delete(&Rules)
							}
						}
					}
				}
				//维护内置角色
				for k, v := range auth_conf.RoleDesc {
					db.Where("role_id=? and role_type=?", k, "default_type").First(&Roles)
					if len(Roles) == 0 {
						r := databases.Roles{RoleId: k, RoleName: k, RoleType: "default_type", RoleDesc: v}
						db.Create(&r)
					}
				}
			}()
		}
		time.Sleep(5 * time.Minute)
	}
}

func VerifyAuthRules() {
	lock := common.SyncMutex{LockKey: "auth_verify_auth_rules_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				var (
					SetUserIds []interface{}
					Privileges []databases.Privileges
					Permission []databases.Permission
					RoleGroup  []databases.RoleGroup
					Rules      []databases.Rules
					Users      []databases.Users
					e          = kits.CasBin()
				)
				Log.Info("VerifyAuthRules start working ......")
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
					lock.UnLock(true)
				}()
				//验证管理员权限规则
				db.Where("is_root=?", 1).Find(&Users)
				if len(Users) > 0 {
					for _, v := range Users {
						db.Where("ptype=? and v0=? and v1=? and v2=?", "g", v.UserId, "root", platform_conf.TenantId).First(&Rules)
						if len(Rules) == 0 {
							ok, _ := e.AddGroupingPolicy(v.UserId, "root", platform_conf.TenantId)
							if ok {
								_ = e.SavePolicy()
							}
						}
						db.Where("ptype=? and v0=? and v1=? and v2=? and v3=?", "p", "root",
							platform_conf.TenantId, "*", "*").First(&Rules)
						if len(Rules) == 0 {
							ok, _ := e.AddPolicy("root", platform_conf.TenantId, "*", "*")
							if ok {
								_ = e.SavePolicy()
							}
						}
					}
				}
				//清除无效用户信息
				db.Select("v0").Where("rules.ptype=?", "g").Find(&Rules)
				if len(Rules) > 0 {
					for _, v := range Rules {
						SetUserIds = append(SetUserIds, v.V0)
					}
					scienceClasses := mapset.NewSetFromSlice(SetUserIds)
					db.Select("user_id").Find(&Users)
					if len(Users) > 0 {
						UserIdSet := mapset.NewSet()
						UserIdSet.Add("platform")
						for _, v := range Users {
							UserIdSet.Add(v.UserId)
						}
						//取差集
						DelUserIds := scienceClasses.Difference(UserIdSet).ToSlice()
						if len(DelUserIds) > 0 {
							for _, v := range DelUserIds {
								ok, _ := e.DeleteUser(v.(string))
								if ok {
									_ = e.SavePolicy()
								}
							}
						}
					}
				}
				//验证角色权限
				sql := "join roles on roles.role_id=role_group.role_id and roles.role_type in ?"
				db.Joins(sql, []string{"default_type", "custom_type"}).Find(&RoleGroup)
				if len(RoleGroup) > 0 {
					for _, v := range RoleGroup {
						RoleId := v.RoleId
						UserId := v.UserId
						TenantId := platform_conf.TenantId
						if RoleId == "platform" && UserId == "platform" {
							TenantId = "platform"
						}
						ok, _ := e.AddGroupingPolicy(UserId, RoleId, TenantId)
						if ok {
							_ = e.SavePolicy()
						}
						db.Where("role_id=?", RoleId).Find(&Permission)
						if len(Permission) > 0 {
							for _, v := range Permission {
								db.Where("privilege_id=?", v.PrivilegeId).First(&Privileges)
								if len(Privileges) > 0 {
									db.Where("v0=? and v1=? and v2=? and v3=?",
										RoleId, TenantId, Privileges[0].ApiUri, Privileges[0].ApiMethod).First(&Rules)
									if len(Rules) == 0 {
										ok, _ = e.AddPolicy(RoleId, TenantId, Privileges[0].ApiUri, Privileges[0].ApiMethod)
										if ok {
											_ = e.SavePolicy()
										}
									}
								}
							}
						}
					}
				}
			}()
		}
		time.Sleep(5 * time.Minute)
	}
}

func VerifyDefaultRoles() {
	lock := common.SyncMutex{LockKey: "auth_verify_default_roles_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				var (
					Roles      []databases.Roles
					Permission []databases.Permission
				)
				//维护内置角色权限一致性
				Log.Info("VerifyDefaultRoles start working ......")
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
					lock.UnLock(true)
				}()
				db.Where("role_type=? and role_id!=?", "default_type", "platform").Find(&Roles)
				if len(Roles) > 0 {
					for _, r := range Roles {
						var PrivilegeIds []string
						db.Where("role_id=?", r.RoleId).Find(&Permission)
						if len(Permission) > 0 {
							for _, p := range Permission {
								PrivilegeIds = append(PrivilegeIds, p.PrivilegeId)
								db.Model(&databases.Privileges{}).Where("privilege_id=?", p.PrivilegeId).Updates(
									map[string]interface{}{r.RoleId: 1})
							}
							if len(PrivilegeIds) > 0 {
								db.Model(&databases.Privileges{}).Where("privilege_id not in ?", PrivilegeIds).Updates(
									map[string]interface{}{r.RoleId: 0})
							}
						} else {
							db.Model(&databases.Privileges{}).Updates(map[string]interface{}{r.RoleId: 0})
						}
					}
				}
			}()
		}
		time.Sleep(5 * time.Minute)
	}
}
