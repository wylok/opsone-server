package platform

import (
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/netutil"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/duke-git/lancet/v2/system"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"inner/conf/platform_conf"
	"inner/modules/databases"
	"inner/modules/kits"
	"net"
	"strconv"
	"strings"
	"time"
)

func SendMsg(data map[string]interface{}) {
	var (
		err     error
		wsc     *websocket.Conn
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
	HostId := data["host_id"].(string)
	w, ok := platform_conf.WscPools.Load(HostId)
	if ok && w != nil {
		wsc = w.(*websocket.Conn)
		if wsc != nil {
			_, ok = data["monitor"]
			if ok {
				err = wsc.WriteMessage(1, []byte(cast.ToString(data["monitor"])))
				if err != nil {
					Log.Error(err)
				}
			}
			_, ok = data["jobShell"]
			if ok {
				err = wsc.WriteMessage(1, []byte(cast.ToString(data["jobShell"])))
				if err == nil {
					Log.Info("成功发送命令执行作业到主机" + HostId)
				} else {
					Log.Error(err)
				}
			}
			_, ok = data["jobFile"]
			if ok {
				var (
					dd = map[string]interface{}{}
					da = map[string]interface{}{}
				)
				err = json.Unmarshal([]byte(cast.ToString(data["jobFile"])), &dd)
				if err == nil {
					sa, _ := Encrypt.DecryptString(cast.ToString(dd["jobFile"]), true)
					err = json.Unmarshal(sa, &da)
					if err == nil {
						files := map[string]string{}
						if cast.ToString(data["job_type"]) == "job_script" {
							if rc.Exists(ctx, "job_send_file_"+cast.ToString(da["script_id"])).Val() == 1 {
								files = rc.HGetAll(ctx, "job_send_file_"+cast.ToString(da["script_id"])).Val()
							}
						} else {
							if rc.Exists(ctx, "job_send_file_"+cast.ToString(da["job_id"])).Val() == 1 {
								files = rc.HGetAll(ctx, "job_send_file_"+cast.ToString(da["job_id"])).Val()
							}
						}
						if len(files) > 0 {
							filenames := cast.ToSlice(da["files"])
							delete(da, "files")
							for _, f := range filenames {
								_, ok = files[cast.ToString(f)]
								if ok {
									da["file_name"] = cast.ToString(f)
									da["file_content"] = files[cast.ToString(f)]
									m := kits.MapToJson(map[string]interface{}{
										"jobFile": Encrypt.EncryptString(kits.MapToJson(da), true)})
									//传输文件
									err = wsc.WriteMessage(websocket.TextMessage, []byte(m))
									if err == nil {
										Log.Info(f.(string) + "文件成功发送到主机" + HostId)
									} else {
										break
									}
								}
							}
						} else {
							s, err := json.Marshal(map[string]interface{}{"host_id": HostId,
								"job_id":   cast.ToString(da["job_id"]),
								"job_type": cast.ToString(data["job_type"]),
								"file":     "",
								"message":  "未找到相关文件信息",
								"status":   false,
								"msg_time": time.Now().Format("2006-01-02 15:04:05")})
							if err == nil {
								platform_conf.Fch <- map[string]interface{}{"jobFile": Encrypt.EncryptString(string(s), true)}
							}
							Log.Error(errors.New(""))
						}
					}
				}
			}
		}
	}
}

func LocalWscSend() {
	//本机wsc消息直接下发
	for {
		data := <-platform_conf.Wch
		if data != nil {
			_, ok := data["host_id"]
			if ok && data["host_id"] != nil {
				_, ok = platform_conf.WscPools.Load(cast.ToString(data["host_id"]))
				if ok {
					go SendMsg(data)
				} else {
					rc.HSet(ctx, platform_conf.WscSend, platform_conf.Uuid+"_"+kits.RandString(6), kits.MapToJson(data))
				}
			}
		}
	}
}

func PoolsWscSend() {
	//获取wsc池本机消息下发
	for {
		for k, v := range rc.HGetAll(ctx, platform_conf.WscSend).Val() {
			data := kits.StringToMap(v)
			_, ok := data["host_id"]
			if ok {
				_, ok = platform_conf.WscPools.Load(cast.ToString(data["host_id"]))
				if ok {
					rc.HDel(ctx, platform_conf.WscSend, k)
					go SendMsg(data)
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func HeartBeatHandle() {
	//agent心跳数据处理
	for {
		hd := <-platform_conf.Hch
		go func(hd platform_conf.HeartbeatData) {
			var (
				AgentAlive []databases.AgentAlive
				sf         string
				upgrade    bool
				discard    bool
				Encrypt    = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
			)
			defer func() {
				if r := recover(); r != nil {
					Log.Error(errors.New(fmt.Sprint(r)))
				}
			}()
			if rc.HExists(ctx, platform_conf.DiscardAssetKey, hd.HostId).Val() {
				discard = true
				Log.Info("设备(" + hd.HostName + ")" + "进行下架处理")
			}
			if discard {
				sf = Encrypt.EncryptString(kits.MapToJson(
					map[string]interface{}{
						"AgentVersion":      "",
						"AssetAgentRun":     0,
						"MonitorAgentRun":   0,
						"HeartBeatInterval": 30,
						"AssetInterval":     15,
						"MonitorInterval":   60,
						"Upgrade":           false,
						"Uninstall":         true}),
					true)
			} else {
				if rc.HExists(ctx, platform_conf.OfflineAssetKey, hd.HostId).Val() {
					db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
						map[string]interface{}{"offline_time": 0})
				}
				if platform_conf.AgentVersion != hd.AgentVersion {
					if rc.HExists(ctx, platform_conf.AgentUpgradeKey, hd.HostId).Val() {
						upgrade = true
					} else {
						if len(rc.HGetAll(ctx, platform_conf.AgentUpgradeKey).Val()) <= 20 {
							upgrade = true
							rc.HSet(ctx, platform_conf.AgentUpgradeKey, hd.HostId, hd.AgentVersion)
						}
					}
					if upgrade {
						Log.Info("服务器(" + hd.HostName + ")升级agent版本到" + platform_conf.AgentVersion)
					}
				}
				if platform_conf.AgentVersion == hd.AgentVersion {
					if rc.HExists(ctx, platform_conf.AgentUpgradeKey, hd.HostId).Val() {
						rc.HDel(ctx, platform_conf.AgentUpgradeKey, hd.HostId)
						db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
							map[string]interface{}{"agent_version": hd.AgentVersion})
						Log.Info("服务器(" + hd.HostName + ")完成agent版本升级")
					}
				}
				//记录心跳检测
				aliveCountKey := "agent_alive_count_" + hd.HostId
				rc.Incr(ctx, aliveCountKey)
				i, _ := strconv.Atoi(rc.Get(ctx, aliveCountKey).Val())
				if i >= 5 {
					rc.Del(ctx, aliveCountKey)
					db.Where("host_id=?", hd.HostId).Find(&AgentAlive)
					if len(AgentAlive) == 0 {
						aa := databases.AgentAlive{HostId: hd.HostId,
							AgentVersion: hd.AgentVersion, OfflineTime: 0, ClamAv: "None", ClamRun: "未知"}
						db.Create(&aa)
					} else {
						db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
							databases.AgentAlive{AgentVersion: hd.AgentVersion})
					}
					ClamAliveKey := "clamAv_alive_" + hd.HostId
					if hd.ClamAv == "clamAv" {
						if rc.Exists(ctx, ClamAliveKey).Val() == 0 {
							db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
								map[string]interface{}{"clamAv": hd.ClamAv})
						}
						rc.Set(ctx, ClamAliveKey, hd.ClamAv, 0)
						if hd.ClamRun != "" {
							db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
								map[string]interface{}{"clamRun": hd.ClamRun})
						}
					} else {
						if rc.Exists(ctx, ClamAliveKey).Val() == 1 {
							db.Model(&AgentAlive).Where("host_id=?", hd.HostId).Updates(
								map[string]interface{}{"clamAv": "None", "clamRun": "未知"})
							rc.Del(ctx, ClamAliveKey)
						}
					}
				}
				rc.HSet(ctx, platform_conf.AgentAliveKey, hd.HostId, time.Now())
				sf = Encrypt.EncryptString(kits.MapToJson(map[string]interface{}{
					"AgentVersion":      platform_conf.AgentVersion,
					"AssetAgentRun":     platform_conf.AssetAgentRun,
					"MonitorAgentRun":   platform_conf.MonitorAgentRun,
					"HeartBeatInterval": platform_conf.HeartBeatInterval,
					"AssetInterval":     platform_conf.AssetInterval,
					"MonitorInterval":   platform_conf.MonitorInterval,
					"Upgrade":           upgrade,
					"Uninstall":         false}),
					true)
			}
			_ = hd.Ws.WriteMessage(1, []byte(kits.MapToJson(map[string]interface{}{"heartbeat": sf})))
		}(hd)
	}
}

func RsyncAgentConf() {
	for {
		var AgentConf []databases.AgentConf
		db.Find(&AgentConf)
		if len(AgentConf) > 0 {
			platform_conf.AssetAgentRun = AgentConf[0].AssetAgentRun
			platform_conf.MonitorAgentRun = AgentConf[0].MonitorAgentRun
			platform_conf.AssetInterval = int(AgentConf[0].AssetInterval)
			platform_conf.HeartBeatInterval = int(AgentConf[0].HeartBeatInterval)
			platform_conf.MonitorInterval = int(AgentConf[0].MonitorInterval)
		}
		time.Sleep(15 * time.Second)
	}
}
func CleanAuditFile() {
	//定期清理ssh审计录像临时文件
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					Log.Error(err)
				}
			}()
			filePath := platform_conf.RootPath + "/opsone/static/webshell"
			if fileutil.IsExist(filePath) {
				files, err := fileutil.ListFileNames(filePath)
				if err == nil {
					for _, f := range files {
						if strings.Contains(f, "index-") || strings.Contains(f, ".cast") {
							mTime, err := fileutil.MTime(filePath + "/" + f)
							if err == nil {
								if time.Now().Add(-30*time.Minute).Unix() > mTime {
									_ = fileutil.RemoveFile(filePath + "/" + f)
								}
							}
						}
					}
				}
			}
		}()
		time.Sleep(30 * time.Minute)
	}
}
func GetRemoteIp() {
	path := platform_conf.RootPath + "/opsone/static"
	p := path + "/config/config.ini"
	f := path + "/agent/install.sh"
	if fileutil.IsExist(p) {
		remoteIp, err := fileutil.ReadFileToString(p)
		remoteIp = strings.TrimSpace(remoteIp)
		if netutil.IsInternalIP(net.ParseIP(remoteIp)) || netutil.IsPublicIP(net.ParseIP(remoteIp)) {
			if platform_conf.RemoteAddr != remoteIp {
				_, _, err = system.ExecCommand("sed -i 's/" + platform_conf.RemoteAddr + "/" + remoteIp + "/g' " + f)
				_, _, err = system.ExecCommand("sed -i 's/<remote_ip>/" + remoteIp + "/g' " + f)
				platform_conf.RemoteAddr = remoteIp
			}
		}
		if err != nil {
			Log.Error(err)
		}
	} else {
		if platform_conf.RemoteAddr != "" {
			_ = fileutil.WriteStringToFile(p, platform_conf.RemoteAddr, false)
		}
	}
}
