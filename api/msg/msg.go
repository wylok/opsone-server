package msg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/msg_conf"
	"inner/modules/common"
	"inner/modules/databases"
)

var (
	db = databases.DB
)

// @Tags 消息中心
// @Summary 消息查询
// @Produce  json
// @Security ApiKeyAuth
// @Param msg_id query string false "消息ID"
// @Param msg_id query string false "消息ID"
// @Param msg_type query string false "消息类型"
// @Param level query string false "消息等级"
// @Param status query string false "消息状态"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/msg [get]
func QueryMsg(c *gin.Context) {
	//消息查询
	var (
		sqlErr   error
		Msg      []databases.Msg
		JsonData msg_conf.QMsg
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
			JsonData.PerPage = 15
		}
		tx := db.Order("create_time desc")
		// 部分参数匹配
		if JsonData.MsgId != "" {
			tx = tx.Where("msg_id = ?", JsonData.MsgId)
		}
		if JsonData.Title != "" {
			tx = tx.Where("title = ?", JsonData.Title)
		}
		if JsonData.MsgType != "" {
			tx = tx.Where("msg_type = ?", JsonData.MsgType)
		}
		if JsonData.Level != "" {
			tx = tx.Where("level = ?", JsonData.Level)
		}
		if JsonData.Status != "" {
			tx = tx.Where("status = ?", JsonData.Status)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&Msg)
	}
}

// @Tags 消息中心
// @Summary 消息详情
// @Produce  json
// @Security ApiKeyAuth
// @Param msg_ids query array true "消息ID"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:{}}"
// @Router /api/v1/msg/detail [get]
func Detail(c *gin.Context) {
	//消息详情
	var (
		sqlErr     error
		Msg        []databases.Msg
		MsgContent []databases.MsgContent
		JsonData   msg_conf.MsgDetail
		Response   = common.Response{C: c}
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
		db.Where("msg_id = ?", JsonData.MsgId).First(&MsgContent)
		if len(MsgContent) > 0 {
			var content interface{}
			err = json.Unmarshal([]byte(MsgContent[0].Content), &content)
			if err == nil {
				Response.Data = []map[string]interface{}{cast.ToStringMap(content)}
				db.Model(&Msg).Where("msg_id = ? and status = ?", MsgContent[0].MsgId, "Unread").Updates(
					databases.Msg{Status: "Read"})
			}
		} else {
			err = fmt.Errorf(JsonData.MsgId + "消息内容不存在")
		}
	}
}

// @Tags 消息中心
// @Summary 删除消息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  msg_conf.DMsg true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/msg [delete]
func DeleteMsg(c *gin.Context) {
	//删除消息
	var (
		sqlErr     error
		Msg        []databases.Msg
		MsgContent []databases.MsgContent
		JsonData   = msg_conf.DMsg{}
		Response   = common.Response{C: c}
	)
	err := c.BindJSON(&JsonData)
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
	if err == nil && len(JsonData.MsgIds) > 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Where("msg_id in ?", JsonData.MsgIds).Delete(&Msg).Error
			if err != nil {
				sqlErr = err
			}
			err = tx.Where("msg_id in ?", JsonData.MsgIds).Delete(&MsgContent).Error
			if err != nil {
				sqlErr = err
			}
			return sqlErr
		})
	}
}
