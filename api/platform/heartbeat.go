package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"inner/conf/platform_conf"
	"inner/modules/kits"
	"time"
)

// @Tags 平台管理
// @Summary Agent心跳接口
// @Router /api/v1/heartbeat/ws/ [get]
func Heartbeat(c *gin.Context) {
	//升级get请求为webSocket协议
	var (
		ws, err = UpGrader.Upgrade(c.Writer, c.Request, nil)
		Encrypt = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
		hd      = platform_conf.HeartbeatData{}
	)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
			Log.Error(err)
		}
		if hd.HostId != "" {
			platform_conf.WscPools.Delete(hd.HostId)
		}
		_ = ws.Close()
	}()
	if err == nil && ws != nil {
		for {
			_, msg, err := ws.ReadMessage()
			if err == nil {
				data := kits.StringToMap(string(msg))
				if len(data) > 0 {
					nowTime := time.Now().Format("2006-01-02 15:04:05")
					for k, v := range data {
						//数据解密
						s, err := Encrypt.DecryptString(cast.ToString(v), true)
						if err == nil {
							switch k {
							case "heartbeat":
								// 心跳信息写入消息队列
								err = json.Unmarshal(s, &hd)
								if err == nil {
									hd.Ws = ws
									platform_conf.WscPools.Store(hd.HostId, ws)
									platform_conf.Hch <- hd
								} else {
									Log.Error(err)
								}
							case "cmdb":
								// 资产信息写入消息队列
								platform_conf.Cch <- map[string]interface{}{"cmdb": string(s),
									"msg_time": nowTime}
							case "monitor":
								//发送监控数据到消息队列
								platform_conf.Mch <- map[string]interface{}{"monitor": string(s),
									"msg_time": nowTime}
							case "jobShell":
								// 作业任务结果写入消息队列
								js := kits.StringToMap(string(s))
								if js["job_id"] != nil {
									js["msg_time"] = nowTime
									platform_conf.Ech <- js
								}
							case "jobFile":
								// 文件传输结果写入消息队列
								jf := kits.StringToMap(string(s))
								if jf["job_id"] != nil {
									jf["msg_time"] = nowTime
									platform_conf.Fch <- jf
								}
							}
						}
					}
				}
			} else {
				break
			}
		}
	}
}
