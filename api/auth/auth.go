package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"inner/modules/middleware"
	"strings"
	"time"
)

// @Tags 授权管理
// @Summary 登录验证
// @Accept  json
// @Produce  json
// @Param body body  auth_conf.LoginConf true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	//登录验证
	var (
		sqlErr   error
		JsonData auth_conf.LoginConf
		Users    []databases.Users
		Token    []databases.Token
		Roles    []databases.Roles
		RoleList []string
		Encrypt  = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
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
		if Response.Data != nil {
			//token信息写入cookie
			for k, v := range Response.Data.(map[string]string) {
				c.SetCookie(k, v, 7*86400, "/", c.Request.Host,
					true, true)
				if k == "token" || k == "roles" {
					//token信息写入header
					c.Writer.Header().Add(k, v)
				}
			}
		}
		Response.Send()
	}()
	if err == nil {
		//验证用户名密码有效性
		pwd := Encrypt.EncryptString(JsonData.Password, true)
		db.Where("user_name=? and password = ? and status=?", JsonData.UserName, pwd, "active").Find(&Users)
		if len(Users) > 0 {
			userId := Users[0].UserId
			nickName := Users[0].NickName
			DepartmentId := Users[0].DepartmentId
			d, _ := time.ParseDuration("8h")
			//判断token是否已存在且有效
			db.Where("user_id=? and expire_at>=?", userId, time.Now()).Find(&Token)
			if len(Token) > 0 {
				Response.Data, err = middleware.ParseToken(Token[0].Token, platform_conf.CryptKey)
				if err == nil {
					if err = db.Model(&Users).Where("user_name=?", JsonData.UserName).Updates(
						databases.Users{LastLoginAt: time.Now()}).Error; err != nil {
						sqlErr = err
					}
				}
			} else {
				if Users[0].IsRoot == 1 {
					RoleList = append(RoleList, "root")
				}
				sql := "join role_group on role_group.role_id=roles.role_id and role_group.user_id=?"
				db.Joins(sql, userId).Find(&Roles)
				if len(Roles) > 0 {
					for _, v := range Roles {
						RoleList = append(RoleList, v.RoleName)
					}
				}
				t, err := middleware.GenerateToken(userId, strings.Join(RoleList, ","), JsonData.UserName,
					nickName, DepartmentId, platform_conf.CryptKey, d)
				//用户token信息已存在直接修改
				db.Where("user_id=?", userId).Find(&Token)
				if len(Token) > 0 {
					if err = db.Model(&Token).Where("user_id=?", userId).Updates(
						databases.Token{Token: t, CreateAt: time.Now(),
							ExpireAt: time.Now().Add(d)}).Error; err != nil {
						sqlErr = err
					}
				} else {
					//新建用户token信息
					tk := databases.Token{UserId: userId, Token: t, CreateAt: time.Now(),
						ExpireAt: time.Now().Add(d)}
					if err = db.Create(&tk).Error; err != nil {
						sqlErr = err
					}
				}
				//更新用户最后登录时间
				if err = db.Model(&Users).Where("user_id=?", userId).Updates(
					databases.Users{LastLoginAt: time.Now()}).Error; err != nil {
					sqlErr = err
				}
				if sqlErr == nil {
					Response.Data = map[string]string{"user_id": userId, "user_name": JsonData.UserName,
						"nick_name": nickName, "department_id": DepartmentId, "token": t,
						"roles":       strings.Join(RoleList, ","),
						"expire_time": time.Now().Add(d).Format("2006-01-02 15:04:05")}
				}
			}
		} else {
			err = errors.New("账号或者密码错误")
		}
	}
}

// @Tags 授权管理
// @Summary 登录注销
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/logout [post]
func Logout(c *gin.Context) {
	//登录注销
	var (
		Token    []databases.Token
		Response = common.Response{C: c}
	)
	// 接口请求返回
	defer func() {
		//清空token信息
		for _, ck := range c.Request.Cookies() {
			c.SetCookie(ck.Name, "", -1, "/", "",
				true, false)
			if ck.Name == "token" {
				//token信息写入header
				c.Writer.Header().Add(ck.Name, "")
			}
		}
		Response.Send()
	}()
	//判断token是否已存在
	token := c.GetString("token")
	if token != "" {
		db.Where("token=?", token).First(&Token)
		if len(Token) > 0 {
			db.Model(&Token).Updates(databases.Token{ExpireAt: time.Now()})
			rc.Del(ctx, "auth_token_verify_"+kits.MD5(token))
		}
	}
}
