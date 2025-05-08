package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags 安全审计
// @Summary 查询日志审计
// @Produce  json
// @Security ApiKeyAuth
// @Param user_id query string false "用户ID"
// @Param user_name query string false "用户名称"
// @Param audit_type query string false "审计类型"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/audit [get]
func Audit(c *gin.Context) {
	//查询日志审计
	var (
		JsonData  auth_conf.Audit
		Audit     []databases.Audit
		Users     []databases.Users
		UserNames = map[string]string{}
		Response  = common.Response{C: c}
		Data      []map[string]interface{}
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Order("create_at desc")
		if JsonData.UserName != "" {
			db.Where("user_name like ?", "%"+JsonData.UserName+"%").Find(&Users)
			if len(Users) > 0 {
				JsonData.UserId = Users[0].UserId
			}
		}
		if JsonData.UserId != "" {
			tx = tx.Where("user_id=?", JsonData.UserId)
		}
		if JsonData.AuditType != "" {
			tx = tx.Where("audit_type=?", JsonData.AuditType)
		}
		db.Where("user_id=? and is_root=?", c.GetString("user_id"), 1).Find(&Users)
		if len(Users) > 0 {
			db.Find(&Users)
			for _, v := range Users {
				UserNames[v.UserId] = v.NickName
			}
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, _ = p.Paging(&Audit)
			for _, v := range Audit {
				handler := v.Handler
				if rc.HExists(ctx, platform_conf.ServerNameKey, v.Handler).Val() {
					handler = rc.HGet(ctx, platform_conf.ServerNameKey, v.Handler).Val()
				}
				Data = append(Data, map[string]interface{}{"audit_id": v.AuditId, "user_id": v.UserId,
					"audit_type": v.AuditType, "user_name": UserNames[v.UserId], "handler": v.Handler,
					"handler_name": handler, "action": v.Action, "create_at": v.CreateAt})
			}
			Response.Data = Data
		} else {
			err = errors.New("该用户无查询权限")
		}
	}
}

// @Tags 安全审计
// @Summary 删除审计日志
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.DeleteAudit true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/audit [delete]
func DeleteAudit(c *gin.Context) {
	//Token验证
	var (
		JsonData auth_conf.DeleteAudit
		Audit    []databases.Audit
		Users    []databases.Users
		Response = common.Response{C: c}
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
		db.Where("user_id=? and is_root=?", c.GetString("user_id"), 1).First(&Users)
		if len(Users) > 0 {
			db.Where("audit_id=?", JsonData.AuditId).Delete(&Audit)
		} else {
			err = errors.New("该用户无删除权限")
		}
	}
}
