package auth_conf

import "sync"

var (
	RoleTypes sync.Map
	RoleDesc  = map[string]string{"admin": "平台管理员", "operator": "运维人员", "user": "普通用户"}
)

type LoginConf struct {
	UserName string `json:"user_name" binding:"required"` //用户名
	Password string `json:"password" binding:"required"`  //密码
}
type Users struct {
	UserId   string `form:"user_id"`
	UserName string `form:"user_name"`
	NickName string `form:"nick_name"`
	Email    string `form:"email"`
	Phone    string `form:"phone"`
	Status   string `form:"status"`
	Token    string `form:"token"`
	NotPage  bool   `form:"not_page"`
	Page     int    `form:"page"`
	PerPage  int    `form:"pre_page"`
}

type CreateUser struct {
	UserName     string `json:"user_name" binding:"required"`     //用户名
	NickName     string `json:"nick_name" binding:"required"`     //昵称
	Email        string `json:"email"`                            //电子邮箱
	Phone        string `json:"phone" binding:"required"`         //手机号
	Password     string `json:"password" binding:"required"`      //密码
	DepartmentId string `json:"department_id" binding:"required"` //部门ID
}

type ChangeUser struct {
	UserId       string `json:"user_id" binding:"required"` //用户ID
	UserName     string `json:"user_name"`                  //用户名
	NickName     string `json:"nick_name"`                  //用户昵称
	Email        string `json:"email"`                      //电子邮箱
	Phone        string `json:"phone"`                      //手机号
	DepartmentId string `json:"department_id"`              //部门ID
}
type DeleteUser struct {
	UserIds []string `json:"user_ids" binding:"required"` //用户ID列表
}
type Password struct {
	Password    string `json:"password" binding:"required"`     //原密码
	NewPassword string `json:"new_password" binding:"required"` //新密码
}
type Roles struct {
	RoleId   string `form:"role_id"`
	RoleName string `form:"role_name"`
	UserId   string `form:"user_id"`
	Page     int    `form:"page"`
	PerPage  int    `form:"pre_page"`
}
type Privileges struct {
	PrivilegeId   string `form:"privilege_id"`
	PrivilegeName string `form:"privilege_name"`
	NotPage       bool   `form:"not_page"`
	Page          int    `form:"page"`
	PerPage       int    `form:"pre_page"`
}
type ChangePrivileges struct {
	PrivilegeId string          `json:"privilege_id" binding:"required"` //权限ID
	Action      map[string]bool `json:"action"`                          //授权操作
}
type CreateRole struct {
	RoleName     string   `json:"role_name" binding:"required"` //角色名称
	RoleDesc     string   `json:"role_desc" binding:"required"` //角色描述
	UserIds      []string `json:"user_ids"`                     //用户ID列表
	PrivilegeIds []string `json:"privilege_ids"`                //权限ID列表
}

type ChangeRole struct {
	RoleId       string   `json:"role_id" binding:"required"` //角色ID
	RoleName     string   `json:"role_name"`                  //角色名称
	RoleDesc     string   `json:"role_desc"`                  //角色描述
	UserIds      []string `json:"user_ids"`                   //用户ID列表
	PrivilegeIds []string `json:"privilege_ids"`              //权限ID列表
}
type DeleteRoles struct {
	RoleIds []string `json:"role_ids" binding:"required"` //角色ID
}
type Department struct {
	DepartmentIds  []string `form:"department_ids"`
	DepartmentName string   `form:"department_name"`
	Page           int      `form:"page"`
	PerPage        int      `form:"pre_page"`
}
type CreateDepartment struct {
	DepartmentName string `json:"department_name" binding:"required"` //部门名称
	DepartmentDesc string `json:"department_desc" binding:"required"` //部门描述
	ParentId       string `json:"parent_id"`                          //上级部门ID
}
type ChangeDepartment struct {
	DepartmentId   string `json:"department_id" binding:"required"` //部门ID
	DepartmentName string `json:"department_name"`                  //部门名称
	DepartmentDesc string `json:"department_desc"`                  //部门描述
	ParentId       string `json:"parent_id"`                        //上级部门ID
}
type DeleteDepartment struct {
	DepartmentId string `json:"department_id" binding:"required"` //部门ID
}
type Business struct {
	DepartmentId string `form:"department_id"`
	Page         int    `form:"page"`
	PerPage      int    `form:"pre_page"`
}
type CreateBusiness struct {
	BusinessName string `json:"business_name" binding:"required"` //业务组名称
	BusinessDesc string `json:"business_desc" binding:"required"` //业务组描述
	DepartmentId string `json:"department_id" binding:"required"` //部门ID
}
type ChangeBusiness struct {
	BusinessId   string `json:"business_id" binding:"required"` //业务组ID
	BusinessName string `json:"business_name"`                  //业务组名称
	BusinessDesc string `json:"business_desc"`                  //业务组描述
	DepartmentId string `json:"department_id"`                  //部门ID
}
type DeleteBusiness struct {
	BusinessId string `json:"business_id" binding:"required"` //业务组ID
}
type CreateGroup struct {
	GroupName string `json:"group_name" binding:"required"` //用户组名称
	GroupDesc string `json:"group_desc"`                    //用户组描述
}
type ModifyGroup struct {
	GroupId   string `json:"group_id" binding:"required"` //用户组ID
	GroupName string `json:"group_name"`                  //用户组名称
	GroupDesc string `json:"group_desc"`                  //用户组描述
}
type DeleteGroup struct {
	GroupId string `json:"group_id" binding:"required"` //用户组ID
}
type Groups struct {
	GroupId   string `form:"group_id"`   //用户组ID
	GroupName string `form:"group_name"` //用户组名称
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type Audit struct {
	UserId    string `form:"user_id"`
	UserName  string `form:"user_name"`
	AuditType string `form:"audit_type"`
	Page      int    `form:"page"`
	PerPage   int    `form:"pre_page"`
}
type DeleteAudit struct {
	AuditId string `json:"audit_id" binding:"required"` //审计日志ID
}
