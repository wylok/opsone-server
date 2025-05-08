package work_order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/work_order_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

var (
	Log kits.Log
	db  = databases.DB
)

// @Tags 工单管理
// @Summary 工单查询
// @Produce  json
// @Security ApiKeyAuth
// @Param order_title query string false "工单名称"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order [get]
func QueryWorkOrder(c *gin.Context) {
	//工单查询
	var (
		JsonData  = work_order_conf.QueryWorkOrder{}
		WorkOrder []databases.WorkOrder
		Response  = common.Response{C: c}
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
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
		tx := db.Where("order_status != ? and applicant=?", "delete", c.GetString("user_id"))
		if JsonData.OrderTitle != "" {
			tx = tx.Where("order_title like ?", "%"+JsonData.OrderTitle+"%")
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&WorkOrder)
	}
}

// @Tags 工单管理
// @Summary 已审批列表查询
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order/approve/ready [get]
func ReadyApproveWorkOrder(c *gin.Context) {
	//已审批列表查询
	var (
		JsonData         = work_order_conf.QueryApproveWorkOrder{}
		Users            []databases.Users
		WorkOrder        []databases.WorkOrder
		WorkOrderApprove []databases.WorkOrderApprove
		Response         = common.Response{C: c}
		UserId           = c.GetString("user_id")
		OrderIds         []string
		data             []map[string]interface{}
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
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
		db.Where("user_id=? and user_name=?", UserId, "admin").First(&Users)
		if len(Users) > 0 {
			db.Where("status!=?", "wait").Find(&WorkOrderApprove)
		} else {
			db.Where("leader_id=? and status!=?", UserId, "wait").Find(&WorkOrderApprove)
		}
		if len(WorkOrderApprove) > 0 {
			for _, v := range WorkOrderApprove {
				OrderIds = append(OrderIds, v.OrderId)
			}
		}
		if OrderIds != nil {
			tx := db.Order("end_approve_at desc").Where("order_id in ?", OrderIds)
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, _ = p.Paging(&WorkOrder)
			for _, v := range WorkOrder {
				if len(Users) > 0 {
					db.Where("order_id=? and status!=?", v.OrderId, "wait").First(&WorkOrderApprove)
				} else {
					db.Where("leader_id=? and order_id=? and status!=?", UserId, v.OrderId, "wait").First(&WorkOrderApprove)
				}
				data = append(data, map[string]interface{}{"order_id": v.OrderId, "order_title": v.OrderTitle,
					"order_content": v.OrderContent, "applicant": v.Applicant, "department_id": v.DepartmentId,
					"order_operate": WorkOrderApprove[0].Status, "approve_time": WorkOrderApprove[0].ApproveTime})
			}
			Response.Data = data
		}
	}
}

// @Tags 工单管理
// @Summary 待审批列表查询
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order/approve/pend [get]
func PendApproveWorkOrder(c *gin.Context) {
	//待审批列表查询
	var (
		JsonData         = work_order_conf.QueryApproveWorkOrder{}
		Users            []databases.Users
		WorkOrder        []databases.WorkOrder
		WorkOrderApprove []databases.WorkOrderApprove
		Response         = common.Response{C: c}
		UserId           = c.GetString("user_id")
		OrderIds         []string
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
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
		db.Where("user_id=? and user_name=?", UserId, "admin").First(&Users)
		if len(Users) > 0 {
			db.Select("order_id").Where("status=?", "wait").Find(&WorkOrderApprove)
		} else {
			db.Select("order_id").Where("leader_id=? and status=?", UserId, "wait").Find(&WorkOrderApprove)
		}
		if len(WorkOrderApprove) > 0 {
			for _, v := range WorkOrderApprove {
				OrderIds = append(OrderIds, v.OrderId)
			}
		}
		tx := db.Order("create_at desc").Where("order_id in ?", OrderIds)
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&WorkOrder)
	}
}

// @Tags 工单管理
// @Summary 审批进度查询
// @Produce  json
// @Security ApiKeyAuth
// @Param order_id query string true "工单ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order/approve/flow [get]
func QueryApproveFlow(c *gin.Context) {
	//审批进度查询
	var (
		JsonData         = work_order_conf.QueryApproveFlow{}
		WorkOrderApprove []databases.WorkOrderApprove
		WorkOrderFlow    []databases.WorkOrderFlow
		WorkOrder        []databases.WorkOrder
		Response         = common.Response{C: c}
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Order("flow_id desc").Where("order_id=?", JsonData.OrderId).Find(&WorkOrderApprove)
		if len(WorkOrderApprove) == 0 {
			db.Where("order_id=?", JsonData.OrderId).Find(&WorkOrder)
			if len(WorkOrder) > 0 {
				db.Where("order_flow=? and flow_id=?", WorkOrder[0].OrderFlow, 1).First(&WorkOrderFlow)
				if len(WorkOrderFlow) > 0 {
					Response.Data = map[string]interface{}{"order_id": JsonData.OrderId,
						"leader_id": WorkOrderFlow[0].LeaderId, "department_id": WorkOrderFlow[0].DepartmentId,
						"flow_id": WorkOrderFlow[0].FlowId, "status": "wait"}
				}
			}
		} else {
			Response.Data = WorkOrderApprove
		}
	}
}

// @Tags 工单管理
// @Summary 新建工单
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.CreateWorkOrder true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order [post]
func AddWorkOrder(c *gin.Context) {
	//新建工单
	var (
		sqlErr        error
		JsonData      = work_order_conf.CreateWorkOrder{}
		WorkOrderFlow []databases.WorkOrderFlow
		Users         []databases.Users
		Response      = common.Response{C: c}
		UserId        = c.GetString("user_id")
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		orderId := kits.RandString(8)
		err = db.Transaction(func(tx *gorm.DB) error {
			//获取部门信息
			db.Where("user_id=?", UserId).First(&Users)
			if len(Users) > 0 {
				//写入工单信息
				db.Order("flow_id desc").Where("order_flow=?", JsonData.OrderFlow).Find(&WorkOrderFlow)
				if len(WorkOrderFlow) > 0 {
					var Approve string
					err = db.Transaction(func(tx *gorm.DB) error {
						for _, v := range WorkOrderFlow {
							if v.FlowId == 1 {
								Approve = v.LeaderId
							}
							wa := databases.WorkOrderApprove{OrderId: orderId, LeaderId: v.LeaderId, DepartmentId: v.DepartmentId,
								FlowId: v.FlowId, Status: "wait", ApproveTime: time.Now()}
							if err = tx.Create(&wa).Error; err != nil {
								sqlErr = err
							}
						}
						wo := databases.WorkOrder{OrderId: orderId, OrderFlow: JsonData.OrderFlow, OrderTitle: JsonData.OrderTitle,
							OrderContent: JsonData.OrderContent, Applicant: UserId, DepartmentId: Users[0].DepartmentId, FlowId: 1,
							Approve: Approve, OrderStatus: "submitted", CreateAt: time.Now(), EndApproveAt: time.Now()}
						if err = tx.Create(&wo).Error; err != nil {
							sqlErr = err
						}
						return sqlErr
					})
				} else {
					err = errors.New("无效的工单类型")
				}
			} else {
				err = errors.New("无效的工单类型")
			}
			return err
		})
	}
}

// @Tags 工单管理
// @Summary 工单变更
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.ChangeWorkOrder true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order [put]
func ModifyWorkOrder(c *gin.Context) {
	//工单变更
	var (
		sqlErr    error
		JsonData  = work_order_conf.ChangeWorkOrder{}
		WorkOrder []databases.WorkOrder
		Response  = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("order_id=?", JsonData.OrderId).First(&WorkOrder)
		if len(WorkOrder) > 0 {
			db.Where("order_id=? and order_status=?", JsonData.OrderId, "submitted").First(&WorkOrder)
			if len(WorkOrder) > 0 {
				err = db.Transaction(func(tx *gorm.DB) error {
					upData := databases.WorkOrder{}
					if JsonData.OrderTitle != "" {
						upData.OrderTitle = JsonData.OrderTitle
					}
					if JsonData.OrderContent != "" {
						upData.OrderContent = JsonData.OrderContent
					}
					if err = tx.Model(&WorkOrder).Where("order_id=?", JsonData.OrderId).Updates(
						upData).Error; err != nil {
						sqlErr = err
					}
					return sqlErr
				})
			} else {
				err = errors.New("工单审批流程已启动,不可修改工单")
			}
		} else {
			err = errors.New("工单(" + JsonData.OrderId + ")不存在")
		}
	}
}

// @Tags 工单管理
// @Summary 注销工单
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.DeleteWorkOrder true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order [delete]
func DelWorkOrder(c *gin.Context) {
	//注销工单
	var (
		sqlErr           error
		JsonData         = work_order_conf.DeleteWorkOrder{}
		WorkOrder        []databases.WorkOrder
		WorkOrderApprove []databases.WorkOrderApprove
		Response         = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("order_id in ?", JsonData.OrderIds).Find(&WorkOrder)
		if len(WorkOrder) == len(JsonData.OrderIds) {
			err = db.Transaction(func(tx *gorm.DB) error {
				for _, orderId := range JsonData.OrderIds {
					if err = tx.Where("order_id = ?", orderId).Delete(&WorkOrder).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("order_id = ?", orderId).Delete(&WorkOrderApprove).Error; err != nil {
						sqlErr = err
					}
				}
				return sqlErr
			})
		} else {
			err = errors.New("包含无效工单ID")
		}
	}
}

// @Tags 工单管理
// @Summary 工单审批
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.ApproveWorkOrder true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order/approve [post]
func ApproveWorkOrder(c *gin.Context) {
	//工单审批
	var (
		JsonData         = work_order_conf.ApproveWorkOrder{}
		Users            []databases.Users
		WorkOrder        []databases.WorkOrder
		WorkOrderApprove []databases.WorkOrderApprove
		Response         = common.Response{C: c}
		UserId           = c.GetString("user_id")
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("order_id=? and order_status in ?", JsonData.OrderId, []string{"submitted", "approve"}).Find(&WorkOrder)
		db.Where("user_id=? and user_name=?", UserId, "admin").Find(&Users)
		db.Where("order_id=? and leader_id=?", JsonData.OrderId, UserId).Find(&WorkOrderApprove)
		if len(WorkOrder) > 0 {
			if len(Users) > 0 || len(WorkOrderApprove) > 0 {
				st1 := "approve"
				st2 := "approved"
				if !JsonData.Approve {
					st1 = "refuse"
					st2 = "refuse"
				}
				db.Where("order_id=? and flow_id=?", JsonData.OrderId, JsonData.FlowId).Find(&WorkOrderApprove)
				if len(WorkOrderApprove) > 0 {

					db.Model(&WorkOrderApprove).Where("order_id=? and flow_id=?",
						JsonData.OrderId, JsonData.FlowId).Updates(
						databases.WorkOrderApprove{Status: st2, ApproveTime: time.Now()})
					db.Where("order_id=?", JsonData.OrderId).Find(&WorkOrderApprove)
					if JsonData.FlowId >= uint(len(WorkOrderApprove)) {
						st1 = "approved"
					} else {
						JsonData.FlowId++
					}
					db.Model(&WorkOrder).Where("order_id=?", JsonData.OrderId).Updates(
						databases.WorkOrder{OrderStatus: st1, Approve: UserId, FlowId: JsonData.FlowId,
							EndApproveAt: time.Now()})
				}
			} else {
				err = errors.New("工单(" + JsonData.OrderId + ")没有审批权限")
			}
		} else {
			err = errors.New("工单(" + JsonData.OrderId + ")已审批结束")
		}
	}
}

// @Tags 工单管理
// @Summary 创建工单流程
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.AddWorkOrderFlow true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order/flow [post]
func AddWorkOrderFLow(c *gin.Context) {
	//创建工单流程
	var (
		sqlErr        error
		JsonData      = work_order_conf.AddWorkOrderFlow{}
		WorkOrderType []databases.WorkOrderType
		Response      = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("order_name=?", JsonData.OrderName).First(&WorkOrderType)
		if len(WorkOrderType) == 0 {
			orderFlow := kits.RandString(8)
			err = db.Transaction(func(tx *gorm.DB) error {
				//创建工单类型
				wot := databases.WorkOrderType{OrderFlow: orderFlow, OrderName: JsonData.OrderName}
				if err = tx.Create(&wot).Error; err != nil {
					sqlErr = err
				}
				//创建工单流
				for _, data := range JsonData.ApproveFlow {
					wof := databases.WorkOrderFlow{OrderFlow: orderFlow,
						DepartmentId: data["department_id"].(string), LeaderId: data["leader_id"].(string),
						FlowId: uint(data["flow_id"].(float64))}
					if err = tx.Create(&wof).Error; err != nil {
						sqlErr = err
					}
				}
				return sqlErr
			})
		} else {
			err = errors.New("工单流程(" + JsonData.OrderName + ")已存在")
		}
	}
}

// @Tags 工单管理
// @Summary 修改工单流程
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.ModifyWorkOrderFlow true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order/flow [put]
func ModifyWorkOrderFLow(c *gin.Context) {
	//修改工单流程
	var (
		sqlErr        error
		JsonData      = work_order_conf.ModifyWorkOrderFlow{}
		WorkOrderFlow []databases.WorkOrderFlow
		WorkOrderType []databases.WorkOrderType
		Response      = common.Response{C: c}
		FlowIds       []uint
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Select("flow_id").Where("order_flow=?", JsonData.OrderFlow).Find(&WorkOrderFlow)
		if len(WorkOrderFlow) > 0 {
			for _, v := range WorkOrderFlow {
				FlowIds = append(FlowIds, v.FlowId)
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				//修改工单名称
				if JsonData.OrderName != "" {
					if err = tx.Model(&WorkOrderType).Where("order_flow=?", JsonData.OrderFlow).Updates(
						databases.WorkOrderType{OrderName: JsonData.OrderName}).Error; err != nil {
						sqlErr = err
					}
				}
				for _, data := range JsonData.ApproveFlow {
					//判断工单流是否存在
					db.Where("order_flow=? and flow_id=?", JsonData.OrderFlow, uint(data["flow_id"].(float64))).First(&WorkOrderFlow)
					if len(WorkOrderFlow) > 0 {
						//存在即修改工单流
						tx.Model(&WorkOrderFlow).Where("order_flow=? and flow_id=?", JsonData.OrderFlow, uint(data["flow_id"].(float64))).Updates(
							databases.WorkOrderFlow{DepartmentId: data["department_id"].(string), LeaderId: data["leader_id"].(string)})
					} else {
						//不存在即新建工单流
						wo := databases.WorkOrderFlow{OrderFlow: JsonData.OrderFlow,
							DepartmentId: data["department_id"].(string), LeaderId: data["leader_id"].(string),
							FlowId: uint(data["flow_id"].(float64))}
						if err = tx.Create(&wo).Error; err != nil {
							sqlErr = err
						}
					}
				}
				for _, v := range FlowIds {
					del := true
					for _, f := range JsonData.ApproveFlow {
						if v == uint(f["flow_id"].(float64)) {
							del = false
						}
					}
					//删除不存在的工单流
					if del {
						if err = tx.Where("order_flow=? and flow_id=?", JsonData.OrderFlow, v).Delete(&WorkOrderFlow).Error; err != nil {
							sqlErr = err
						}
					}
				}
				return sqlErr
			})
		} else {
			err = errors.New("工单类型(" + JsonData.OrderFlow + ")不存在")
		}
	}
}

// @Tags 工单管理
// @Summary 删除工单流程
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  body work_order_conf.DelWorkOrderFlow true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/work_order/flow [delete]
func DelWorkOrderFLow(c *gin.Context) {
	//删除工单流程
	var (
		sqlErr        error
		JsonData      = work_order_conf.DelWorkOrderFlow{}
		WorkOrderFlow []databases.WorkOrderFlow
		WorkOrder     []databases.WorkOrder
		WorkOrderType []databases.WorkOrderType
		Response      = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			Log.Error(sqlErr)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		sql := "join work_order on work_order.order_flow=work_order_flow.order_flow and work_order.order_status!=?"
		db.Joins(sql, "close").Where("order_flow=?", JsonData.OrderFlow).First(&WorkOrderFlow)
		if len(WorkOrderFlow) == 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("order_flow=?", JsonData.OrderFlow).Delete(&WorkOrderFlow).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Model(&WorkOrder).Where("order_flow=?", JsonData.OrderFlow).Updates(
					databases.WorkOrder{OrderStatus: "delete"}).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("order_flow=?", JsonData.OrderFlow).Delete(&WorkOrderType).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New("该工单类型(" + JsonData.OrderFlow + ")有审批任务，无法删除")
		}
	}
}

// @Tags 工单管理
// @Summary 查询工单流程
// @Produce  json
// @Security ApiKeyAuth
// @Param order_type query string false "工单类型"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order/flow [get]
func QueryWorkOrderFlow(c *gin.Context) {
	//查询工单流程
	var (
		JsonData      = work_order_conf.QueryWorkOrderFlow{}
		WorkOrderType []databases.WorkOrderType
		WorkOrderFlow []databases.WorkOrderFlow
		Response      = common.Response{C: c}
		data          = map[string]interface{}{}
	)
	err := c.BindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.OrderFlow != "" {
			db.Where("order_flow=?", JsonData.OrderFlow).First(&WorkOrderType)
		} else {
			db.Find(&WorkOrderType)
		}
		if len(WorkOrderType) > 0 {
			for _, v := range WorkOrderType {
				var Flows []map[string]interface{}
				db.Where("order_flow=?", v.OrderFlow).Find(&WorkOrderFlow)
				if len(WorkOrderFlow) > 0 {
					for _, d := range WorkOrderFlow {
						Flows = append(Flows, map[string]interface{}{"flow_id": d.FlowId, "department_id": d.DepartmentId,
							"leader_id": d.LeaderId})
					}
					data[v.OrderFlow] = map[string]interface{}{"order_name": v.OrderName, "flows": Flows}
				}
			}
		}
		Response.Data = data
	}
}

// @Tags 工单管理
// @Summary 查询工单类型
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/work_order/type [get]
func QueryWorkOrderType(c *gin.Context) {
	//查询工单类型
	var (
		err           error
		WorkOrderType []databases.WorkOrderType
		Response      = common.Response{C: c}
	)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Find(&WorkOrderType)
		Response.Data = WorkOrderType
	}
}
