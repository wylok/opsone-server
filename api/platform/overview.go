package platform

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strconv"
)

// @Tags 平台管理
// @Summary 总览数据
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/platform/overview [get]
func Overview(c *gin.Context) {
	//总览数据
	var (
		WorkOrder []databases.WorkOrder
		CloudKeys []databases.CloudKeys
		ws        int64
		ck        int64
		Response  = common.Response{C: c}
		err       error
	)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Model(&WorkOrder).Count(&ws)
		db.Model(&CloudKeys).Count(&ck)
		d := rc.HGetAll(ctx, platform_conf.OverViewKey).Val()
		d["WorkOrder"] = strconv.FormatInt(ws, 10)
		d["cloud"] = strconv.FormatInt(ck, 10)
		d["your-node-ip"] = platform_conf.RemoteAddr
		Response.Data = d
	}
}
