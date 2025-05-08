package cloud

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/cloud_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

// @Tags 多云管理
// @Summary Oss列表
// @Produce  json
// @Security ApiKeyAuth
// @Param cloud query string false "公有云"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cloud/oss [get]
func QueryOss(c *gin.Context) {
	//Oss列表
	var (
		CloudOss []databases.CloudOss
		JsonData cloud_conf.QOss
		Response = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 10
		}
		tx := db.Order("CreationDate desc")
		if JsonData.Cloud != "" {
			tx = tx.Where("cloud=?", JsonData.Cloud)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&CloudOss)
	}
}
