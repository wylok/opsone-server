package cmdb

import (
	"fmt"
	"github.com/duke-git/lancet/convertor"
	"github.com/duke-git/lancet/netutil"
	"github.com/pkg/errors"
	"github.com/shenbowei/switch-ssh-go"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"inner/conf/cmdb_conf"
	"inner/conf/msg_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"net"
	"strconv"
	"strings"
	"time"
)

func CleanOwnership() {
	lock := common.SyncMutex{LockKey: "cmdb_clean_ownership_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				Log.Info("清除关联部门信息任务开始执行......")
				var (
					sqlErr        error
					AssetUnder    []databases.AssetUnder
					AssetServer   []databases.AssetServer
					CmdbPartition []databases.CmdbPartition
					err           error
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
				//清除关联部门信息
				departmentId := rc.SPop(ctx, platform_conf.DepartmentDeleteKey).Val()
				if departmentId != "" {
					db.Where("department_id=?", departmentId).Find(&AssetUnder)
					if len(AssetUnder) > 0 {
						err = db.Transaction(func(tx *gorm.DB) error {
							for _, v := range AssetUnder {
								if v.AssetType == "server" {
									if err = tx.Model(&AssetServer).Where("host_id=?", v.AssetId).Updates(
										databases.AssetServer{AssetStatus: "available"}).Error; err != nil {
										sqlErr = err
									}
								}
								if err = tx.Model(&CmdbPartition).Where("object_type=? and object_id=?",
									v.AssetType, v.AssetId).Updates(databases.CmdbPartition{DepartmentId: "None"}).Error; err != nil {
									sqlErr = err
								}
							}
							if err = tx.Delete(&AssetUnder).Error; err != nil {
								sqlErr = err
							}
							return sqlErr
						})
					}
				}
				//清除关联业务组信息
				businessId := rc.SPop(ctx, platform_conf.BusinessDeleteKey).Val()
				if businessId != "" {
					db.Where("business_id=?", businessId).Find(&AssetUnder)
					if len(AssetUnder) > 0 {
						err = db.Transaction(func(tx *gorm.DB) error {
							for _, v := range AssetUnder {
								if v.AssetType == "server" {
									if err = tx.Model(&AssetServer).Where("host_id=?", v.AssetId).Updates(
										databases.AssetServer{AssetStatus: "available"}).Error; err != nil {
										sqlErr = err
									}
								}
								if err = tx.Model(&CmdbPartition).Where("object_type=? and object_id=?",
									v.AssetType, v.AssetId).Updates(databases.CmdbPartition{DepartmentId: "None"}).Error; err != nil {
									sqlErr = err
								}
							}
							tx.Model(&AssetUnder).Updates(databases.AssetUnder{BusinessId: "None"})
							return sqlErr
						})
					}
				}
			}()
		}
		time.Sleep(1 * time.Minute)
	}
}

func OverViewCmdb() {
	lock := common.SyncMutex{LockKey: "cmdb_overview_lock"}
	for {
		//加锁
		if lock.Lock() {
			func() {
				Log.Info("资产总览数据抓取任务开始执行......")
				var (
					err         error
					JobCount    int64
					HostCount   int64
					SwitchCount int64
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
				db.Model(databases.AssetServer{}).Count(&HostCount)
				rc.HSet(ctx, platform_conf.OverViewKey, "server_count", HostCount)
				db.Model(databases.JobOverview{}).Count(&JobCount)
				rc.HSet(ctx, platform_conf.OverViewKey, "jobs", JobCount)
				db.Model(databases.AssetSwitch{}).Where("switch_id != ?", "none").Count(&SwitchCount)
				rc.HSet(ctx, platform_conf.OverViewKey, "switch_count", SwitchCount)
			}()
		}
		time.Sleep(1 * time.Minute)
	}
}

func DiscardAssets() {
	var err error
	lock := common.SyncMutex{LockKey: "cmdb_discard_asset_lock"}
	//加锁
	if lock.Lock() {
		Log.Info("清除无效资源任务开始执行......")
		defer func() {
			lock.UnLock(true)
		}()
		//清除无效交换机信息
		go func() {
			var (
				AssetSwitchPool     []databases.AssetSwitchPool
				AssetSwitch         []databases.AssetSwitch
				AssetSwitchPort     []databases.AssetSwitchPort
				AssetSwitchVlan     []databases.AssetSwitchVlan
				AssetSwitchRelation []databases.AssetSwitchRelation
				ipKey               = "platform_asset_switch_ips"
			)
			db.Find(&AssetSwitchPool)
			if len(AssetSwitchPool) > 0 {
				rc.Del(ctx, ipKey)
				for _, d := range AssetSwitchPool {
					SIps := strings.Split(d.StartIp, ".")
					EIps := strings.Split(d.EndIp, ".")
					if len(SIps) >= 1 && len(EIps) >= 1 {
						SIp := strings.Join(SIps[:len(SIps)-1], ".")
						i, _ := strconv.Atoi(SIps[len(SIps)-1])
						e, _ := strconv.Atoi(EIps[len(EIps)-1])
						var s strings.Builder
						s.WriteString(SIp)
						s.WriteString(".")
						for i <= e {
							rc.HSet(ctx, ipKey, s.String()+strconv.Itoa(i), "")
							i++
						}
					}
				}
				var switchIds []string
				db.Select("switch_id", "switch_ip").Find(&AssetSwitch)
				if len(AssetSwitch) > 0 {
					for _, v := range AssetSwitch {
						if !rc.HExists(ctx, ipKey, v.SwitchIp).Val() {
							switchIds = append(switchIds, v.SwitchId)
						}
					}
					if len(switchIds) > 0 {
						err = db.Transaction(func(tx *gorm.DB) error {
							err = db.Where("switch_id in ?", switchIds).Delete(&AssetSwitch).Error
							err = db.Where("switch_id in ?", switchIds).Delete(&AssetSwitchPort).Error
							err = db.Where("switch_id in ?", switchIds).Delete(&AssetSwitchVlan).Error
							err = db.Where("switch_id in ?", switchIds).Delete(&AssetSwitchRelation).Error
							return err
						})
					}
				}
			}
		}()
		//清除无效资源信息
		go func() {
			var (
				err           error
				sqlErr        error
				AssetServer   []databases.AssetServer
				AssetNet      []databases.AssetNet
				AssetDisk     []databases.AssetDisk
				AssetExtend   []databases.AssetExtend
				AgentAlive    []databases.AgentAlive
				AssetUnder    []databases.AssetUnder
				GroupServer   []databases.GroupServer
				CmdbPartition []databases.CmdbPartition
			)
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
			}()
			d := rc.HGetAll(ctx, platform_conf.DiscardAssetKey).Val()
			if len(d) > 0 {
				var hostIds []string
				for hostId := range d {
					hostIds = append(hostIds, hostId)
				}
				err = db.Transaction(func(tx *gorm.DB) error {
					if err = tx.Where("host_id in ?", hostIds).Delete(&AssetServer).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&AssetNet).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&AssetDisk).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&AssetExtend).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("asset_id in ?", hostIds).Delete(&AssetUnder).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&GroupServer).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("object_id in ? and object_type=?", hostIds, "server").Delete(&CmdbPartition).Error; err != nil {
						sqlErr = err
					}
					if err = tx.Where("host_id in ?", hostIds).Delete(&AgentAlive).Error; err != nil {
						sqlErr = err
					}
					return sqlErr
				})
				if err == nil && len(hostIds) > 0 {
					for _, hostId := range hostIds {
						rc.HDel(ctx, platform_conf.AgentAliveKey, hostId)
						rc.HDel(ctx, platform_conf.HostCpuCoreKey, hostId)
						rc.HDel(ctx, platform_conf.OfflineAssetKey, hostId)
					}
				}
			}
		}()
	}
}

func HandleCmdb() {
	for {
		data := <-platform_conf.Cch
		if data != nil {
			go func(data map[string]interface{}) {
				var (
					err           error
					sqlErr        error
					SendMsg       bool
					DiskSize      uint64
					AssetStatus   = "available"
					AssetServer   []databases.AssetServer
					AssetNet      []databases.AssetNet
					AssetDisk     []databases.AssetDisk
					AssetExtend   []databases.AssetExtend
					GroupServer   []databases.GroupServer
					CmdbPartition []databases.CmdbPartition
				)
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprint(r))
					}
					if err != nil {
						Log.Error(err)
					}
				}()
				jd := kits.StringToMap(cast.ToString(data["cmdb"]))
				_, ok1 := jd["hardware"]
				_, ok2 := jd["system"]
				_, ok3 := jd["ipmi"]
				if ok1 && ok2 && ok3 {
					HardWare := jd["hardware"].(map[string]interface{})
					System := jd["system"].(map[string]interface{})
					Ipmi := jd["ipmi"].(map[string]interface{})
					hostId := cast.ToString(System["host_id"])
					hostName := cast.ToString(System["host_name"])
					err = db.Transaction(func(tx *gorm.DB) error {
						for device, value := range HardWare["disk"].(map[string]interface{}) {
							v := value.(map[string]interface{})
							if strings.Contains(device, "/dev/") {
								DiskSize = uint64(v["size"].(float64)) + DiskSize
							}
						}
						InternetIp := net.ParseIP(cast.ToString(System["internet_ip"]))
						rc.HSet(ctx, platform_conf.HostCpuCoreKey, hostId, HardWare["cpu_core"])
						db.Where("host_id = ?", hostId).First(&AssetServer)
						if len(AssetServer) == 0 {
							// 新主机信息写入数据库
							hostType := "physical"
							for k, v := range cmdb_conf.HostType {
								if strings.Contains(cast.ToString(HardWare["product_name"]), k) {
									hostType = v
									break
								}
							}
							db.Where("host_id=?", hostId).First(&GroupServer)
							if len(GroupServer) > 0 {
								AssetStatus = "assigned"
							}
							as := databases.AssetServer{HostId: hostId,
								Hostname:        hostName,
								NickName:        hostName,
								Sn:              strings.TrimSpace(cast.ToString(HardWare["serial_number"])),
								ProductName:     cast.ToString(HardWare["product_name"]),
								Manufacturer:    cast.ToString(HardWare["manufacturer"]),
								HostType:        hostType,
								HostTypeCn:      cmdb_conf.HostTypeCn[hostType],
								Cpu:             int(HardWare["cpu_core"].(float64)),
								CpuInfo:         cast.ToString(HardWare["cpu_info"]),
								Memory:          uint64(HardWare["mem_size"].(float64)),
								Disk:            DiskSize,
								Os:              cast.ToString(System["os"]),
								Platform:        cast.ToString(System["platform"]),
								PlatformVersion: cast.ToString(System["platformVersion"]),
								KernelVersion:   cast.ToString(System["kernelVersion"]),
								InternetIp:      cast.ToString(InternetIp),
								PoolId:          0,
								CreateTime:      time.Now(),
								UpdateTime:      time.Now(),
								AssetTag:        "None",
								AssetStatus:     AssetStatus}
							if err = tx.Create(&as).Error; err != nil {
								sqlErr = err
							}
							SendMsg = true
						} else {
							// 主机信息变动修改数据库
							niceName := AssetServer[0].NickName
							poolId := AssetServer[0].PoolId
							if rc.HExists(ctx, platform_conf.AssetPoolIdsKey, hostId).Val() {
								id, err := convertor.ToInt(rc.HGet(ctx, platform_conf.AssetPoolIdsKey, hostId).Val())
								if err == nil {
									poolId = int(id)
								}
							}
							if niceName == AssetServer[0].Hostname || niceName == "none" {
								niceName = hostName
							}
							if err = tx.Model(&AssetServer).Where("host_id = ?",
								hostId).Updates(databases.AssetServer{
								Hostname:        hostName,
								NickName:        niceName,
								Cpu:             int(HardWare["cpu_core"].(float64)),
								Memory:          uint64(HardWare["mem_size"].(float64)),
								Disk:            DiskSize,
								PlatformVersion: cast.ToString(System["platformVersion"]),
								KernelVersion:   cast.ToString(System["kernelVersion"]),
								InternetIp:      cast.ToString(InternetIp),
								PoolId:          poolId,
								UpdateTime:      time.Now(),
							}).Error; err != nil {
								sqlErr = err
							}
						}
						// 主机磁盘信息
						dmd5s := map[string]struct{}{}
						dmd5n := map[string]struct{}{}
						db.Where("host_id = ?", hostId).Find(&AssetDisk)
						if len(AssetDisk) > 0 {
							for _, n := range AssetDisk {
								dmd5s[n.Md5Verify] = struct{}{}
							}
						}
						for device, value := range HardWare["disk"].(map[string]interface{}) {
							if strings.Contains(device, "/dev/") {
								v := value.(map[string]interface{})
								M5 := kits.MD5(hostId + device)
								dmd5n[M5] = struct{}{}
								_, ok := dmd5s[M5]
								if !ok {
									ad := databases.AssetDisk{HostId: hostId, DiskName: device,
										MountPoint: cast.ToString(v["mount_point"]), FsType: cast.ToString(v["fs_type"]),
										DiskSize: uint64(v["size"].(float64)), Md5Verify: M5}
									if err = tx.Create(&ad).Error; err != nil {
										sqlErr = err
									}
								} else {
									//修改磁盘信息变动
									if err = tx.Model(&AssetDisk).Where("md5_verify = ?", M5).Updates(
										databases.AssetDisk{DiskSize: uint64(v["size"].(float64)),
											FsType:     cast.ToString(v["fs_type"]),
											MountPoint: cast.ToString(v["mount_point"])}).Error; err != nil {
										sqlErr = err
									}
								}
							}
						}
						// 删除无效磁盘信息
						if len(dmd5s) > 0 && len(dmd5n) > 0 {
							for k := range dmd5s {
								_, ok := dmd5n[k]
								if ok == false {
									if err = tx.Where("md5_verify = ?", k).Delete(&AssetDisk).Error; err != nil {
										sqlErr = err
									}
								}
							}
						}
						// 主机网络信息
						nmd5s := map[string]struct{}{}
						nmd5n := map[string]struct{}{}
						db.Where("host_id = ?", hostId).Find(&AssetNet)
						if len(AssetNet) > 0 {
							for _, n := range AssetNet {
								nmd5s[n.Md5Verify] = struct{}{}
							}
						}
						// 写入新主机ip信息
						idcId := "None"
						for name, value := range HardWare["net"].(map[string]interface{}) {
							if kits.ExcludeNetName(name, []string{}) {
								v := value.(map[string]interface{})
								if cast.ToString(v["hardwareaddr"]) != "" {
									_, err = net.ParseMAC(cast.ToString(v["hardwareaddr"]))
									if v["addrs"] != nil && err == nil {
										for _, p := range v["addrs"].([]interface{}) {
											var netmask string
											IP, _, err := net.ParseCIDR(cast.ToString(p))
											if err == nil {
												if IP.To4() != nil {
													if netutil.IsPublicIP(IP) {
														rc.Set(ctx, platform_conf.HostWanKey+hostId, IP.String(), 30*time.Minute)
													}
													if IP.String() != "127.0.0.1" {
														if netutil.IsInternalIP(IP) {
															rc.Set(ctx, platform_conf.ServerIpKey+hostId, IP.String(), 30*time.Minute)
														}
														rc.HSet(ctx, platform_conf.IpHostIdKey, IP.String(), hostId)
														if rc.HExists(ctx, platform_conf.IdcIdKey, IP.String()).Val() {
															idcId = rc.HGet(ctx, platform_conf.IdcIdKey, IP.String()).Val()
														}
														n, err := strconv.Atoi(strings.Split(cast.ToString(p), "/")[1])
														if n >= 0 && 32 >= n {
															netmask = kits.LenToSubNetMask(n)
														}
														M5 := kits.MD5(hostId + cast.ToString(v["hardwareaddr"]) + IP.String())
														nmd5n[M5] = struct{}{}
														_, ok := nmd5s[M5]
														if !ok {
															an := databases.AssetNet{HostId: hostId, Name: name,
																Addr: cast.ToString(v["hardwareaddr"]), Ip: IP.String(),
																Netmask: netmask, Md5Verify: M5,
															}
															if err = tx.Create(&an).Error; err != nil {
																sqlErr = err
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
						// 删除无效ip信息
						if len(nmd5s) > 0 && len(nmd5n) > 0 {
							for k := range nmd5s {
								_, ok := nmd5n[k]
								if ok == false {
									if err = tx.Where("md5_verify = ?", k).Delete(&AssetNet).Error; err != nil {
										sqlErr = err
									}
								}
							}
						}
						// 扩展信息
						ip := cast.ToString(Ipmi["ip"])
						if ip == "" {
							ip = "None"
						}
						if rc.HExists(ctx, platform_conf.IdcIdKey, hostId).Val() {
							idcId = rc.HGet(ctx, platform_conf.IdcIdKey, hostId).Val()
						}
						db.Where("host_id = ?", hostId).Find(&AssetExtend)
						if len(AssetExtend) == 0 {
							ae := databases.AssetExtend{HostId: hostId, IdcId: idcId, Ipmi: ip, Cabinet: "", BuyTime: time.Now(),
								ExpiredTime: time.Now().AddDate(3, 0, 0)}
							if err = tx.Create(&ae).Error; err != nil {
								sqlErr = err
							}
						} else {
							if idcId != "None" {
								db.Model(&AssetExtend).Where("host_id = ?", hostId).Updates(databases.AssetExtend{Ipmi: ip, IdcId: idcId})
							}
						}
						// 部门信息
						db.Where("object_id = ?", hostId).Find(&CmdbPartition)
						if len(CmdbPartition) == 0 {
							cp := databases.CmdbPartition{
								ObjectId: hostId, ObjectType: "server", DepartmentId: "None"}
							if err = tx.Create(&cp).Error; err != nil {
								sqlErr = err
							}
						}
						return sqlErr
					})
					if err == nil {
						Log.Info("服务器(" + hostName + ")上报配置信息成功!")
						if SendMsg {
							// 站内信通知
							content := kits.MapToJson(map[string]interface{}{
								"host_name":    hostName,
								"product_name": cast.ToString(HardWare["product_name"]),
								"platform": cast.ToString(System["platform"]) +
									cast.ToString(System["platformVersion"]),
								"cpu":    strconv.Itoa(int(HardWare["cpu_core"].(float64))),
								"memory": uint64(HardWare["mem_size"].(float64)),
								"disk":   DiskSize,
								"ip":     rc.Get(ctx, platform_conf.ServerIpKey+hostId).Val(),
								"time":   time.Now().Format("2006-01-02 15:04:05")})
							if kits.RecordMsg(msg_conf.RMsg{MsgType: "system", Level: "info",
								Content: content, Title: "新服务器:" + hostName}) {
							}
						}
					}
				}
			}(data)
		}
	}
}

func SyncAssets() {
	var (
		err         error
		AssetServer []databases.AssetServer
		AssetGroups []databases.AssetGroups
		GroupServer []databases.GroupServer
	)
	lock := common.SyncMutex{LockKey: "cmdb_sync_asset_lock"}
	//加锁
	if lock.Lock() {
		Log.Info("同步资源信息任务开始执行......")
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		db.Select("host_id", "host_name", "sn").Find(&AssetServer)
		if len(AssetServer) > 0 {
			for _, v := range AssetServer {
				rc.HSet(ctx, platform_conf.ServerNameKey, v.HostId, v.Hostname)
				rc.Set(ctx, platform_conf.ServerSnKey+"_"+v.Sn, v.Sn, 15*time.Minute)
			}
		}
		db.Select("group_id", "group_name").Find(&AssetGroups)
		if len(AssetGroups) > 0 {
			rc.Del(ctx, platform_conf.GroupNameKey)
			for _, v := range AssetGroups {
				rc.HSet(ctx, platform_conf.GroupNameKey, v.GroupId, v.GroupName)
			}
		}
		db.Find(&GroupServer)
		if len(GroupServer) > 0 {
			rc.HDel(ctx, platform_conf.GroupServersKey)
			for _, v := range GroupServer {
				rc.HSet(ctx, platform_conf.GroupServersKey, v.HostId, v.GroupId)
			}
		}
	}
}

func DiscoverSwitch() {
	var (
		err             error
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	lock := common.SyncMutex{LockKey: "cmdb_discover_switch_lock"}
	//加锁
	if lock.Lock() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		if err == nil {
			db.Find(&AssetSwitchPool)
			if len(AssetSwitchPool) > 0 {
				Log.Info("交换机巡检任务开始执行......")
				for _, s := range AssetSwitchPool {
					go func(as databases.AssetSwitchPool) {
						defer func() {
							if r := recover(); r != nil {
								err = errors.New(fmt.Sprint(r))
							}
							if err != nil {
								Log.Error(err)
							}
						}()
						var (
							ips         []string
							c           int64
							st          strings.Builder
							ASwitchPool []databases.AssetSwitchPool
							noPage      = map[string]string{"huawei": ssh.HuaweiNoPage,
								"h3c": ssh.H3cNoPage, "cisco": ssh.CiscoNoPage}
							Show = map[string]string{"huawei": "display", "h3c": "display", "cisco": "show"}
						)
						SIps := strings.Split(as.StartIp, ".")
						EIps := strings.Split(as.EndIp, ".")
						if len(SIps) >= 1 && len(EIps) >= 1 {
							SIp := strings.Join(SIps[:len(SIps)-1], ".")
							i, _ := strconv.Atoi(SIps[len(SIps)-1])
							e, _ := strconv.Atoi(EIps[len(EIps)-1])
							st.WriteString(SIp)
							st.WriteString(".")
							for i <= e {
								ip := st.String() + strconv.Itoa(i)
								ips = append(ips, ip)
								if netutil.IsPingConnected(ip) {
									func(d databases.AssetSwitchPool, ip string) {
										var (
											ASwitch         []databases.AssetSwitch
											ASwitchPort     []databases.AssetSwitchPort
											ASwitchVlan     []databases.AssetSwitchVlan
											ASwitchRelation []databases.AssetSwitchRelation
										)
										db.Where("switch_ip = ?", ip).Find(&ASwitch)
										if len(ASwitch) == 0 {
											as := databases.AssetSwitch{SwitchPoolId: d.Id, SwitchIp: ip, SwitchId: "none", SwitchName: "none",
												SwitchBrand: "none", SwitchVersion: "none", IdcId: as.IdcId, Status: "online", SyncTime: time.Now()}
											err = db.Create(&as).Error
										} else {
											v := ASwitch[0]
											if d.SwitchStatus == "enable" {
												pwd, e := Encrypt.DecryptString(d.SwitchPassword, true)
												if e == nil {
													session, e := ssh.NewSSHSession(d.SwitchUser, string(pwd), v.SwitchIp+":"+strconv.Itoa(d.SwitchPort))
													if e == nil {
														defer session.Close()
														//更新交换机信息
														brand := session.GetSSHBrand()
														if brand != "" {
															session.ClearChannel()
															session.WriteChannel(noPage[brand], Show[brand]+" device manuinfo")
															for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																if strings.Contains(line, "DEVICE_SERIAL_NUMBER") {
																	v.SwitchId = strings.TrimSpace(strings.Split(line, ":")[1])
																	break
																}
															}
															session.ClearChannel()
															session.WriteChannel(Show[brand] + " current-configuration")
															for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																if strings.Contains(line, "sysname") {
																	lines := strings.Split(line, " ")
																	v.SwitchName = strings.TrimSpace(lines[len(lines)-1])
																	break
																}
															}
															session.ClearChannel()
															session.WriteChannel(Show[brand] + " version")
															for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																if strings.Contains(line, "version") && strings.Contains(line, "Release") {
																	lines := strings.Split(strings.Split(line, ",")[0], " ")
																	v.SwitchVersion = strings.TrimSpace(lines[len(lines)-1])
																	break
																}
															}
															err = db.Model(&ASwitch).Where("switch_ip = ?", v.SwitchIp).Updates(
																databases.AssetSwitch{SwitchPoolId: d.Id, SwitchId: v.SwitchId, SwitchBrand: brand,
																	SwitchName: v.SwitchName, SwitchVersion: v.SwitchVersion, IdcId: d.IdcId,
																	Status: "online", SyncTime: time.Now()}).Error
															if err != nil {
																Log.Error(err)
															}
															//更新交换机vlan和端口信息
															if v.SwitchId != "none" {
																maces := map[string]string{}
																session.ClearChannel()
																session.WriteChannel(Show[brand] + " mac-address")
																for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																	if strings.Contains(line, "GE") || strings.Contains(line, "BAG") {
																		if !strings.Contains(line, " 1 ") {
																			lines := strings.Split(line, " ")
																			for _, p := range lines {
																				if strings.Contains(p, "GE") || strings.Contains(p, "BAG") {
																					maces[p] = lines[0]
																					break
																				}
																			}
																		}
																	}
																}
																session.ClearChannel()
																session.WriteChannel(Show[brand] + " interface brief")
																for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																	if strings.Contains(line, " A ") || strings.Contains(line, " T ") {
																		if strings.Contains(line, " UP ") || strings.Contains(line, " DOWN ") {
																			infos := strings.Split(line, " ")
																			if (strings.Contains(infos[0], "GE") || strings.Contains(infos[0], "BAG")) && len(infos) >= 20 {
																				var vlan int
																				i := 20
																				for {
																					if vlan >= 1 || i >= len(infos) {
																						break
																					}
																					vlan, err = strconv.Atoi(infos[i])
																					i = i + 1
																				}
																				if vlan >= 1 && err == nil {
																					stat := "DOWN"
																					if strings.Contains(line, " UP ") {
																						stat = "UP"
																					}
																					portType := "Access"
																					if strings.Contains(line, " T ") {
																						portType = "Trunk"
																					}
																					mac := "none"
																					_, ok := maces[infos[0]]
																					if ok {
																						mac = maces[infos[0]]
																					}
																					db.Where("switch_id = ? and port_name = ?", v.SwitchId, infos[0]).First(&ASwitchPort)
																					if len(ASwitchPort) > 0 {
																						err = db.Model(&ASwitchPort).Updates(databases.AssetSwitchPort{SwitchVlan: uint32(vlan),
																							MacAddress: mac, PortType: portType, PortStat: stat, LastTime: time.Now()}).Error
																					} else {
																						asp := databases.AssetSwitchPort{SwitchId: v.SwitchId, PortName: infos[0], PortType: portType,
																							SwitchVlan: uint32(vlan), MacAddress: mac, PortStat: stat, LastTime: time.Now()}
																						err = db.Create(&asp).Error
																					}
																				}
																			}
																		}
																	}
																}
																session.ClearChannel()
																session.WriteChannel(Show[brand] + " vlan brief")
																for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																	if strings.Contains(line, "VLAN 00") {
																		VlanId, _ := strconv.Atoi(strings.Split(line, " ")[0])
																		db.Where("switch_id = ? and switch_vlan = ?", v.SwitchId, VlanId).First(&ASwitchVlan)
																		if len(ASwitchVlan) == 0 {
																			asv := databases.AssetSwitchVlan{SwitchId: v.SwitchId, SwitchVlan: uint32(VlanId), LastTime: time.Now()}
																			err = db.Create(&asv).Error
																		}
																	}
																}
																//更新交换机关联信息
																neighbor := map[string]struct{}{}
																session.ClearChannel()
																session.WriteChannel(Show[brand] + " lldp neighbor-information list")
																for _, line := range strings.Split(session.ReadChannelTiming(30*time.Second), "\n") {
																	lines := strings.Split(line, " ")
																	if len(lines) > 3 && strings.Contains(line, "switch-") {
																		sw := strings.TrimSpace(lines[len(lines)-1])
																		if strings.Contains(sw, "switch-") {
																			neighbor[sw] = struct{}{}
																		}
																	}
																}
																if len(neighbor) > 0 {
																	relation := map[string]string{}
																	db.Where("switch_id=?", v.SwitchId).Find(&ASwitchRelation)
																	if len(ASwitchRelation) == 0 {
																		for _, d := range ASwitchRelation {
																			relation[d.NeighborName] = d.NeighborId
																		}
																	}
																	for k := range neighbor {
																		_, ok := relation[k]
																		if !ok {
																			db.Where("switch_name=?", k).First(&ASwitch)
																			if len(ASwitch) > 0 {
																				db.Where("switch_id=? and neighbor_id=?", v.SwitchId, ASwitch[0].SwitchId).Find(&ASwitchRelation)
																				if len(ASwitchRelation) == 0 {
																					ar := databases.AssetSwitchRelation{SwitchId: v.SwitchId,
																						SwitchName: v.SwitchName, NeighborId: ASwitch[0].SwitchId,
																						NeighborName: k, UpdateTime: time.Now()}
																					db.Create(&ar)
																				}
																			}
																		} else {
																			db.Model(&databases.AssetSwitchRelation{}).Where("switch_id=? and neighbor_id=?", v.SwitchId, relation[k]).Updates(
																				databases.AssetSwitchRelation{SwitchName: v.SwitchName, NeighborName: k})
																		}
																	}
																}
															}
														}
														session.WriteChannel("quit")
													} else {
														Log.Error("GetSSHBrand err:" + v.SwitchIp + " " + e.Error())
													}
												}
											}
										}
									}(as, ip)
								} else {
									db.Where("switch_ip = ?", ip).Find(&AssetSwitch)
									if len(AssetSwitch) > 0 {
										db.Model(&AssetSwitch).Where("switch_ip = ?", ip).Updates(
											databases.AssetSwitch{Status: "offline"})
									}
								}
								i++
							}
						}
						db.Model(&AssetSwitch).Where("switch_ip in ?", ips).Count(&c)
						err = db.Model(&ASwitchPool).Where("id=?", as.Id).Updates(databases.AssetSwitchPool{
							Discover: int(c), SyncTime: time.Now()}).Error
						if err != nil {
							Log.Error(err)
						}
					}(s)
				}
			}
		} else {
			Log.Error(err)
		}
	}
}

func DiscoverServer() {
	var (
		err             error
		AssetServerPool []databases.AssetServerPool
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	lock := common.SyncMutex{LockKey: "cmdb_discover_server_lock"}
	//加锁
	if lock.Lock() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		db.Where("status=?", "enable").Find(&AssetServerPool)
		if len(AssetServerPool) > 0 {
			Log.Info("服务器巡检任务开始执行......")
			for _, d := range AssetServerPool {
				go func(d databases.AssetServerPool) {
					var (
						SshKey      []databases.SshKey
						AServerPool []databases.AssetServerPool
						ASNet       []databases.AssetNet
						AssetServer []databases.AssetServer
						c           int64
						ips         []string
						pkey        string
					)
					SIps := strings.Split(d.StartIp, ".")
					EIps := strings.Split(d.EndIp, ".")
					if len(SIps) == 4 && len(EIps) == 4 {
						i, _ := strconv.Atoi(SIps[len(SIps)-1])
						e, _ := strconv.Atoi(EIps[len(EIps)-1])
						SIp := strings.Join(SIps[:len(SIps)-1], ".")
						if d.SshPassword != "none" {
							passwd, err := Encrypt.DecryptString(d.SshPassword, true)
							if err == nil {
								d.SshPassword = string(passwd)
							}
						}
						if d.SshKeyName != "none" {
							db.Where("key_name=?", d.SshKeyName).First(&SshKey)
							if len(SshKey) > 0 {
								k, err := Encrypt.DecryptString(SshKey[0].SshKey, true)
								if err == nil {
									pkey = string(k)
								}
							}
						}
						for i <= e {
							var hostId string
							ip := SIp + "." + strconv.Itoa(i)
							ips = append(ips, ip)
							rc.HSet(ctx, platform_conf.IdcIdKey, ip, d.IdcId)
							if netutil.IsPingConnected(ip) {
								if rc.HExists(ctx, platform_conf.IpHostIdKey, ip).Val() {
									hostId = rc.HGet(ctx, platform_conf.IpHostIdKey, ip).Val()
									rc.HSet(ctx, platform_conf.IdcIdKey, hostId, d.IdcId)
									rc.HSet(ctx, platform_conf.AssetPoolIdsKey, hostId, d.Id)
								}

							}
							if net.ParseIP(platform_conf.RemoteAddr) != nil {
								var (
									do  bool
									cmd = "yum -y install curl && curl -s http://" +
										platform_conf.RemoteAddr + "/api/v1/ag/install.sh|bash"
								)
								if hostId == "" {
									do = true
								} else {
									db.Where("host_id=?", hostId).Find(&AssetServer)
									if len(AssetServer) == 0 {
										do = true
									} else {
										if rc.Exists(ctx, platform_conf.AgentAliveTraceKey+"_"+hostId).Val() == 1 {
											do = true
										}
									}
								}
								if do {
									client, err := common.SshClient(ip, strconv.Itoa(d.SshPort), d.SshUser, d.SshPassword, pkey)
									if err == nil {
										session, err := client.NewSession()
										if err == nil {
											if _, err := session.CombinedOutput(cmd); err != nil {
												Log.Error("Failed to run: " + err.Error())
											}
										}
										_ = session.Close()
									}
								}
							}
							i++
						}
						db.Model(&ASNet).Where("ip in ?", ips).Count(&c)
						err = db.Model(&AServerPool).Where("id=?", d.Id).Updates(
							map[string]interface{}{"discover": c, "sync_time": time.Now()}).Error
					}
				}(d)
			}
		}
	}
}

func CheckSwitch() {
	var (
		err         error
		AssetSwitch []databases.AssetSwitch
	)
	lock := common.SyncMutex{LockKey: "cmdb_check_switch_lock"}
	//加锁
	if lock.Lock() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				Log.Error(err)
			}
			lock.UnLock(true)
		}()
		db.Find(&AssetSwitch)
		if len(AssetSwitch) > 0 {
			for _, s := range AssetSwitch {
				status := "offline"
				if netutil.IsPingConnected(s.SwitchIp) {
					status = "online"
				}
				db.Model(&AssetSwitch).Where("switch_id=?", s.SwitchId).Updates(databases.AssetSwitch{
					Status: status})
			}
		}
	}
}
