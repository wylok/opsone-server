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

// @Tags 业务组管理
// @Summary 业务组查询
// @Produce  json
// @Security ApiKeyAuth
// @Param department_id query string false "部门id"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/auth/business [get]
func QueryBusiness(c *gin.Context) {
	//业务组查询
	var (
		JsonData   = auth_conf.Business{}
		Business   []databases.Business
		Department []databases.Department
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
		db.Where("department_id=?", JsonData.DepartmentId).First(&Department)
		if len(Department) > 0 {
			sql := "join department_business on department_business.business_id=business.business_id" +
				" and department_business.department_id = ?"
			tx := db.Joins(sql, JsonData.DepartmentId)
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, Response.Data = p.Paging(&Business)
		} else {
			err = errors.New(JsonData.DepartmentId + "无效的部门ID")
		}
	}
}

// @Tags 业务组管理
// @Summary 新建业务组
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.CreateBusiness true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/business [post]
func AddBusiness(c *gin.Context) {
	//新建业务组
	var (
		sqlErr     error
		JsonData   = auth_conf.CreateBusiness{}
		Business   []databases.Business
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
		db.Where("parent_id=?", JsonData.DepartmentId).First(&Department)
		if len(Department) > 0 {
			err = errors.New("非末级部门不允许创建业务组")
		} else {
			db.Where("business_name=?", JsonData.BusinessName).First(&Business)
			if len(Business) == 0 {
				businessId := kits.RandString(0)
				err = db.Transaction(func(tx *gorm.DB) error {
					//写入业务组信息
					u := databases.Business{BusinessId: businessId, BusinessName: JsonData.BusinessName,
						BusinessDesc: JsonData.BusinessDesc, CreateAt: time.Now()}
					if err = tx.Create(&u).Error; err != nil {
						sqlErr = err
					}
					//写入业务组与部门信息
					deb := databases.DepartmentBusiness{DepartmentId: JsonData.DepartmentId, BusinessId: businessId}
					if err = tx.Create(&deb).Error; err != nil {
						sqlErr = err
					}
					return sqlErr
				})
			} else {
				err = errors.New("业务组名称已存在")
			}
		}
	}
}

// @Tags 业务组管理
// @Summary 业务组变更
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.ChangeBusiness true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/business [put]
func ModifyBusiness(c *gin.Context) {
	//部门变更
	var (
		sqlErr             error
		JsonData           = auth_conf.ChangeBusiness{}
		Business           []databases.Business
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
		db.Where("business.business_id=?", JsonData.BusinessId).First(&Business)
		if len(Business) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				upData := databases.Business{}
				if JsonData.BusinessName != "" {
					upData.BusinessName = JsonData.BusinessName
				}
				if JsonData.BusinessDesc != "" {
					upData.BusinessDesc = JsonData.BusinessDesc
				}
				if JsonData.DepartmentId != "" {
					if err = tx.Model(&DepartmentBusiness).Where("business_id=? and department_id=?",
						JsonData.BusinessId, JsonData.DepartmentId).Updates(
						databases.DepartmentBusiness{DepartmentId: JsonData.DepartmentId}).Error; err != nil {
						sqlErr = err
					}
				}
				if err = tx.Model(&Business).Where("business_id=?", JsonData.BusinessId).Updates(
					upData).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New("业务组不存在")
		}
	}
}

// @Tags 业务组管理
// @Summary 注销业务组
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  auth_conf.DeleteBusiness true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/auth/business [delete]
func DelBusiness(c *gin.Context) {
	//注销业务组
	var (
		sqlErr             error
		JsonData           = auth_conf.DeleteBusiness{}
		BusinessUser       []databases.BusinessUser
		Business           []databases.Business
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
		db.Where("business_id =?", JsonData.BusinessId).Find(&Business)
		if len(Business) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("business_id = ?", JsonData.BusinessId).Delete(&Business).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("business_id = ?", JsonData.BusinessId).Delete(&BusinessUser).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("business_id = ?", JsonData.BusinessId).Delete(&DepartmentBusiness).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
			if err == nil {
				//删除业务组异步消息通知
				rc.SAdd(ctx, platform_conf.BusinessDeleteKey, JsonData.BusinessId)
			}
		} else {
			err = errors.New("无效业务组ID")
		}
	}
}
