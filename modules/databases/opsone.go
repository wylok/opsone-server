package databases

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"inner/conf/platform_conf"
	"strings"
	"time"
)

var (
	DB, err = MysqlConnect()
)

func MysqlConnect() (*gorm.DB, error) {
	// 获取数据库配置信息
	var (
		err error
		db  *gorm.DB
	)
	for {
		db, err = func() (*gorm.DB, error) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(errors.New(fmt.Sprint(r)))
				}
			}()
			var d *gorm.DB
			database := "opsone"
			cf := platform_conf.Setting()
			//open a db connection
			d, err = gorm.Open(mysql.Open(strings.Replace(cf.SqlConfig, "<db>", "mysql", 1)), &gorm.Config{})
			if err == nil && d != nil {
				d, err = gorm.Open(mysql.Open(strings.Replace(cf.SqlConfig, "<db>", database, 1)), &gorm.Config{})
				if err == nil && d != nil {
					CDB, Err := d.DB()
					if Err == nil {
						CDB.SetMaxIdleConns(15)
						CDB.SetMaxOpenConns(50)
						CDB.SetConnMaxLifetime(5 * time.Minute)
					}
				} else {
					d.Exec("create database " + database)
				}
			}
			return d, err
		}()
		if db != nil && err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	return db, err
}
func init() {
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Token{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Rules{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Roles{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&RoleGroup{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Privileges{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Permission{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Department{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Business{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&DepartmentBusiness{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&BusinessUser{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Tenants{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Audit{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Users{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&CloudKeys{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&CloudOss{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&CloudServer{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&CmdbPartition{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetIdc{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetServer{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetNet{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetDisk{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetExtend{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetGroups{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&GroupServer{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetUnder{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetSwitch{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetSwitchVlan{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetSwitchPort{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetSwitchPool{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetServerPool{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&SshKey{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AssetSwitchRelation{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JumpServerKey{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorKeys{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorMetrics{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorRules{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorProcess{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AlarmChannel{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AlarmHistory{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AlarmSend{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AlarmStages{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorGroups{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MonitorJobs{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&CustomMetrics{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&GroupMetrics{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobOverview{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobExec{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobFile{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobScript{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&ScriptContents{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobRun{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&FileContents{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&JobResults{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Msg{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&MsgContent{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AgentConf{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&AgentAlive{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&WorkOrder{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&WorkOrderApprove{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&WorkOrderFlow{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&WorkOrderType{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&GroupProcess{})
	err = DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&K8sCluster{})
}
