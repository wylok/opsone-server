package job

import (
	"fmt"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/golang-module/carbon"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/job_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strings"
	"time"
)

func JobExecResults() {
	var (
		sqlErr  error
		err     error
		JobExec []databases.JobExec
		JobRun  []databases.JobRun
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	Log.Info("监听作业任务结果开始执行......")
	for {
		data := <-platform_conf.Ech
		if data != nil {
			results := cast.ToString(data["stdout"])
			if data["stderr"].(string) != "" {
				results = cast.ToString(data["stderr"])
			}
			Log.Info("接收到作业任务(" + cast.ToString(data["job_id"]) + ")执行结果")
			successKey := "job_exec_results_success" + cast.ToString(data["job_id"])
			failKey := "job_exec_results_fail" + cast.ToString(data["job_id"])
			if cast.ToBool(data["status"]) {
				rc.HSet(ctx, successKey, data["host_id"], "success")
			} else {
				rc.HSet(ctx, failKey, data["host_id"], "fail")
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				tx.Model(&databases.JobOverview{}).Where("job_id=?", cast.ToString(data["job_id"])).Updates(
					databases.JobOverview{Success: cast.ToInt64(len(rc.HGetAll(ctx, successKey).Val())),
						Fail: cast.ToInt64(len(rc.HGetAll(ctx, failKey).Val()))})
				if cast.ToString(data["job_type"]) == "job_script" {
					if err = db.Model(&JobRun).Where("job_id=? and host_id=?", cast.ToString(data["job_id"]),
						cast.ToString(data["host_id"])).Updates(databases.JobRun{Status: "completed"}).Error; err != nil {
						sqlErr = err
					}
				} else {
					if err = db.Model(&JobExec).Where("job_id=? and host_id=?", cast.ToString(data["job_id"]),
						cast.ToString(data["host_id"])).Updates(databases.JobExec{Status: "completed"}).Error; err != nil {
						sqlErr = err
					}
				}
				return sqlErr
			})
			if err == nil {
				jr := databases.JobResults{JobId: cast.ToString(data["job_id"]), HostId: cast.ToString(data["host_id"]),
					Results: results, CreateTime: time.Now()}
				err = db.Create(&jr).Error
			}
		}
	}
}

func JobFileResults() {
	var (
		sqlErr  error
		err     error
		JobFile []databases.JobFile
		JobRun  []databases.JobRun
		Encrypt = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
	}()
	Log.Info("监听文件分发结果开始执行......")
	for {
		data := <-platform_conf.Fch
		if data != nil {
			file := cast.ToString(data["file"])
			jobId := cast.ToString(data["job_id"])
			hostId := cast.ToString(data["host_id"])
			status := cast.ToString(data["status"])
			if cast.ToString(data["job_type"]) == "job_script" {
				exec := "sh " + file
				if strings.Contains(file, ".py") {
					exec = "python3 " + file
				}
				m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
					"job_id":   jobId,
					"exec":     exec,
					"file":     file,
					"job_type": "job_script",
					"host_id":  hostId}), true)
				platform_conf.Wch <- map[string]interface{}{"jobShell": kits.MapToJson(map[string]interface{}{"jobShell": m}),
					"job_type": "job_script", "host_id": hostId,
					"msg_time": time.Now().Format("2006-01-02 15:04:05")}
			} else {
				results := cast.ToString(data["file"]) + "分发" + cast.ToString(data["message"])
				Log.Info("接收到文件分发(" + jobId + ")执行结果")
				successKey := "job_file_results_success" + jobId
				failKey := "job_file_results_fail" + jobId
				if cast.ToBool(status) {
					rc.HSet(ctx, successKey, hostId, "success")
					rc.Expire(ctx, successKey, 8*time.Hour)
				}
				if !cast.ToBool(status) {
					rc.HSet(ctx, failKey, hostId, "fail")
					rc.Expire(ctx, failKey, 8*time.Hour)
				}
				err = db.Transaction(func(tx *gorm.DB) error {
					tx.Model(&databases.JobOverview{}).Where("job_id=?", jobId).Updates(
						databases.JobOverview{Success: cast.ToInt64(len(rc.HGetAll(ctx, successKey).Val())),
							Fail: cast.ToInt64(len(rc.HGetAll(ctx, failKey).Val()))})
					if err = tx.Model(&JobFile).Where("job_id=? and host_id=?", jobId,
						hostId).Updates(databases.JobFile{Status: "completed"}).Error; err != nil {
						sqlErr = err
					}
					db.Where("job_id=? and host_id=?", jobId, hostId).First(&JobRun)
					if len(JobRun) > 0 {
						if err = tx.Model(&JobRun).Where("job_id=? and host_id=?", jobId,
							hostId).Updates(databases.JobRun{Status: "completed"}).Error; err != nil {
							sqlErr = err
						}
					}
					return sqlErr
				})
				if err == nil {
					jr := databases.JobResults{JobId: cast.ToString(data["job_id"]), HostId: cast.ToString(data["host_id"]),
						Results: results, CreateTime: time.Now()}
					err = db.Create(&jr).Error
				}
			}
		}
	}
}

func JobCron() {
	lock := common.SyncMutex{LockKey: "job_cron_lock"}
	//加锁
	if lock.Lock() {
		var (
			sqlErr    error
			err       error
			JobRun    []databases.JobRun
			JobScript []databases.JobScript
			JobExec   []databases.JobExec
			JobFile   []databases.JobFile
			Encrypt   = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		)
		defer func() {
			lock.UnLock(true)
		}()
		Log.Info("定时作业任务开始执行......")
		//读取cron作业任务
		nt := carbon.Time2Carbon(time.Now()).ToDateTimeString()
		h := strings.Split(nt, ":")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					Log.Error(fmt.Sprint(r))
				}
			}()
			db.Where("cron=? and run_time like ? and status=?", 1,
				strings.Join(h[:len(h)-1], ":")+"%", "pending").Find(&JobExec)
			if len(JobExec) > 0 {
				for _, v := range JobExec {
					//作业任务下发
					if rc.HExists(ctx, platform_conf.OfflineAssetKey, v.HostId).Val() {
						failKey := "job_exec_results_fail" + v.JobId
						rc.HSet(ctx, failKey, v.HostId, "fail")
						err = db.Transaction(func(tx *gorm.DB) error {
							db.Model(&databases.JobExec{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobExec{Status: "fail"})
							return sqlErr
						})
						if err == nil {
							jr := databases.JobResults{JobId: v.JobId, HostId: v.HostId,
								Results: "agent离线无法执行作业任务", CreateTime: time.Now()}
							err = db.Create(&jr).Error
						}
					} else {
						m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
							"job_id": v.JobId, "exec": v.Exec, "job_type": "job_exec", "host_id": v.HostId}), true)
						platform_conf.Wch <- map[string]interface{}{"jobShell": kits.MapToJson(map[string]interface{}{"jobShell": m}),
							"job_type": "job_exec", "host_id": v.HostId,
							"msg_time": time.Now().Format("2006-01-02 15:04:05")}
						if err == nil {
							db.Model(&databases.JobExec{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobExec{Status: "running"})
						}
					}
				}
			}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					Log.Error(fmt.Sprint(r))
				}
			}()
			db.Where("cron=? and send_time like ? and status=?", 1,
				strings.Join(h[:len(h)-1], ":")+"%", "pending").Find(&JobFile)
			if len(JobFile) > 0 {
				for _, v := range JobFile {
					if rc.HExists(ctx, platform_conf.OfflineAssetKey, v.HostId).Val() {
						failKey := "job_file_results_fail" + v.JobId
						rc.HSet(ctx, failKey, v.HostId, "fail")
						rc.Expire(ctx, failKey, 8*time.Hour)
						err = db.Transaction(func(tx *gorm.DB) error {
							db.Model(&databases.JobExec{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobExec{Status: "fail"})
							return sqlErr
						})
						if err == nil {
							jr := databases.JobResults{JobId: v.JobId, HostId: v.HostId,
								Results: "agent离线无法执行作业任务", CreateTime: time.Now()}
							err = db.Create(&jr).Error
						}
					} else {
						m := kits.MapToJson(map[string]interface{}{
							"job_id": v.JobId, "dst_path": v.DstPath, "files": strings.Split(v.Files, ","),
							"job_type": "job_file", "host_id": v.HostId})
						platform_conf.Wch <- map[string]interface{}{"jobFile": kits.MapToJson(map[string]interface{}{"jobFile": m}),
							"job_type": "job_file", "host_id": v.HostId,
							"msg_time": time.Now().Format("2006-01-02 15:04:05")}
						if err == nil {
							db.Model(&databases.JobFile{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobFile{Status: "sending"})
						}
					}
				}
			}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					Log.Error(fmt.Sprint(r))
				}
			}()
			db.Where("cron=? and run_time like ? and status=?", 1,
				strings.Join(h[:len(h)-1], ":")+"%", "pending").Find(&JobRun)
			if len(JobRun) > 0 {
				for _, v := range JobRun {
					if rc.HExists(ctx, platform_conf.OfflineAssetKey, v.HostId).Val() {
						err = db.Transaction(func(tx *gorm.DB) error {
							db.Model(&databases.JobRun{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobRun{Status: "fail"})
							return sqlErr
						})
						if err == nil {
							jr := databases.JobResults{JobId: v.JobId, HostId: v.HostId,
								Results: "agent离线无法执行作业任务", CreateTime: time.Now()}
							err = db.Create(&jr).Error
						}
					} else {
						db.Where("script_id=?", v.ScriptId).First(&JobScript)
						m := Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
							"job_id": v.JobId, "dst_path": "/tmp/" + v.ScriptId + "/", "job_type": "job_script",
							"files": []string{JobScript[0].ScriptName}, "script_id": v.ScriptId, "host_id": v.HostId}),
							true)
						platform_conf.Wch <- map[string]interface{}{"jobFile": kits.MapToJson(map[string]interface{}{"jobFile": m}),
							"job_type": "job_script", "host_id": v.HostId,
							"msg_time": time.Now().Format("2006-01-02 15:04:05")}
						if err == nil {
							db.Model(&databases.JobRun{}).Where("job_id=? and host_id=?",
								v.JobId, v.HostId).Updates(databases.JobRun{Status: "sending"})
						}
					}
				}
			}
		}()
	}
}

func CleanAsset() {
	lock := common.SyncMutex{LockKey: "job_clean_asset_lock"}
	//加锁
	if lock.Lock() {
		var (
			sqlErr      error
			err         error
			JobExec     []databases.JobExec
			JobFile     []databases.JobFile
			JobRun      []databases.JobRun
			JobOverview []databases.JobOverview
		)
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		Log.Info("下架设备清除任务开始执行......")
		d := rc.HGetAll(ctx, platform_conf.DiscardAssetKey).Val()
		if len(d) > 0 {
			var ds []string
			for k := range d {
				ds = append(ds, k)
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where("host_id in ?", ds).Delete(&JobExec).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", ds).Delete(&JobFile).Error; err != nil {
					sqlErr = err
				}
				if err = tx.Where("host_id in ?", ds).Delete(&JobRun).Error; err != nil {
					sqlErr = err
				}
				return sqlErr
			})
			if err == nil {
				db.Select("job_id", "job_type").Find(&JobOverview)
				if len(JobOverview) > 0 {
					for _, v := range JobOverview {
						var del bool
						if v.JobType == "job_exec" {
							db.Where("job_id=?", v.JobId).First(&JobExec)
							if len(JobExec) == 0 {
								del = true
							}
						}
						if v.JobType == "job_file" {
							db.Where("job_id=?", v.JobId).First(&JobFile)
							if len(JobFile) == 0 {
								del = true
							}
						}
						if v.JobType == "job_script" {
							db.Where("job_id=?", v.JobId).First(&JobRun)
							if len(JobRun) == 0 {
								del = true
							}
						}
						if del {
							db.Where("job_id=?", v.JobId).Delete(&JobOverview)
						}
					}
				}
			}
		}
	}
}

func CleanJobs() {
	lock := common.SyncMutex{LockKey: "clean_jobs_lock"}
	//加锁
	if lock.Lock() {
		var (
			err          error
			JobExec      []databases.JobExec
			JobRun       []databases.JobRun
			JobFile      []databases.JobFile
			FileContents []databases.FileContents
		)
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		Log.Info("清除运维作业任务开始执行......")
		sf := rc.SMembers(ctx, job_conf.SendFileJobKey).Val()
		if len(sf) > 0 {
			for _, v := range sf {
				var del bool
				db.Where("job_id=?", v).Find(&JobFile)
				if len(JobFile) == 0 {
					del = true
				} else {
					db.Where("job_id=? and status in ?", v, []string{"pending", "sending"}).Find(&JobFile)
					if len(JobFile) == 0 {
						del = true
					}
				}
				if del {
					err = db.Where("job_id=?", v).Delete(&FileContents).Error
					if err == nil {
						rc.SRem(ctx, job_conf.SendFileJobKey, v)
					}
				}
			}
		}
		db.Where("status=?", "running").Find(&JobExec)
		if len(JobExec) > 0 {
			for _, v := range JobExec {
				if carbon.Time2Carbon(v.RunTime).DiffInMinutes(carbon.Now()) > 30 {
					failKey := "job_exec_results_fail" + v.JobId
					rc.HSet(ctx, failKey, v.HostId, "fail")
					platform_conf.Ech <- map[string]interface{}{"host_id": v.HostId,
						"job_id":   v.JobId,
						"job_type": "job_exec",
						"stdout":   "",
						"stderr":   "",
						"msg_time": datetime.GetNowDateTime(),
						"status":   false,
					}
				}
			}
		}
		db.Where("status=?", "running").Find(&JobRun)
		if len(JobRun) > 0 {
			for _, v := range JobRun {
				if carbon.Time2Carbon(v.RunTime).DiffInMinutes(carbon.Now()) > 30 {
					failKey := "job_exec_results_fail" + v.JobId
					rc.HSet(ctx, failKey, v.HostId, "fail")
					platform_conf.Ech <- map[string]interface{}{"host_id": v.HostId,
						"job_id":   v.JobId,
						"job_type": "job_script",
						"stdout":   "",
						"stderr":   "",
						"msg_time": datetime.GetNowDateTime(),
						"status":   false,
					}
				}
			}
		}
		db.Where("status=?", "running").Find(&JobFile)
		if len(JobFile) > 0 {
			for _, v := range JobFile {
				if carbon.Time2Carbon(v.SendTime).DiffInMinutes(carbon.Now()) > 30 {
					failKey := "job_file_results_fail" + v.JobId
					rc.HSet(ctx, failKey, v.HostId, "fail")
					platform_conf.Ech <- map[string]interface{}{"host_id": v.HostId,
						"job_id":   v.JobId,
						"job_type": "job_file",
						"file":     "",
						"message":  "",
						"msg_time": datetime.GetNowDateTime(),
						"status":   false,
					}
				}
			}
		}
	}
}
