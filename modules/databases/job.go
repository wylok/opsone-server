package databases

import (
	"time"
)

type JobOverview struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	JobId      string    `gorm:"column:job_id;type:varchar(100);uniqueIndex" json:"job_id"`
	JobType    string    `gorm:"column:job_type;type:enum('job_exec','job_file','job_script')" json:"job_type"`
	Cron       int       `gorm:"column:cron;type:tinyint(1)" json:"cron"`
	Contents   string    `gorm:"column:contents;type:varchar(500)" json:"contents"`
	Counts     int64     `gorm:"column:counts;type:int(8)" json:"counts"`
	Success    int64     `gorm:"column:success;type:int(8)" json:"success"`
	Fail       int64     `gorm:"column:fail;type:int(8)" json:"fail"`
	UserId     string    `gorm:"column:user_id;type:varchar(100);index" json:"user_id"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (JobOverview) TableName() string {
	return "job_overview"
}

type JobExec struct {
	Id      uint64    `gorm:"primary_key" json:"id"`
	JobId   string    `gorm:"column:job_id;type:varchar(100);uniqueIndex:job_host_id" json:"job_id"`
	HostId  string    `gorm:"column:host_id;type:varchar(100);uniqueIndex:job_host_id" json:"host_id"`
	Exec    string    `gorm:"column:exec;type:varchar(500)" json:"exec"`
	Cron    int       `gorm:"column:cron;type:tinyint(1)" json:"cron"`
	RunTime time.Time `gorm:"column:run_time;type:datetime" json:"run_time"`
	Status  string    `gorm:"column:status;type:enum('pending','running','completed','fail')" json:"status"`
}

func (JobExec) TableName() string {
	return "job_exec"
}

type JobFile struct {
	Id       uint64    `gorm:"primary_key" json:"id"`
	JobId    string    `gorm:"column:job_id;type:varchar(100);uniqueIndex:job_host_id" json:"job_id"`
	HostId   string    `gorm:"column:host_id;type:varchar(100);uniqueIndex:job_host_id" json:"host_id"`
	DstPath  string    `gorm:"column:dst_path;type:varchar(200)" json:"dst_path"`
	Files    string    `gorm:"column:files;type:varchar(200)" json:"files"`
	Cron     int       `gorm:"column:cron;type:tinyint(1)" json:"cron"`
	SendTime time.Time `gorm:"column:send_time;type:datetime" json:"send_time"`
	Status   string    `gorm:"column:status;type:enum('pending','sending','completed','fail')" json:"status"`
}

func (JobFile) TableName() string {
	return "job_file"
}

type FileContents struct {
	Id          uint64 `gorm:"primary_key" json:"id"`
	JobId       string `gorm:"column:job_id;type:varchar(100);uniqueIndex:job_file_name" json:"job_id"`
	FileName    string `gorm:"column:file_name;type:varchar(100);uniqueIndex:job_file_name" json:"file_name"`
	FileContent []byte `gorm:"column:file_content;type:longblob" json:"file_content"`
}

func (FileContents) TableName() string {
	return "file_contents"
}

type JobScript struct {
	Id            uint64    `gorm:"primary_key" json:"id"`
	ScriptId      string    `gorm:"column:script_id;type:varchar(100);uniqueIndex" json:"script_id"`
	ScriptType    string    `gorm:"column:script_type;type:varchar(100);index" json:"script_type"`
	ScriptName    string    `gorm:"column:script_name;type:varchar(100);index" json:"script_name"`
	ScriptDesc    string    `gorm:"column:script_desc;type:varchar(100)" json:"script_desc"`
	ScriptPurpose string    `gorm:"column:script_purpose;type:varchar(100)" json:"script_purpose"`
	UserId        string    `gorm:"column:user_id;type:varchar(100)" json:"user_id"`
	CreateTime    time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (JobScript) TableName() string {
	return "job_script"
}

type ScriptContents struct {
	Id            uint64 `gorm:"primary_key" json:"id"`
	ScriptId      string `gorm:"column:script_id;type:varchar(100);uniqueIndex:script_id_name" json:"script_id"`
	ScriptName    string `gorm:"column:script_name;type:varchar(100);uniqueIndex:script_id_name" json:"script_name"`
	ScriptContent string `gorm:"column:script_content;type:text" json:"script_content"`
}

func (ScriptContents) TableName() string {
	return "script_contents"
}

type JobRun struct {
	Id       uint64    `gorm:"primary_key" json:"id"`
	JobId    string    `gorm:"column:job_id;type:varchar(100);uniqueIndex:job_host_id" json:"job_id"`
	HostId   string    `gorm:"column:host_id;type:varchar(100);uniqueIndex:job_host_id" json:"host_id"`
	ScriptId string    `gorm:"column:script_id;type:varchar(100);index" json:"script_id"`
	Cron     int       `gorm:"column:cron;type:tinyint(1)" json:"cron"`
	RunTime  time.Time `gorm:"column:run_time;type:datetime" json:"run_time"`
	Status   string    `gorm:"column:status;type:enum('pending','running','completed','fail')" json:"status"`
}

func (JobRun) TableName() string {
	return "job_run"
}

type JobResults struct {
	Id         uint64    `gorm:"primary_key" json:"id"`
	JobId      string    `gorm:"column:job_id;type:varchar(100);uniqueIndex:job_host_id" json:"job_id"`
	HostId     string    `gorm:"column:host_id;type:varchar(100);uniqueIndex:job_host_id" json:"host_id"`
	Results    string    `gorm:"column:results;type:text" json:"results"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
}

func (JobResults) TableName() string {
	return "job_results"
}
