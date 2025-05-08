package monitor

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inner/conf/monitor_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
	"time"
)

// @Tags 监控平台
// @Summary 监控进程查询
// @Produce  json
// @Security ApiKeyAuth
// @Param host_id query string false "主机ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/process [get]
func QueryProcess(c *gin.Context) {
	//监控进程查询
	var (
		JsonData       monitor_conf.QueryProcess
		Response       = common.Response{C: c}
		MonitorProcess []databases.MonitorProcess
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
		if JsonData.HostId != "" {
			db.Where("host_id=? and status=?",
				JsonData.HostId, "active").Find(&MonitorProcess)
		} else {
			db.Distinct("process").Where("status=?", "active").Find(&MonitorProcess)
		}
		if len(MonitorProcess) > 0 {
			Response.Data = MonitorProcess
		}
	}
}

// @Tags 监控平台
// @Summary 新增监控进程
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.AddProcess true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/process [post]
func AddProcess(c *gin.Context) {
	//新增监控进程
	var (
		sqlErr         error
		JsonData       monitor_conf.AddProcess
		Response       = common.Response{C: c}
		MonitorProcess []databases.MonitorProcess
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
		db.Where("host_id in ? and process in ?", JsonData.HostIds, JsonData.Process).Find(&MonitorProcess)
		if len(MonitorProcess) == 0 {
			err = db.Transaction(func(tx *gorm.DB) error {
				for _, HostId := range JsonData.HostIds {
					for _, p := range JsonData.Process {
						jp := databases.MonitorProcess{HostId: HostId, Process: p, CreateTime: time.Now(), Status: "active"}
						if err = tx.Create(&jp).Error; err != nil {
							sqlErr = err
						}
					}
				}
				return sqlErr
			})
		} else {
			err = errors.New(strings.Join(JsonData.Process, ",") + "包含已经存在的进程名称")
		}
	}
}

// @Tags 监控平台
// @Summary 删除监控进程
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteProcess true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/process [delete]
func DeleteProcess(c *gin.Context) {
	//删除监控进程
	var (
		JsonData       monitor_conf.DeleteProcess
		Response       = common.Response{C: c}
		MonitorProcess []databases.MonitorProcess
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Process == nil {
			err = db.Where("host_id in ?", JsonData.HostIds).Delete(&MonitorProcess).Error
		} else {
			err = db.Where("host_id in ? and process in ?", JsonData.HostIds, JsonData.Process).Delete(&MonitorProcess).Error
		}
	}
}

// @Tags 监控平台
// @Summary 组监控进程查询
// @Produce  json
// @Security ApiKeyAuth
// @Param group_id query string true "资源组ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/group/process [get]
func QueryGroupProcess(c *gin.Context) {
	//组监控进程查询
	var (
		JsonData     monitor_conf.QueryGroupProcess
		Response     = common.Response{C: c}
		GroupProcess []databases.GroupProcess
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
		db.Where("group_id=?", JsonData.GroupId).Find(&GroupProcess)
		if len(GroupProcess) > 0 {
			Response.Data = GroupProcess
		}
	}
}

// @Tags 监控平台
// @Summary 新增资源组监控进程
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.AddGroupProcess true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/group/process [post]
func AddGroupProcess(c *gin.Context) {
	//新增资源组监控进程
	var (
		sqlErr         error
		JsonData       monitor_conf.AddGroupProcess
		Response       = common.Response{C: c}
		GroupProcess   []databases.GroupProcess
		MonitorProcess []databases.MonitorProcess
		GroupServer    []databases.GroupServer
		hostIds        []string
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("group_id = ? and process in ?", JsonData.GroupId, JsonData.Process).Find(&GroupProcess)
		if len(GroupProcess) == 0 {
			db.Where("group_id = ?", JsonData.GroupId).Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, d := range GroupServer {
					hostIds = append(hostIds, d.HostId)
				}
			}
			if hostIds != nil {
				err = db.Transaction(func(tx *gorm.DB) error {
					for _, p := range JsonData.Process {
						jp := databases.GroupProcess{GroupId: JsonData.GroupId, Process: p, CreateTime: time.Now()}
						if err = tx.Create(&jp).Error; err != nil {
							sqlErr = err
						}
						for _, h := range hostIds {
							db.Where("host_id=? and process=?", h, p).First(&MonitorProcess)
							if len(MonitorProcess) == 0 {
								mp := databases.MonitorProcess{HostId: h, Process: p, CreateTime: time.Now(), Status: "active"}
								if err = tx.Create(&mp).Error; err != nil {
									sqlErr = err
								}
							}
						}
					}
					return sqlErr
				})
			}
			if err != nil {
				Response.Err = err
			}
		} else {
			err = errors.New(strings.Join(JsonData.Process, ",") + "包含已经存在的进程名称")
		}
	}
}

// @Tags 监控平台
// @Summary 删除监控进程
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  monitor_conf.DeleteGroupProcess true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/group/process [delete]
func DeleteGroupProcess(c *gin.Context) {
	//删除监控进程
	var (
		sqlErr         error
		JsonData       monitor_conf.DeleteGroupProcess
		Response       = common.Response{C: c}
		GroupProcess   []databases.GroupProcess
		MonitorProcess []databases.MonitorProcess
		GroupServer    []databases.GroupServer
		hostIds        []string
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
		db.Where("group_id = ?", JsonData.GroupId).Find(&GroupServer)
		if len(GroupServer) > 0 {
			for _, d := range GroupServer {
				hostIds = append(hostIds, d.HostId)
			}
		}
		if hostIds != nil {
			err = db.Transaction(func(tx *gorm.DB) error {
				if JsonData.Process == nil {
					if err = tx.Where("group_id = ?", JsonData.GroupId).Delete(&GroupProcess).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&MonitorProcess).Error; err != nil {
						sqlErr = err
					}
				} else {
					if err = tx.Where("group_id = ? and process in ?", JsonData.GroupId,
						JsonData.Process).Delete(&GroupProcess).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ? and process in ?", hostIds,
						JsonData.Process).Delete(&MonitorProcess).Error; err != nil {
						sqlErr = err
					}
				}
				return sqlErr
			})
		}
		if err != nil {
			Response.Err = err
		}
	}
}

// @Tags 监控平台
// @Summary 查询进程TOP
// @Produce  json
// @Security ApiKeyAuth
// @Param host_id query string true "服务器ID"
// @Success 200 {} json "{success:true,message:"ok",data:[]}"
// @Router /api/v1/monitor/process/top [get]
func QueryProcessTop(c *gin.Context) {
	//查询进程TOP
	var (
		JsonData monitor_conf.QueryProcessTop
		Response = common.Response{C: c}
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
		cpu := rc.HGetAll(ctx, monitor_conf.ProcessTop+"_cpu_"+JsonData.HostId).Val()
		mem := rc.HGetAll(ctx, monitor_conf.ProcessTop+"_mem_"+JsonData.HostId).Val()
		Response.Data = map[string]interface{}{"cpu": cpu, "mem": mem}
	}
}
