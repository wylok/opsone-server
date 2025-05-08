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
// @Summary 脚本上传接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/script/upload [post]
func ScriptUpdate(c *gin.Context) {
	//脚本上传接口
	var (
		sqlErr         error
		Response       = common.Response{C: c}
		JobScript      []databases.JobScript
		ScriptContents []databases.ScriptContents
		files          = map[string]string{}
		UserId         = c.GetString("user_id")
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
		purpose, ok := c.GetQuery("purpose")
		if !ok {
			purpose = "job"
		}
		ScriptId, _ := c.GetQuery("script_id")
		fs := form.File["file"]
		for _, f := range fs {
			if ScriptId == "" {
				ScriptId = kits.RandString(8)
			}
			files[ScriptId] = f.Filename
			F, _ := f.Open()
			b, _ := io.ReadAll(F)
			var scriptType string
			if strings.Contains(f.Filename, ".sh") {
				scriptType = "Shell"
			}
			if strings.Contains(f.Filename, ".py") {
				scriptType = "Python"
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				for ScriptId, fileName := range files {
					db.Where("script_id=?", ScriptId).First(&JobScript)
					if len(JobScript) == 0 {
						js := databases.JobScript{ScriptId: ScriptId, ScriptType: scriptType,
							ScriptName: fileName, ScriptDesc: "", ScriptPurpose: purpose,
							UserId: UserId, CreateTime: time.Now()}
						if err = tx.Create(&js).Error; err != nil {
							sqlErr = err
						}
						if sqlErr == nil {
							sc := databases.ScriptContents{ScriptId: ScriptId, ScriptName: fileName, ScriptContent: string(b)}
							if err = tx.Create(&sc).Error; err != nil {
								sqlErr = err
							}
						}
					} else {
						if err = db.Model(&JobScript).Where("script_id=?", ScriptId).Updates(
							databases.JobScript{ScriptType: scriptType, ScriptName: fileName,
								ScriptPurpose: purpose, UserId: UserId, CreateTime: time.Now()}).Error; err != nil {
							sqlErr = err
						}
						if sqlErr == nil {
							sqlErr = db.Model(&ScriptContents).Where("script_id=?", ScriptId).Updates(
								databases.ScriptContents{ScriptContent: string(b)}).Error
						}
					}
				}
				return sqlErr
			})
		}
	}
}

// @Tags 作业平台
// @Summary 脚本列表接口
// @Produce  json
// @Security ApiKeyAuth
// @Param script_id query string false "脚本ID"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/job/script [get]
func ScriptList(c *gin.Context) {
	//脚本列表接口
	var (
		JobScript []databases.JobScript
		JsonData  job_conf.QScript
		Response  = common.Response{C: c}
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
		if JsonData.ScriptIds != nil {
			JsonData.ScriptIds = kits.FormListFormat(JsonData.ScriptIds)
			tx = tx.Where("script_id in ?", JsonData.ScriptIds)
		}
		if JsonData.Purpose != "" {
			tx = tx.Where("script_purpose = ?", JsonData.Purpose)
		}
		if JsonData.NotPage {
			tx.Find(&JobScript)
			Response.Data = JobScript
		} else {
			p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
			Response.Pages, Response.Data = p.Paging(&JobScript)
		}
	}
}

// @Tags 作业平台
// @Summary 脚本详情接口
// @Produce  json
// @Security ApiKeyAuth
// @Param script_id query string true "作业ID"
// @Success 200 {} json "{success:true,message:"ok",data:{}}"
// @Router /api/v1/job/script/detail [get]
func ScriptDetail(c *gin.Context) {
	//脚本详情接口
	var (
		JobScript      []databases.JobScript
		ScriptContents []databases.ScriptContents
		JsonData       job_conf.ScriptDetail
		Response       = common.Response{C: c}
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
		db.Where("script_id = ?", JsonData.ScriptId).First(&JobScript)
		if len(JobScript) > 0 {
			db.Where("script_id = ?", JsonData.ScriptId).First(&ScriptContents)
			if len(ScriptContents) > 0 {
				Response.Data = map[string]interface{}{"content": ScriptContents[0].ScriptContent}
			} else {
				err = errors.New(JsonData.ScriptId + "未找到脚本ID对应的脚本文件")
			}
		} else {
			err = errors.New(JsonData.ScriptId + "脚本ID不存在")
		}
	}
}

// @Tags 作业平台
// @Summary 脚本删除接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.ScriptDelete true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/script [delete]
func ScriptDelete(c *gin.Context) {
	//脚本删除接口
	var (
		sqlErr         error
		Response       = common.Response{C: c}
		JsonData       job_conf.ScriptDelete
		JobScript      []databases.JobScript
		JobRun         []databases.JobRun
		ScriptContents []databases.ScriptContents
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
		db.Where("script_id in ? and status=?", JsonData.ScriptIds, "running").Find(&JobRun)
		if len(JobRun) > 0 {
			err = errors.New("禁止删除正在执行的脚本")
		}
		if err == nil {
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = db.Where("script_id in ?", JsonData.ScriptIds).Delete(&JobScript).Error; err != nil {
					sqlErr = err
				}
				if err = db.Where("script_id in ?", JsonData.ScriptIds).Delete(&JobRun).Error; err != nil {
					sqlErr = err
				}
				if err = db.Where("script_id in ?", JsonData.ScriptIds).Delete(&ScriptContents).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})

		}
	}
}

// @Tags 作业平台
// @Summary 脚本修改接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.ScriptModify true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/script [put]
func ScriptModify(c *gin.Context) {
	//脚本修改接口
	var (
		Response = common.Response{C: c}
		JsonData job_conf.ScriptModify
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
		err = db.Model(&databases.JobScript{}).Where("script_id = ?", JsonData.ScriptId).Updates(
			databases.JobScript{ScriptDesc: JsonData.ScriptDesc}).Error
	}
}

// @Tags 作业平台
// @Summary 脚本执行接口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  job_conf.RunScript true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/job/script/run [post]
func ScriptRun(c *gin.Context) {
	//脚本执行接口
	var (
		sqlErr         error
		files          []string
		JsonData       job_conf.RunScript
		ScriptContents []databases.ScriptContents
		GroupServer    []databases.GroupServer
		Response       = common.Response{C: c}
		Encrypt        = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		UserId         = c.GetString("user_id")
		HostIds        []string
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
		//调用cmdb接口
		if len(JsonData.AssetGroupIds) > 0 {
			db.Where("group_id in ?", JsonData.AssetGroupIds).Find(&GroupServer)
			if len(GroupServer) > 0 {
				for _, d := range GroupServer {
					HostIds = append(HostIds, d.HostId)
				}
			}
		}
		if JsonData.HostIds != nil {
			HostIds = JsonData.HostIds
		}
		if JsonData.DstPath == "" {
			JsonData.DstPath = "/tmp/"
		}
		if len(HostIds) > 0 {
			//判断script_id对应脚本是否存在
			db.Where("script_id=?", JsonData.ScriptId).First(&ScriptContents)
			if len(ScriptContents) > 0 {
				files = append(files, ScriptContents[0].ScriptName)
				rc.HSet(ctx, "job_send_file_"+JsonData.ScriptId, ScriptContents[0].ScriptName, ScriptContents[0].ScriptContent)
				rc.Expire(ctx, "job_send_file_"+JsonData.ScriptId, 30*time.Minute)
				runTime := time.Now()
				if JsonData.Cron {
					runTime = carbon.Parse(JsonData.RunTime).Carbon2Time()
				}
				JobId := kits.RandString(8)
				if files != nil {
					failKey := "job_exec_results_fail" + JobId
					for _, h := range HostIds {
						if rc.HExists(ctx, platform_conf.OfflineAssetKey, h).Val() {
							rc.HSet(ctx, failKey, h, "fail")
						}
					}
					// 写入表数据
					err = db.Transaction(func(tx *gorm.DB) error {
						// 写入表数据
						jp := databases.JobOverview{JobId: JobId,
							JobType: "job_script", Cron: cast.ToInt(JsonData.Cron), Contents: strings.Join(files, ","),
							Counts: int64(len(HostIds)), Success: 0, Fail: 0, UserId: UserId, CreateTime: time.Now()}
						if err = tx.Create(&jp).Error; err != nil {
							sqlErr = err
						}
						for _, hostId := range HostIds {
							// 写入表数据
							je := databases.JobRun{JobId: JobId, HostId: hostId,
								ScriptId: JsonData.ScriptId, Cron: cast.ToInt(JsonData.Cron),
								RunTime: runTime, Status: "pending"}
							if err = tx.Create(&je).Error; err != nil {
								sqlErr = err
							}
						}
						return sqlErr
					})
				} else {
					err = errors.New("未获取到文件列表")
				}
				if err == nil {
					if !JsonData.Cron {
						//文件传输下发
						if JsonData.DstPath == "" {
							JsonData.DstPath = "/tmp/" + JsonData.ScriptId + "/"
						}
						for _, hostId := range HostIds {
							m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
								"job_id": JobId, "dst_path": JsonData.DstPath, "job_type": "job_script",
								"files": files, "script_id": JsonData.ScriptId, "host_id": hostId}), true)
							platform_conf.Wch <- map[string]interface{}{"jobFile": kits.MapToJson(map[string]interface{}{"jobFile": m}),
								"job_type": "job_script", "host_id": hostId,
								"msg_time": time.Now().Format("2006-01-02 15:04:05")}
							if err == nil {
								err = db.Model(&databases.JobRun{}).Where("job_id=? and host_id=?",
									JobId, hostId).Updates(databases.JobRun{Status: "running"}).Error
								Log.Info("主机:" + hostId + " 脚本执行作业任务:" + JobId + "已下发")
							}
						}
					}
					Response.Data = map[string]string{"job_id": JobId}
				}
			} else {
				err = errors.New("未找到ScriptId(" + JsonData.ScriptId + ")对应的文件信息")
			}
		} else {
			err = errors.New("获取主机列表出现异常")
		}
	}
}

// @Tags 作业平台
// @Summary 脚本运行列表接口
// @Produce  json
// @Security ApiKeyAuth
// @Param job_id query string false "作业ID"
// @Param script_id query string false "脚本ID"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/job/script/run [get]
func ScriptRunList(c *gin.Context) {
	//脚本运行列表接口
	var (
		JobRun   []databases.JobRun
		JsonData job_conf.QRScript
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
		if JsonData.ScriptId != "" {
			tx = tx.Where("script_id = ?", JsonData.ScriptId)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&JobRun)
	}
}
