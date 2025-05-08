package databases

import (
	"time"
)

type WorkOrder struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	OrderId      string    `gorm:"column:order_id;type:varchar(100);uniqueIndex" json:"order_id"`
	OrderFlow    string    `gorm:"column:order_flow;type:varchar(100);index" json:"order_flow"`
	OrderTitle   string    `gorm:"column:order_title;type:varchar(200)" json:"order_title"`
	OrderContent string    `gorm:"column:order_content;type:varchar(500)" json:"order_content"`
	Applicant    string    `gorm:"column:applicant;type:varchar(100);index" json:"applicant"`
	DepartmentId string    `gorm:"column:department_id;type:varchar(100);index" json:"department_id"`
	FlowId       uint      `gorm:"column:flow_id;type:int(8)" json:"flow_id"`
	Approve      string    `gorm:"column:approve;type:varchar(100);index" json:"approve"`
	OrderStatus  string    `gorm:"column:order_status;type:varchar(100);index" json:"order_status"`
	CreateAt     time.Time `gorm:"column:create_at;type:datetime" json:"create_at"`
	EndApproveAt time.Time `gorm:"column:end_approve_at;type:datetime" json:"end_approve_at"`
}

func (WorkOrder) TableName() string {
	return "work_order"
}

type WorkOrderApprove struct {
	Id           uint64    `gorm:"primary_key" json:"id"`
	OrderId      string    `gorm:"column:order_id;type:varchar(100);uniqueIndex:order_leader_id" json:"order_id"`
	LeaderId     string    `gorm:"column:leader_id;type:varchar(100);uniqueIndex:order_leader_id" json:"leader_id"`
	DepartmentId string    `gorm:"column:department_id;type:varchar(100)" json:"department_id"`
	FlowId       uint      `gorm:"column:flow_id;type:int(8)" json:"flow_id"`
	Status       string    `gorm:"column:status;type:enum('wait','approved','refuse')" json:"status"`
	ApproveTime  time.Time `gorm:"column:approve_time;type:datetime" json:"approve_time"`
}

func (WorkOrderApprove) TableName() string {
	return "work_order_approve"
}

type WorkOrderFlow struct {
	Id           uint64 `gorm:"primary_key" json:"id"`
	OrderFlow    string `gorm:"column:order_flow;type:varchar(100);uniqueIndex:order_leader_id" json:"order_flow"`
	DepartmentId string `gorm:"column:department_id;type:varchar(100)" json:"department_id"`
	LeaderId     string `gorm:"column:leader_id;type:varchar(100);uniqueIndex:order_leader_id" json:"leader_id"`
	FlowId       uint   `gorm:"column:flow_id;type:int(8)" json:"flow_id"`
}

func (WorkOrderFlow) TableName() string {
	return "work_order_flow"
}

type WorkOrderType struct {
	Id        uint64 `gorm:"primary_key" json:"id"`
	OrderFlow string `gorm:"column:order_flow;type:varchar(100);Index" json:"order_flow"`
	OrderName string `gorm:"column:order_name;type:varchar(100);uniqueIndex" json:"order_name"`
}

func (WorkOrderType) TableName() string {
	return "work_order_type"
}
