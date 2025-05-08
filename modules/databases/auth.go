package databases

import (
	"time"
)

type Users struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	UserId       string    `gorm:"column:user_id;type:varchar(100);uniqueIndex" json:"user_id"`
	UserName     string    `gorm:"column:user_name;type:varchar(100);index" json:"user_name"`
	NickName     string    `gorm:"column:nick_name;type:varchar(100);index" json:"nick_name"`
	Email        string    `gorm:"column:email;type:varchar(100)" json:"email"`
	Phone        string    `gorm:"column:phone;type:varchar(100)" json:"phone"`
	Password     string    `gorm:"column:password;type:varchar(100)" json:"password"`
	LastLoginAt  time.Time `gorm:"column:last_login_at;type:datetime" json:"last_login_at"`
	IsRoot       int       `gorm:"column:is_root;type:tinyint(1)" json:"is_root"`
	DepartmentId string    `gorm:"column:department_id;type:varchar(100)" json:"department_id"`
	CreateAt     time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	UpdateAt     time.Time `gorm:"column:update_at;type:datetime" json:"update_at"`
	Status       string    `gorm:"column:status;type:enum('active','close')" json:"status"`
}

func (Users) TableName() string {
	return "users"
}

type Token struct {
	Id       uint64    `gorm:"primary_key" json:"id"`
	UserId   string    `gorm:"column:user_id;type:varchar(100);uniqueIndex" json:"user_id"`
	Token    string    `gorm:"column:token;type:text" json:"token"`
	CreateAt time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	ExpireAt time.Time `gorm:"column:expire_at;type:datetime" json:"expire_at"`
}

func (Token) TableName() string {
	return "token"
}

type Rules struct {
	Id    uint64 `gorm:"primary_key" json:"id"`
	Ptype string `gorm:"column:ptype;type:varchar(100)" json:"ptype"`
	V0    string `gorm:"column:v0;type:varchar(100);uniqueIndex:idx_rules" json:"v0"`
	V1    string `gorm:"column:v1;type:varchar(100);uniqueIndex:idx_rules" json:"v1"`
	V2    string `gorm:"column:v2;type:varchar(100);uniqueIndex:idx_rules" json:"v2"`
	V3    string `gorm:"column:v3;type:varchar(100);uniqueIndex:idx_rules" json:"v3"`
	V4    string `gorm:"column:v4;type:varchar(100)" json:"v4"`
	V5    string `gorm:"column:v5;type:varchar(100)" json:"v5"`
}

func (Rules) TableName() string {
	return "rules"
}

type Roles struct {
	Id       uint64 `gorm:"primary_key" json:"id"`
	RoleId   string `gorm:"column:role_id;type:varchar(100);uniqueIndex" json:"role_id"`
	RoleName string `gorm:"column:role_name;type:varchar(100);index" json:"role_name"`
	RoleType string `gorm:"column:role_type;type:varchar(100)" json:"role_type"`
	RoleDesc string `gorm:"column:role_desc;type:varchar(300)" json:"role_desc"`
}

func (Roles) TableName() string {
	return "roles"
}

type RoleGroup struct {
	Id     uint64 `gorm:"primary_key" json:"id"`
	RoleId string `gorm:"column:role_id;type:varchar(100);uniqueIndex:role_user_id" json:"role_id"`
	UserId string `gorm:"column:user_id;type:varchar(100);uniqueIndex:role_user_id" json:"user_id"`
}

func (RoleGroup) TableName() string {
	return "role_group"
}

type Privileges struct {
	Id            uint64 `gorm:"primary_key" json:"id"`
	PrivilegeId   string `gorm:"column:privilege_id;type:varchar(100);uniqueIndex" json:"privilege_id"`
	PrivilegeName string `gorm:"column:privilege_name;type:varchar(100)" json:"privilege_name"`
	ApiUri        string `gorm:"column:api_uri;type:varchar(100);index" json:"api_uri"`
	ApiMethod     string `gorm:"column:api_method;type:varchar(100)" json:"api_method"`
	VerifyAuth    int    `gorm:"column:verify_auth;type:tinyint(1)" json:"verify_auth"`
	Admin         int    `gorm:"column:admin;type:tinyint(1)" json:"admin"`
	Operator      int    `gorm:"column:operator;type:tinyint(1)" json:"operator"`
	User          int    `gorm:"column:user;type:tinyint(1)" json:"user"`
}

func (Privileges) TableName() string {
	return "privileges"
}

type Permission struct {
	Id          uint64 `gorm:"primary_key" json:"id"`
	RoleId      string `gorm:"column:role_id;type:varchar(100);uniqueIndex:role_privilege_id" json:"role_id"`
	PrivilegeId string `gorm:"column:privilege_id;type:varchar(100);uniqueIndex:role_privilege_id" json:"privilege_id"`
}

func (Permission) TableName() string {
	return "permission"
}

type Department struct {
	Id             uint64    `gorm:"primary_key" json:"id"`
	DepartmentId   string    `gorm:"column:department_id;type:varchar(100);uniqueIndex" json:"department_id"`
	DepartmentName string    `gorm:"column:department_name;type:varchar(100)" json:"department_name"`
	DepartmentDesc string    `gorm:"column:department_desc;type:varchar(100)" json:"department_desc"`
	ParentId       string    `gorm:"column:parent_id;type:varchar(100);index" json:"parent_id"`
	CreateAt       time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	UpdateAt       time.Time `gorm:"column:update_at;type:datetime" json:"update_at"`
}

func (Department) TableName() string {
	return "department"
}

type Business struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	BusinessId   string    `gorm:"column:business_id;type:varchar(100);uniqueIndex" json:"business_id"`
	BusinessName string    `gorm:"column:business_name;type:varchar(100);index" json:"business_name"`
	BusinessDesc string    `gorm:"column:business_desc;type:varchar(100)" json:"business_desc"`
	CreateAt     time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
}

func (Business) TableName() string {
	return "business"
}

type DepartmentBusiness struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	BusinessId   string `gorm:"column:business_id;type:varchar(100);uniqueIndex:business_department_id" json:"business_id"`
	DepartmentId string `gorm:"column:department_id;type:varchar(100);uniqueIndex:business_department_id" json:"department_id"`
}

func (DepartmentBusiness) TableName() string {
	return "department_business"
}

type BusinessUser struct {
	Id         uint64 `gorm:"primary_key" json:"id"`
	BusinessId string `gorm:"column:business_id;type:varchar(100);uniqueIndex:business_user_id" json:"business_id"`
	UserId     string `gorm:"column:user_id;type:varchar(100);uniqueIndex:business_user_id" json:"user_id"`
}

func (BusinessUser) TableName() string {
	return "business_user"
}

type Tenants struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	TenantId   string    `gorm:"column:tenant_id;type:varchar(100);uniqueIndex" json:"tenant_id"`
	TenantName string    `gorm:"column:tenant_name;type:varchar(100)" json:"tenant_name"`
	TenantDesc string    `gorm:"column:tenant_desc;type:varchar(100)" json:"tenant_desc"`
	CreateAt   time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	UpdateAt   time.Time `gorm:"column:update_at;type:datetime" json:"update_at"`
}

func (Tenants) TableName() string {
	return "tenants"
}

type Audit struct {
	Id        uint64    `gorm:"primary_key" json:"id"`
	AuditId   string    `gorm:"column:audit_id;type:varchar(100);uniqueIndex" json:"audit_id"`
	AuditType string    `gorm:"column:audit_type;type:enum('api','ssh','container');index" json:"audit_type"`
	UserId    string    `gorm:"column:user_id;type:varchar(100);index" json:"user_id"`
	Handler   string    `gorm:"column:handler;type:varchar(200);index" json:"handler"`
	Action    string    `gorm:"column:action;type:varchar(500)" json:"action"`
	CreateAt  time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
}

func (Audit) TableName() string {
	return "audit"
}
