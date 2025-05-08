package job

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/job_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags 作业平台
// @Summary 作业总览接口
// @Produce  json
// @Security ApiKeyAuth
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/job/overview [get]
func Overview(c *gin.Context) {
	//作业总览接口
	var (
		JobOverview []databases.JobOverview
		JsonData    job_conf.QOverview
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回结果
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
			JsonData.PerPage = 10
		}
		tx := db.Order("create_time desc")
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&JobOverview)
	}
}

// @Tags 作业平台
// @Summary 作业删除接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.OverviewDelete true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/overview [delete]
func OverviewDelete(c *gin.Context) {
	//作业删除接口
	var (
		sqlErr       error
		JsonData     job_conf.OverviewDelete
		JobExec      []databases.JobExec
		JobFile      []databases.JobFile
		FileContents []databases.FileContents
		JobOverview  []databases.JobOverview
		Response     = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if sqlErr != nil {
			err = sqlErr
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("job_id in ?", JsonData.JobIds).Find(&JobOverview)
		if len(JobOverview) > 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("job_id in ?", JsonData.JobIds).Delete(&JobOverview).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("job_id in ?", JsonData.JobIds).Delete(&JobExec).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("job_id in ?", JsonData.JobIds).Delete(&JobFile).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("job_id in ?", JsonData.JobIds).Delete(&FileContents).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
		} else {
			err = errors.New("无法删除系统作业任务ID")
		}
	}
}

// @Tags 作业平台
// @Summary 作业结果接口
// @Produce  json
// @Security ApiKeyAuth
// @Param job_id query string true "作业ID"
// @Param host_id query string false "主机ID"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{success:true,message:"ok",data:{}}"
// @Router /api/v1/job/results [get]
func Results(c *gin.Context) {
	//作业结果接口
	var (
		JobResults []databases.JobResults
		JsonData   job_conf.JobResults
		Response   = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if JsonData.Page == 0 {
		JsonData.Page = 1
	}
	if JsonData.PerPage == 0 {
		JsonData.PerPage = 10
	}
	if err == nil {
		tx := db.Where("job_id=?", JsonData.JobId)
		if JsonData.HostId != "" {
			tx.Where("host_id=?", JsonData.HostId)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&JobResults)
	}
}
