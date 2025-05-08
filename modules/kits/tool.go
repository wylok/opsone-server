package kits

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/random"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/modules/databases"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func MD5(str string) string {
	// 字符串Md5值
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func RandString(len int) string {
	// 随机字符串生成
	var uuid string
	if len == 0 {
		uuid, _ = random.UUIdV4()
	} else {
		uuid = random.RandNumeralOrLetter(len)
	}
	return uuid
}
func MapToJson(param map[string]interface{}) string {
	//字典转为Json
	dataType, _ := json.Marshal(param)
	return string(dataType)
}
func StringToMap(content string) map[string]interface{} {
	//字符串转为字典
	resMap := map[string]interface{}{}
	_ = json.Unmarshal([]byte(content), &resMap)
	return resMap
}
func TimeFormat(SubTime float64) string {
	// 秒
	var result string
	if SubTime < 60 {
		result = cast.ToString(SubTime) + "秒"
	}
	// 分钟
	if SubTime >= 60 && SubTime < 60*60 {
		result = cast.ToString(fmt.Sprintf("%.1f", math.Floor(SubTime/60.0))) + "分钟"
	}
	// 小时
	if SubTime >= 60*60 && SubTime < 60*60*24 {
		result = cast.ToString(fmt.Sprintf("%.1f", math.Floor(SubTime/(60.0*60.0)))) + "小时"
	}
	// 天
	if SubTime >= 60*60*24 {
		result = cast.ToString(fmt.Sprintf("%.1f", math.Floor(SubTime/(60.0*60.0*24)))) + "天"
	}
	return result
}
func GetFunctionName(i interface{}, sep string) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	fields := strings.Split(fn, sep)
	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return fn
}

func FormListFormat(data []string) []string {
	if len(data) == 1 {
		if strings.Contains(data[0], ",") {
			data = strings.Split(data[0], ",")
		}
	}
	return data
}
func FormatMonitorValue(value float64, unit string) float64 {
	if strings.HasPrefix(unit, "MB") || strings.HasPrefix(unit, "Mbit") {
		value = value / 1000 / 1000
	}
	if strings.HasPrefix(unit, "GB") || strings.HasPrefix(unit, "Gbit") {
		value = value / 1000 / 1000 / 1000
	}
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", value), 64)
	return value
}
func SshAudit(userId, hostId, assetType, fileName string) string {
	var auditId string
	if userId != "" && hostId != "" {
		auditId = RandString(12)
		_ = db.Transaction(func(tx *gorm.DB) error {
			sa := databases.SshAudit{AuditId: auditId, AssetId: hostId, AssetType: assetType, UserId: userId,
				FileName: fileName, StartTime: time.Now()}
			sqlErr := tx.Create(&sa).Error
			sc := databases.SshContent{AuditId: auditId, ShellContent: ""}
			sqlErr = tx.Create(&sc).Error
			return sqlErr
		})
	}
	return auditId
}
