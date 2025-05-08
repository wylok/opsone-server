package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
)

// @Tags 角色管理
// @Summary 创建角色
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.CreateRole true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/role [post]
func AddRole(c *gin.Context) {
	//创建角色
	var (
		sqlErr     error
		JsonData   = auth_conf.CreateRole{}
		Rules      []databases.Rules
		Roles      []databases.Roles
		Users      []databases.Users
		Privileges []databases.Privileges
		Response   = common.Response{C: c}
		e          = kits.CasBin()
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("role_name=?", JsonData.RoleName).First(&Roles)
		if len(Roles) == 0 {
			RoleId := kits.RandString(12)
			err = db.Transaction(func(tx *gorm.DB) error {
				//新建角色信息
				r := databases.Roles{RoleId: RoleId, RoleName: JsonData.RoleName,
					RoleType: "custom_type", RoleDesc: JsonData.RoleDesc}
				if err = tx.Create(&r).Error; err != nil {
					sqlErr = err
				}
				//新建角色与用户关联信息
				if JsonData.UserIds != nil {
					db.Where("user_id in ?", JsonData.UserIds).Find(&Users)
					if len(Users) != len(JsonData.UserIds) {
						err = errors.New("用户Id列表包含有无效数据")
					} else {
						for _, UserId := range JsonData.UserIds {
							rg := databases.RoleGroup{RoleId: RoleId, UserId: UserId}
							if err = tx.Create(&rg).Error; err != nil {
								sqlErr = err
							}
						}
					}
				}
				//写入角色与权限关联信息
				if JsonData.PrivilegeIds != nil {
					db.Where("privilege_id in ?", JsonData.PrivilegeIds).Find(&Privileges)
					if len(Privileges) != len(JsonData.PrivilegeIds) {
						err = errors.New("权限Id列表包含有无效数据")
					} else {
						for _, PrivilegeId := range JsonData.PrivilegeIds {
							p := databases.Permission{RoleId: RoleId, PrivilegeId: PrivilegeId}
							if err = tx.Create(&p).Error; err != nil {
								sqlErr = err
							}
						}
					}
				}
				if sqlErr == nil {
					//新增用户角色规则
					err = e.LoadPolicy()
					if JsonData.UserIds != nil {
						for _, UserId := range JsonData.UserIds {
							_, err = e.AddGroupingPolicy(UserId, RoleId, platform_conf.TenantId)
							if err != nil {
								break
							}
						}
					}
					//新增角色权限规则
					if JsonData.PrivilegeIds != nil {
						db.Where("privilege_id in ?", JsonData.PrivilegeIds).Find(&Privileges)
						for _, v := range Privileges {
							db.Where("v0=? and v1=? and v2=? and v3=?",
								RoleId, platform_conf.TenantId, v.ApiUri, v.ApiMethod).First(&Rules)
							if len(Rules) == 0 {
								ok, _ := e.AddPolicy(RoleId, platform_conf.TenantId, v.ApiUri, v.ApiMethod)
								if ok {
									err = e.SavePolicy()
								}
							}
						}
					}
				}
				return sqlErr
			})
			if err == nil {
				auth_conf.RoleTypes.Store(RoleId, struct{}{})
			}
		} else {
			err = errors.New("角色名称已存在")
		}
	}
}

// @Tags 角色管理
// @Summary 角色变更
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.ChangeRole true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/role [put]
func ModifyRole(c *gin.Context) {
	//角色变更
	var (
		sqlErr     error
		JsonData   = auth_conf.ChangeRole{}
		Users      []databases.Users
		Roles      []databases.Roles
		RoleGroup  []databases.RoleGroup
		Rules      []databases.Rules
		Privileges []databases.Privileges
		Permission []databases.Permission
		Response   = common.Response{C: c}
		e          = kits.CasBin()
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		//验证角色Id有效性
		db.Where("role_id=?", JsonData.RoleId).First(&Roles)
		_, ok := auth_conf.RoleTypes.Load(JsonData.RoleId)
		if len(Roles) > 0 && ok {
			err = db.Transaction(func(tx *gorm.DB) error {
				upData := databases.Roles{}
				if JsonData.RoleName != "" {
					upData.RoleName = JsonData.RoleName
				}
				if JsonData.RoleDesc != "" {
					upData.RoleDesc = JsonData.RoleDesc
				}
				if JsonData.UserIds != nil {
					db.Where("user_id in ?", JsonData.UserIds).Find(&Users)
					if len(JsonData.UserIds) != len(Users) {
						sqlErr = errors.New("UserId列表包含无效数据")
					} else {
						if err = tx.Where("role_id=?", JsonData.RoleId).Delete(&RoleGroup).Error; err != nil {
							sqlErr = err
						}
						for _, UserId := range JsonData.UserIds {
							rg := databases.RoleGroup{RoleId: JsonData.RoleId, UserId: UserId}
							if err = tx.Create(&rg).Error; err != nil {
								sqlErr = err
							}
						}
						if err = tx.Where("ptype=? and v1=? and v2=?", "g",
							JsonData.RoleId, platform_conf.TenantId).Delete(&Rules).Error; err != nil {
							sqlErr = err
						}
					}
				}
				if JsonData.PrivilegeIds != nil {
					db.Where("privilege_id in ?", JsonData.PrivilegeIds).Find(&Privileges)
					if len(JsonData.PrivilegeIds) != len(Privileges) {
						sqlErr = errors.New("PrivilegeId列表包含无效数据")
					} else {
						if err = tx.Where("role_id=?", JsonData.RoleId).Delete(&Permission).Error; err != nil {
							sqlErr = nil
						}
						for _, v := range JsonData.PrivilegeIds {
							jp := databases.Permission{RoleId: JsonData.RoleId, PrivilegeId: v}
							if err = tx.Create(&jp).Error; err != nil {
								sqlErr = err
							}
						}
						if err = tx.Where("ptype=? and v0=? and v1=?", "p",
							JsonData.RoleId, platform_conf.TenantId).Delete(&Rules).Error; err != nil {
							sqlErr = err
						}
					}
				}
				if err = tx.Model(&Rules).Where("role_id=?", JsonData.RoleId).Updates(upData).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
			if err == nil {
				//新增用户角色规则
				err = e.LoadPolicy()
				if JsonData.UserIds != nil {
					for _, UserId := range JsonData.UserIds {
						_, err = e.AddGroupingPolicy(UserId, JsonData.RoleId, platform_conf.TenantId)
						if err != nil {
							break
						}
					}
				}
				//新增角色权限规则
				if JsonData.PrivilegeIds != nil {
					db.Where("privilege_id in ? and user=?", JsonData.PrivilegeIds, 0).Find(&Privileges)
					for _, v := range Privileges {
						db.Where("v0=? and v1=? and v2=? and v3=?",
							JsonData.RoleId, platform_conf.TenantId, v.ApiUri, v.ApiMethod).First(&Rules)
						if len(Rules) == 0 {
							ok, _ = e.AddPolicy(JsonData.RoleId, platform_conf.TenantId, v.ApiUri, v.ApiMethod)
							if ok {
								err = e.SavePolicy()
							}
						}
					}
				}
			}
		} else {
			err = errors.New("role_id不存在")
		}
	}
}

// @Tags 角色管理
// @Summary 删除角色
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.DeleteRoles true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/role [delete]
func DelRoles(c *gin.Context) {
	//删除角色
	var (
		sqlErr     error
		JsonData   = auth_conf.DeleteRoles{}
		Roles      []databases.Roles
		RoleGroup  []databases.RoleGroup
		Permission []databases.Permission
		Response   = common.Response{C: c}
		e          = kits.CasBin()
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		err = db.Transaction(func(tx *gorm.DB) error {
			if err = tx.Where("role_id in ?", JsonData.RoleIds).Delete(&Roles).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Where("role_id in ?", JsonData.RoleIds).Delete(&RoleGroup).Error; err != nil {
				sqlErr = err
			}
			if err = tx.Where("role_id in ?", JsonData.RoleIds).Delete(&Permission).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
		if err == nil {
			//删除权限规则中的角色信息
			for _, RoleId := range JsonData.RoleIds {
				_, err = e.DeleteRole(RoleId)
				if err != nil {
					break
				}
			}
			if err == nil {
				err = e.SavePolicy()
			}
		}
	}
}

// @Tags 角色管理
// @Summary 角色列表
// @Produce  json
// @Security ApiKeyAuth
// @Param user_id query string false "用户ID"
// @Param role_id query string false "角色ID"
// @Param role_name query string false "角色名"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/roles [get]
func Roles(c *gin.Context) {
	//角色列表
	var (
		RoleIds    []string
		JsonData   = auth_conf.Roles{}
		Roles      []databases.Roles
		Privileges []databases.Privileges
		RoleGroup  []databases.RoleGroup
		Users      []databases.Users
		Data       []map[string]interface{}
		Response   = common.Response{C: c}
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		// 获取角色列表
		tx := db.Order("id")
		if JsonData.RoleId != "" {
			tx = tx.Where("role_id=?", JsonData.RoleId)
		}
		if JsonData.RoleName != "" {
			tx = tx.Where("role_name=?", JsonData.RoleName)
		}
		if JsonData.UserId != "" {
			db.Where("user_id=?", JsonData.UserId).Find(&RoleGroup)
			if len(RoleGroup) > 0 {
				for _, v := range RoleGroup {
					RoleIds = append(RoleIds, v.RoleId)
				}
				tx = tx.Where("role_id in ?", RoleIds)
			}
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&Roles)
		// 附加用户、权限信息
		if len(Roles) > 0 {
			for _, v := range Roles {
				var (
					UserIds []string
					ur      []map[string]interface{}
					pr      []map[string]interface{}
				)
				if JsonData.UserId == "" {
					db.Where("role_id=?", v.RoleId).Find(&RoleGroup)
					for _, d := range RoleGroup {
						UserIds = append(UserIds, d.UserId)
					}
				} else {
					UserIds = append(UserIds, JsonData.UserId)
				}
				db.Where("user_id in ?", UserIds).Find(&Users)
				sql := "join permission on permission.privilege_id=privileges.privilege_id and permission.role_id=?"
				db.Joins(sql, v.RoleId).Where("privileges.verify_auth = ?",
					1).Find(&Privileges)
				for _, v := range Users {
					m, _ := json.Marshal(v)
					ur = append(ur, kits.StringToMap(string(m)))
				}
				for _, v := range Privileges {
					m, _ := json.Marshal(v)
					pr = append(pr, kits.StringToMap(string(m)))
				}
				Data = append(Data, map[string]interface{}{"role": v, "user": ur, "permission": pr})
			}
			Response.Data = Data
		}
	}
}
