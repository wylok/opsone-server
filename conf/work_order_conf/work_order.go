package work_order_conf

type QueryWorkOrder struct {
	OrderTitle string `form:"order_title"` //工单名称
	Page       int    `form:"page"`
	PerPage    int    `form:"pre_page"`
}
type QueryApproveWorkOrder struct {
	Page    int `form:"page"`
	PerPage int `form:"pre_page"`
}
type QueryApproveFlow struct {
	OrderId string `json:"order_id" binding:"required"` //工单ID
}
type CreateWorkOrder struct {
	OrderTitle   string `json:"order_title" binding:"required"`   //工单名称
	OrderContent string `json:"order_content" binding:"required"` //工单内容
	OrderFlow    string `json:"order_flow" binding:"required"`    //工单类型
}
type ChangeWorkOrder struct {
	OrderId      string `json:"order_id" binding:"required"` //工单ID
	OrderTitle   string `json:"order_title"`                 //工单名称
	OrderContent string `json:"order_content"`               //工单内容
}
type ApproveWorkOrder struct {
	OrderId   string `json:"order_id" binding:"required"`   //工单ID
	OrderFlow string `json:"order_flow" binding:"required"` //工单类型
	FlowId    uint   `json:"flow_id" binding:"required"`    //流程ID
	Approve   bool   `json:"approve"  binding:"required"`   //是否审批通过
}
type DeleteWorkOrder struct {
	OrderIds []string `json:"order_ids" binding:"required"` //工单ID列表
}
type AddWorkOrderFlow struct {
	OrderName   string                   `json:"order_name" binding:"required"`    //工单流程名称
	ApproveFlow []map[string]interface{} `json:"approve_flow"  binding:"required"` //审批流[{department_id:string,leader_id:string,flow_id:int}]
}
type ModifyWorkOrderFlow struct {
	OrderFlow   string                   `json:"order_flow" binding:"required"`    //工单流程类型
	OrderName   string                   `json:"order_name"`                       //工单流程名称
	ApproveFlow []map[string]interface{} `json:"approve_flow"  binding:"required"` //审批流[{department_id:string,leader_id:string,flow_id:int}]
}
type DelWorkOrderFlow struct {
	OrderFlow string `json:"order_flow" binding:"required"` //工单流程类型
}
type QueryWorkOrderFlow struct {
	OrderFlow string `json:"order_flow"` //工单流程类型
}
