package auth

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/validator"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 用户管理
// @Summary 新建用户
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.CreateUser true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/user [post]
func AddUser(c *gin.Context) {
	//新建用户
	var (
		JsonData auth_conf.CreateUser
		Users    []databases.Users
		Encrypt  = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response = common.Response{C: c}
		now      = time.Now()
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
		db.Where("user_name=?", JsonData.UserName).First(&Users)
		if len(Users) == 0 {
			if JsonData.Email != "" {
				if !validator.IsEmail(JsonData.Email) {
					err = errors.New(JsonData.Email + "邮箱格式错误")
				}
				if err == nil {
					db.Where("email=?", JsonData.Email).First(&Users)
					if len(Users) > 0 {
						err = errors.New(JsonData.Email + "邮箱地址已存在")
					}
				}
			}
			if !validator.IsChineseMobile(JsonData.Phone) {
				err = errors.New(JsonData.Phone + "手机号码格式错误")
			}
			if err == nil {
				db.Where("phone=?", JsonData.Phone).First(&Users)
				if len(Users) > 0 {
					err = errors.New(JsonData.Phone + "手机号码已存在")
				}
			}
			if err == nil {
				userId := kits.RandString(0)
				err = db.Transaction(func(tx *gorm.DB) error {
					//写入用户信息
					u := databases.Users{UserId: userId, UserName: JsonData.UserName,
						NickName: JsonData.NickName, Email: JsonData.Email, Phone: cast.ToString(JsonData.Phone),
						Password: Encrypt.EncryptString(JsonData.Password, true), LastLoginAt: now,
						IsRoot: 0, DepartmentId: JsonData.DepartmentId,
						CreateAt: now, UpdateAt: now, Status: "active"}
					err = tx.Create(&u).Error
					return err
				})
				if err == nil {
					//增加权限规则
					rg := databases.RoleGroup{RoleId: "user", UserId: userId}
					db.Create(&rg)
				}
			}
		} else {
			err = errors.New("用户名已存在")
		}
	}
}

// @Tags 用户管理
// @Summary 修改用户
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.ChangeUser true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/user [put]
func ModifyUser(c *gin.Context) {
	//修改用户
	var (
		JsonData auth_conf.ChangeUser
		Users    []databases.Users
		Response = common.Response{C: c}
		now      = time.Now()
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
		db.Where("user_id=?", JsonData.UserId).First(&Users)
		if len(Users) > 0 {
			upData := databases.Users{UpdateAt: now}
			if Users[0].IsRoot == 0 {
				if JsonData.UserName != "" {
					upData.UserName = JsonData.UserName
				}
				if JsonData.NickName != "" {
					upData.NickName = JsonData.NickName
				}
			}
			if JsonData.Email != "" {
				if !validator.IsEmail(JsonData.Email) {
					err = errors.New(JsonData.Email + "邮箱格式错误")
				}
				if err == nil {
					db.Where("email=?", JsonData.Email).First(&Users)
					if len(Users) > 0 {
						err = errors.New(JsonData.Email + "已存在")
					} else {
						upData.Email = JsonData.Email
					}
				}
			}
			if JsonData.Phone != "" {
				if !validator.IsChineseMobile(JsonData.Phone) {
					err = errors.New(JsonData.Phone + "手机号码格式错误")
				}
				if err == nil {
					db.Where("phone=?", JsonData.Phone).First(&Users)
					if len(Users) > 0 {
						err = errors.New(JsonData.Phone + "已存在")
					} else {
						upData.Phone = JsonData.Phone
					}
				}
			}
			if JsonData.DepartmentId != "" {
				upData.DepartmentId = JsonData.DepartmentId
			}
			if err == nil {
				err = db.Model(&Users).Where("user_id=?", JsonData.UserId).Updates(upData).Error
			}
		} else {
			err = errors.New("用户不存在")
		}
	}
}

// @Tags 用户管理
// @Summary 删除用户
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.DeleteUser true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/user [delete]
func DelUser(c *gin.Context) {
	//删除用户
	var (
		sqlErr    error
		JsonData  = auth_conf.DeleteUser{}
		Users     []databases.Users
		Token     []databases.Token
		RoleGroup []databases.RoleGroup
		Response  = common.Response{C: c}
		e         = kits.CasBin()
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
		db.Where("user_id in ? and is_root=?", JsonData.UserIds, 0).Find(&Users)
		if len(Users) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("user_id in ?", JsonData.UserIds).Delete(&Users).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("user_id in ?", JsonData.UserIds).Delete(&Token).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("user_id in ?", JsonData.UserIds).Delete(&RoleGroup).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
			if err == nil {
				//权限规则中用户信息
				for _, UserId := range JsonData.UserIds {
					_, err = e.DeleteUser(UserId)
					if err != nil {
						break
					}
				}
				if err == nil {
					_ = e.SavePolicy()
				}
			}
		}
	}
}

// @Tags 用户管理
// @Summary 修改密码
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.Password true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/user/password [post]
func ChangePassword(c *gin.Context) {
	//修改密码
	var (
		JsonData auth_conf.Password
		Users    []databases.Users
		Encrypt  = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response = common.Response{C: c}
		now      = time.Now()
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
		if err == nil {
			//清空token信息
			for _, ck := range c.Request.Cookies() {
				c.SetCookie(ck.Name, "", -1, "/", "",
					true, false)
				if ck.Name == "token" {
					//token信息写入header
					c.Writer.Header().Add(ck.Name, "")
				}
			}
		} else {
			Response.Err = err
		}
		Response.Send()
	}()
	if err == nil && c.GetString("user_name") != "guest" {
		//验证密码有效性
		db.Where("user_id=? and password=?", c.GetString("user_id"),
			Encrypt.EncryptString(JsonData.Password, true)).Find(&Users)
		if len(Users) > 0 {
			err = db.Model(&Users).Updates(databases.Users{Password: Encrypt.EncryptString(
				JsonData.NewPassword, true), UpdateAt: now}).Error
		} else {
			err = errors.New("原密码错误")
		}
	}
}

// @Tags 用户管理
// @Summary 用户列表
// @Produce  json
// @Security ApiKeyAuth
// @Param user_id query string false "用户ID"
// @Param username query string false "用户名"
// @Param nickname query string false "用户昵称"
// @Param email query string false "电子邮箱"
// @Param phone query string false "手机号码"
// @Param status query string false "用户状态"
// @Param token query string false "token"
// @Param not_page query boolean false "不分页"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/users [get]
func Users(c *gin.Context) {
	//用户列表
	var (
		JsonData auth_conf.Users
		Users    []databases.Users
		Response = common.Response{C: c}
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
		//验证密码有效性
		tx := db.Order("last_login_at desc")
		if JsonData.UserId != "" {
			tx = tx.Where("user_id=?", JsonData.UserId)
		}
		if JsonData.UserName != "" {
			tx = tx.Where("user_name like ?", "%"+JsonData.UserName+"%")
		}
		if JsonData.NickName != "" {
			tx = tx.Where("nick_name like ?", "%"+JsonData.NickName+"%")
		}
		if JsonData.Email != "" {
			tx = tx.Where("email=?", JsonData.Email)
		}
		if JsonData.Phone != "" {
			tx = tx.Where("phone=?", JsonData.Phone)
		}
		if JsonData.Status != "" {
			tx = tx.Where("status=?", JsonData.Status)
		}
		if JsonData.Token != "" {
			tx = tx.Where("user_id=?", c.GetString("user_id"))
		}
		if JsonData.NotPage {
			tx.Find(&Users)
			Response.Data = Users
		} else {
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, Response.Data = p.Paging(&Users)
		}
	}
}
