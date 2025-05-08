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
	"io"
	"strings"
	"time"
)

// @Tags 作业平台
// @Summary 文件上传接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/file/upload [post]
func FileUpdate(c *gin.Context) {
	//文件上传接口
	var (
		Response = common.Response{C: c}
	)
	form, err := c.MultipartForm()
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		fs := form.File["file"]
		// 生成作业ID
		JobId := kits.RandString(8)
		for _, f := range fs {
			F, _ := f.Open()
			b, _ := io.ReadAll(F)
			fc := databases.FileContents{JobId: JobId, FileName: f.Filename, FileContent: b}
			err = db.Create(&fc).Error
			if err == nil {
				rc.SAdd(ctx, job_conf.SendFileJobKey, JobId)
			} else {
				//清除上传文件
				rc.SRem(ctx, job_conf.SendFileJobKey, JobId)
				break
			}
		}
		if err == nil {
			Response.Data = map[string]interface{}{"job_id": JobId}
		}
	}
}

// @Tags 作业平台
// @Summary 文件分发接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.RFile true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/file/send [post]
func FileSend(c *gin.Context) {
	//文件分发接口
	var (
		sqlErr       error
		files        []string
		JsonData     job_conf.RFile
		Response     = common.Response{C: c}
		FileContents []databases.FileContents
		GroupServer  []databases.GroupServer
		Encrypt      = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		UserId       = c.GetString("user_id")
	)
	err := c.BindJSON(&JsonData)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if len(JsonData.HostIds) == 0 && len(JsonData.AssetGroupIds) == 0 {
			err = errors.New("host_ids或者asset_group_id参数二选一为必填项")
		} else {
			if len(JsonData.HostIds) == 0 {
				db.Where("group_id in ?", JsonData.AssetGroupIds).Find(&GroupServer)
				if len(GroupServer) > 0 {
					for _, d := range GroupServer {
						JsonData.HostIds = append(JsonData.HostIds, d.HostId)
					}
				}
			}
			if len(JsonData.HostIds) > 0 {
				//判断job_id是否存在
				db.Where("job_id=?", JsonData.JobId).Find(&FileContents)
				if len(FileContents) > 0 {
					failKey := "job_file_results_fail" + JsonData.JobId
					for _, h := range JsonData.HostIds {
						if rc.HExists(ctx, platform_conf.OfflineAssetKey, h).Val() {
							rc.HSet(ctx, failKey, h, "fail")
							rc.Expire(ctx, failKey, 8*time.Hour)
						}
					}
					for _, v := range FileContents {
						files = append(files, v.FileName)
						rc.HSet(ctx, "job_send_file_"+JsonData.JobId, v.FileName, v.FileContent)
					}
					rc.Expire(ctx, "job_send_file_"+JsonData.JobId, 30*time.Minute)
					sendTime := time.Now()
					if JsonData.Cron {
						sendTime = carbon.Parse(JsonData.SendTime).Carbon2Time()
					}
					if files != nil {
						// 写入表数据
						err = db.Transaction(func(tx *gorm.DB) error {
							// 写入表数据
							jp := databases.JobOverview{JobId: JsonData.JobId,
								JobType: "job_file", Cron: cast.ToInt(JsonData.Cron), Contents: strings.Join(files, ","),
								Counts: int64(len(JsonData.HostIds)), Success: 0, Fail: 0, UserId: UserId,
								CreateTime: time.Now()}
							if err = tx.Create(&jp).Error; err != nil {
								sqlErr = err
							}
							for _, hostId := range JsonData.HostIds {
								// 写入表数据
								je := databases.JobFile{JobId: JsonData.JobId, HostId: hostId,
									DstPath: JsonData.DstPath, Files: strings.Join(files, ","),
									Cron:     cast.ToInt(JsonData.Cron),
									SendTime: sendTime, Status: "pending"}
								if err = tx.Create(&je).Error; err != nil {
									sqlErr = err
								}
							}
							return sqlErr
						})
					} else {
						err = errors.New("未获取到文件列表")
					}
					if err == nil && !JsonData.Cron {
						//文件传输下发
						for _, hostId := range JsonData.HostIds {
							m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
								"job_id": JsonData.JobId, "dst_path": JsonData.DstPath, "files": files,
								"job_type": "job_file", "host_id": hostId}), true)
							platform_conf.Wch <- map[string]interface{}{"jobFile": kits.MapToJson(map[string]interface{}{"jobFile": m}),
								"job_type": "job_file", "host_id": hostId,
								"msg_time": time.Now().Format("2006-01-02 15:04:05")}
							if err == nil {
								db.Model(&databases.JobFile{}).Where("job_id=? and host_id=?",
									JsonData.JobId, hostId).Updates(databases.JobFile{Status: "sending"})
							}
						}
						Log.Info("文件分发作业任务:" + JsonData.JobId + "已下发")
					}
				} else {
					err = errors.New("未找到JobId(" + JsonData.JobId + ")对应的文件信息")
				}
			} else {
				err = errors.New("获取主机列表出现异常")
			}
		}
	}
}

// @Tags 作业平台
// @Summary 文件分发列表接口
// @Produce  json
// @Security ApiKeyAuth
// @Param job_id query string false "作业ID"
// @Param status query string false "作业状态"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/job/file [get]
func FileList(c *gin.Context) {
	//文件分发列表接口
	var (
		JobFile  []databases.JobFile
		JsonData job_conf.QFile
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
		if JsonData.Status != "" {
			tx = tx.Where("status = ?", JsonData.Status)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&JobFile)
	}
}
