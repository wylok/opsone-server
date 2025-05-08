package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/auth_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags 权限管理
// @Summary 权限列表
// @Produce  json
// @Security ApiKeyAuth
// @Param privilege_id query string false "权限ID"
// @Param privilege_name query string false "权限名称"
// @Param not_page query boolean false "不分页"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/privileges [get]
func Privileges(c *gin.Context) {
	//权限列表
	var (
		JsonData   = auth_conf.Privileges{}
		Privileges []databases.Privileges
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
		tx := db.Where("verify_auth=?", 1)
		if JsonData.PrivilegeId != "" {
			tx = tx.Where("privilege_id like ?", "%"+JsonData.PrivilegeId+"%")
		}
		if JsonData.PrivilegeName != "" {
			tx = tx.Where("privilege_name like ?", "%"+JsonData.PrivilegeName+"%")
		}
		if JsonData.NotPage {
			tx.Find(&Privileges)
			Response.Data = Privileges
		} else {
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, Response.Data = p.Paging(&Privileges)
		}
	}
}

// @Tags 权限管理
// @Summary 权限变更
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.ChangePrivileges true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/privileges [put]
func ModifyPrivilege(c *gin.Context) {
	//权限变更
	var (
		sqlErr     error
		JsonData   = auth_conf.ChangePrivileges{}
		Permission []databases.Permission
		Response   = common.Response{C: c}
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
		updates := map[string]interface{}{}
		err = db.Transaction(func(tx *gorm.DB) error {
			if JsonData.Action != nil {
				for k, v := range JsonData.Action {
					if err = tx.Model(&databases.Privileges{}).Where("privilege_id=?",
						JsonData.PrivilegeId).Updates(map[string]interface{}{k: cast.ToInt(v)}).Error; err != nil {
						sqlErr = err
					}
					db.Where("role_id=? and privilege_id=?", k, JsonData.PrivilegeId).First(&Permission)
					if v {
						if len(Permission) == 0 {
							ps := databases.Permission{RoleId: k, PrivilegeId: JsonData.PrivilegeId}
							if err = tx.Create(&ps).Error; err != nil {
								sqlErr = err
							}
						}
					} else {
						if len(Permission) > 0 {
							if err = tx.Where("role_id=? and privilege_id=?", k, JsonData.PrivilegeId).Delete(
								&Permission).Error; err != nil {
								sqlErr = err
							}
						}
					}
				}
			}
			if err = tx.Model(&databases.Privileges{}).Where("privilege_id=?",
				JsonData.PrivilegeId).Updates(updates).Error; err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
}
