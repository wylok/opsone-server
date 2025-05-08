package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/auth_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 部门管理
// @Summary 部门详情接口
// @Produce  json
// @Security ApiKeyAuth
// @Param department_ids query []string false "部门id"
// @Param department_name query string false "部门名称"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/department [get]
func QueryDepartment(c *gin.Context) {
	//部门详情接口
	var (
		JsonData   = auth_conf.Department{}
		Department []databases.Department
		Business   []databases.Business
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		data := map[string]interface{}{}
		var val interface{}
		tx := db.Order("department.create_at desc")
		if len(JsonData.DepartmentIds) > 0 {
			JsonData.DepartmentIds = kits.FormListFormat(JsonData.DepartmentIds)
			tx = tx.Where("department_id in ?", JsonData.DepartmentIds)
		}
		if JsonData.DepartmentName != "" {
			tx = tx.Where("department_name like ?", "%"+JsonData.DepartmentName+"%")
		}
		db.Find(&Department)
		if len(Department) > 0 {
			for _, d := range Department {
				bus := map[string]interface{}{}
				//部门附加业务组信息
				sql := "join department_business on department_business.business_id=business.business_id" +
					" and department_business.department_id=?"
				db.Joins(sql, d.DepartmentId).Find(&Business)
				bus["business"] = false
				bus["department"] = false
				bus["department_name"] = d.DepartmentName
				bus["business_group"] = map[string]string{}
				if len(Business) > 0 {
					bus["business"] = true
					for _, b := range Business {
						bus["business_group"].(map[string]string)[b.BusinessId] = b.BusinessName
					}
				} else {
					bus["department"] = true
				}
				db.Where("parent_id=?", d.DepartmentId).Find(&Department)
				if len(Department) > 0 {
					bus["business"] = false
					if len(Business) == 0 {
						bus["department"] = true
					}
				} else {
					if len(Business) == 0 {
						bus["business"] = true
						bus["department"] = true
					}
				}
				data[d.DepartmentId] = bus
			}
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, val = p.Paging(&Department)
		data["departments"] = val
		Response.Data = data
	}
}

// @Tags 部门管理
// @Summary 服务树接口
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/tree [get]
func QueryTree(c *gin.Context) {
	//服务树接口
	var (
		Department  []databases.Department
		Tenants     []databases.Tenants
		Response    = common.Response{C: c}
		err         error
		parentIds   = map[string]struct{}{}
		data        []interface{}
		parents     = map[string]interface{}{}
		rootParents []string
	)
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
		db.Find(&Department)
		if len(Department) > 0 {
			for _, d := range Department {
				parents[d.DepartmentId] = map[string]interface{}{"id": d.DepartmentId,
					"label": d.DepartmentName, "children": []interface{}{}}
				parentIds[d.ParentId] = struct{}{}
				if d.ParentId == "None" {
					rootParents = append(rootParents, d.DepartmentId)
				}
			}
			var ids []string
			for p := range parentIds {
				ids = append(ids, p)
			}
			db.Where("department_id not in ?", ids).Find(&Department)
			if len(Department) > 0 {
				parentIds = map[string]struct{}{}
				for _, v := range Department {
					if parents[v.ParentId] != nil {
						children := parents[v.ParentId].(map[string]interface{})["children"].([]interface{})
						children = append(children, map[string]interface{}{"id": v.DepartmentId, "label": v.DepartmentName})
						parents[v.ParentId].(map[string]interface{})["children"] = children
						parentIds[v.ParentId] = struct{}{}
					}
				}
			}
		loop:
			if len(parentIds) > 0 {
				var ids []string
				for p := range parentIds {
					ids = append(ids, p)
				}
				parentIds = map[string]struct{}{}
				for _, id := range ids {
					db.Where("department_id = ?", id).Find(&Department)
					if len(Department) > 0 {
						for _, v := range Department {
							if parents[v.ParentId] != nil {
								children := parents[v.ParentId].(map[string]interface{})["children"].([]interface{})
								children = append(children, map[string]interface{}{"id": v.DepartmentId, "label": v.DepartmentName,
									"children": parents[v.DepartmentId].(map[string]interface{})["children"]})
								parents[v.ParentId].(map[string]interface{})["children"] = children
								parentIds[v.ParentId] = struct{}{}
							}
						}
					}
				}
				goto loop
			}
		}
		if len(parents) > 0 {
			for _, v := range rootParents {
				_, ok := parents[v]
				if ok {
					data = append(data, parents[v])
				}
			}
		}
	i:
		db.Where("tenant_id=?", platform_conf.TenantId).First(&Tenants)
		if len(Tenants) == 0 {
			dt := databases.Tenants{TenantId: platform_conf.TenantId, TenantName: "OPSONE", TenantDesc: "OPSONE", CreateAt: time.Now(), UpdateAt: time.Now()}
			err = db.Create(&dt).Error
			if err == nil {
				goto i
			}
		}
		Response.Data = []map[string]interface{}{{"id": platform_conf.TenantId, "label": Tenants[0].TenantName,
			"children": data}}
	}
}

// @Tags 部门管理
// @Summary 新建部门接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.CreateDepartment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/department [post]
func AddDepartment(c *gin.Context) {
	//新建部门
	var (
		sqlErr             error
		JsonData           = auth_conf.CreateDepartment{}
		Department         []databases.Department
		DepartmentBusiness []databases.DepartmentBusiness
		Response           = common.Response{C: c}
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
		db.Where("department_name=?", JsonData.DepartmentName).First(&Department)
		if len(Department) == 0 {
			//写入部门信息
			if JsonData.ParentId == "" {
				JsonData.ParentId = "None"
			} else {
				db.Select("id").Where("department_id=?", JsonData.ParentId).Find(&DepartmentBusiness)
				if len(DepartmentBusiness) > 0 {
					panic("业务组已存在,不允许新建子部门")
				}
			}
			departmentId := kits.RandString(8)
			err = db.Transaction(func(tx *gorm.DB) error {
				u := databases.Department{DepartmentId: departmentId, DepartmentName: JsonData.DepartmentName,
					DepartmentDesc: JsonData.DepartmentDesc, ParentId: JsonData.ParentId, CreateAt: time.Now(),
					UpdateAt: time.Now()}
				if err = tx.Create(&u).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New("部门名称已存在")
		}
	}
}

// @Tags 部门管理
// @Summary 部门变更接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.ChangeDepartment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/department [put]
func ModifyDepartment(c *gin.Context) {
	//部门变更
	var (
		JsonData   = auth_conf.ChangeDepartment{}
		Department []databases.Department
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
		db.Where("department_id=?", JsonData.DepartmentId).First(&Department)
		if len(Department) > 0 {
			db.Where("department_name=?", JsonData.DepartmentName).First(&Department)
			if len(Department) == 0 {
				upData := databases.Department{UpdateAt: time.Now()}
				if JsonData.DepartmentName != "" {
					upData.DepartmentName = JsonData.DepartmentName
				}
				if JsonData.DepartmentDesc != "" {
					upData.DepartmentDesc = JsonData.DepartmentDesc
				}
				if JsonData.ParentId != "" {
					upData.ParentId = JsonData.ParentId
				}
				if err == nil {
					err = db.Model(&Department).Where("department_id=?", JsonData.DepartmentId).Updates(upData).Error
				}
			} else {
				err = errors.New("部门名称已存在")
			}
		} else {
			err = errors.New("部门不存在")
		}
	}
}

// @Tags 部门管理
// @Summary 注销部门接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.DeleteDepartment true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/department [delete]
func DelDepartment(c *gin.Context) {
	//注销部门
	var (
		sqlErr             error
		JsonData           = auth_conf.DeleteDepartment{}
		Users              []databases.Users
		Department         []databases.Department
		DepartmentBusiness []databases.DepartmentBusiness
		Response           = common.Response{C: c}
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
		db.Where("department_id=?", JsonData.DepartmentId).Find(&Department)
		if len(Department) > 0 {
			db.Where("department_id=?", JsonData.DepartmentId).Find(&Users)
			if len(Users) > 0 {
				err = errors.New("注销部门名下不能有用户存在")
			}
			db.Where("parent_id=?", JsonData.DepartmentId).First(&Department)
			if len(Department) > 0 {
				err = errors.New("注销部门名下不能有子部门存在")
			}
			if err == nil {
				err = db.Transaction(func(tx *gorm.DB) error {
					if err = tx.Where("department_id = ?", JsonData.DepartmentId).Delete(&Department).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("department_id = ?", JsonData.DepartmentId).Delete(&DepartmentBusiness).Error; err != nil {
						sqlErr = err
					}
					return sqlErr
				})
				if err == nil {
					//删除部门异步消息通知
					rc.SAdd(ctx, platform_conf.DepartmentDeleteKey, JsonData.DepartmentId)
				}
			}
		} else {
			err = errors.New("无效部门ID")
		}
	}
}
