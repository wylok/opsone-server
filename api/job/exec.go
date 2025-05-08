package job

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/job_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"time"
)

// @Tags 作业平台
// @Summary 命令执行接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.RExec true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/exec [post]
func ExecUpdate(c *gin.Context) {
	//消息上报接口
	var (
		sqlErr      error
		JsonData    job_conf.RExec
		Encrypt     = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		Response    = common.Response{C: c}
		GroupServer []databases.GroupServer
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
	if err == nil {
		// 初始化数据库连接
		if len(JsonData.HostIds) == 0 && len(JsonData.AssetGroupIds) == 0 {
			err = errors.New("host_ids或者asset_group_ids参数二选一为必填项")
		} else {
			if len(JsonData.HostIds) == 0 {
				db.Where("group_id in ?", JsonData.AssetGroupIds).Find(&GroupServer)
				if len(GroupServer) > 0 {
					for _, d := range GroupServer {
						JsonData.HostIds = append(JsonData.HostIds, d.HostId)
					}
				}
			}
			if err == nil {
				UserId := c.GetString("user_id")
				JobId := kits.RandString(8)
				RunTime := time.Now()
				failKey := "job_exec_results_fail" + JobId
				if JsonData.Cron {
					RunTime = carbon.Parse(JsonData.RunTime).Carbon2Time()
				}
				for _, h := range JsonData.HostIds {
					if rc.HExists(ctx, platform_conf.OfflineAssetKey, h).Val() {
						rc.HSet(ctx, failKey, h, "fail")
					}
				}
				// 写入表数据
				err = db.Transaction(func(tx *gorm.DB) error {
					// 写入表数据
					jp := databases.JobOverview{UserId: UserId, JobId: JobId, JobType: "job_exec",
						Cron: cast.ToInt(JsonData.Cron), Contents: JsonData.Exec, Counts: int64(len(JsonData.HostIds)),
						Success: 0, Fail: 0, CreateTime: time.Now()}
					if err = tx.Create(&jp).Error; err != nil {
						sqlErr = err
					}
					for _, hostId := range JsonData.HostIds {
						// 写入表数据
						je := databases.JobExec{JobId: JobId, HostId: hostId, Exec: JsonData.Exec,
							Cron: cast.ToInt(JsonData.Cron), RunTime: RunTime, Status: "pending"}
						if err = tx.Create(&je).Error; err != nil {
							sqlErr = err
						}
					}
					return sqlErr
				})
				if err == nil {
					if !JsonData.Cron {
						//作业任务下发
						for _, hostId := range JsonData.HostIds {
							m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
								"job_id": JobId, "exec": JsonData.Exec, "job_type": "job_exec", "host_id": hostId}), true)
							platform_conf.Wch <- map[string]interface{}{"jobShell": kits.MapToJson(map[string]interface{}{"jobShell": m}),
								"job_type": "job_exec", "host_id": hostId,
								"msg_time": time.Now().Format("2006-01-02 15:04:05")}
							if err == nil {
								db.Model(&databases.JobExec{}).Where("job_id=? and host_id=?",
									JobId, hostId).Updates(databases.JobExec{Status: "running"})
							}
						}
						Log.Info("命令执行作业任务:" + JobId + "已下发")
					}
					Response.Data = map[string]string{"job_id": JobId}
				}
			}
		}
	}
}

// @Tags 作业平台
// @Summary 命令执行列表接口
// @Produce  json
// @Security ApiKeyAuth
// @Param job_id query string false "作业ID"
// @Param run_time query string false "运行时间"
// @Param status query string false "作业状态"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/job/exec [get]
func ExecList(c *gin.Context) {
	//命令执行列表接口
	var (
		JobExec  []databases.JobExec
		JsonData job_conf.QExec
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
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 10
		}
		tx := db.Order("id desc")
		// 部分参数匹配
		if JsonData.JobId != "" {
			tx = tx.Where("job_id = ?", JsonData.JobId)
		}
		if JsonData.Runtime != "" {
			tx = tx.Where("run_time like ?", "%"+JsonData.Runtime+"%")
		}
		if JsonData.Status != "" {
			tx = tx.Where("status = ?", JsonData.Status)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&JobExec)
	}
}
