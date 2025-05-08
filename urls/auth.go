package urls

import (
	"github.com/gin-gonic/gin"
	"inner/api/auth"
	"inner/conf/platform_conf"
	"inner/modules/middleware"
)

func AuthGroup(r *gin.Engine) {
	v1 := r.Group("/api/v1/auth")
	{
		v1.POST("/login", auth.Login)
	}
	v2 := r.Group("/api/v1/auth")
	{
		v2.Use(middleware.VerifyToken())
		v2.Use(middleware.VerifyPermission())
		v2.Use(middleware.Audit())
		v2.POST("/logout", auth.Logout)
		v2.POST("/user", auth.AddUser)
		v2.PUT("/user", auth.ModifyUser)
		v2.DELETE("/user", auth.DelUser)
		v2.GET("/users", auth.Users)
		v2.POST("/user/password", auth.ChangePassword)
		v2.POST("/role", auth.AddRole)
		v2.PUT("/role", auth.ModifyRole)
		v2.DELETE("/role", auth.DelRoles)
		v2.GET("/roles", auth.Roles)
		v2.GET("/privileges", auth.Privileges)
		v2.PUT("/privileges", auth.ModifyPrivilege)
		v2.GET("/department", auth.QueryDepartment)
		v2.GET("/tree", auth.QueryTree)
		v2.POST("/department", auth.AddDepartment)
		v2.PUT("/department", auth.ModifyDepartment)
		v2.DELETE("/department", auth.DelDepartment)
		v2.GET("/business", auth.QueryBusiness)
		v2.POST("/business", auth.AddBusiness)
		v2.PUT("/business", auth.ModifyBusiness)
		v2.DELETE("/business", auth.DelBusiness)
		v2.GET("/audit", auth.Audit)
		v2.DELETE("/audit", auth.DeleteAudit)
	}
}

func init() {
	platform_conf.RouteNames["auth.AddUser"] = "新增用户"
	platform_conf.RouteNames["auth.ModifyUser"] = "修改用户"
	platform_conf.RouteNames["auth.DelUser"] = "删除用户"
	platform_conf.RouteNames["auth.Users"] = "查询用户"
	platform_conf.RouteNames["auth.ChangePassword"] = "修改密码"
	platform_conf.RouteNames["auth.AddRole"] = "新增角色"
	platform_conf.RouteNames["auth.ModifyRole"] = "修改角色"
	platform_conf.RouteNames["auth.DelRoles"] = "删除角色"
	platform_conf.RouteNames["auth.Roles"] = "查询角色"
	platform_conf.RouteNames["auth.Privileges"] = "查询权限"
	platform_conf.RouteNames["auth.ModifyPrivilege"] = "修改权限"
	platform_conf.RouteNames["auth.QueryDepartment"] = "查询部门"
	platform_conf.RouteNames["auth.QueryTree"] = "查看服务树"
	platform_conf.RouteNames["auth.AddDepartment"] = "新建部门"
	platform_conf.RouteNames["auth.ModifyDepartment"] = "修改部门"
	platform_conf.RouteNames["auth.DelDepartment"] = "删除部门"
	platform_conf.RouteNames["auth.QueryBusiness"] = "查询业务组"
	platform_conf.RouteNames["auth.AddBusiness"] = "新增业务组"
	platform_conf.RouteNames["auth.DelBusiness"] = "删除业务组"
	platform_conf.RouteNames["auth.ModifyBusiness"] = "修改业务组"
	platform_conf.RouteNames["auth.Audit"] = "查询审计日志"
	platform_conf.RouteNames["auth.DeleteAudit"] = "删除审计日志"
}
