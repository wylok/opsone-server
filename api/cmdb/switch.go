package cmdb

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ssh "github.com/shenbowei/switch-ssh-go"
	"gorm.io/gorm"
	"inner/conf/cmdb_conf"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"strconv"
	"strings"
	"time"
)

// @Tags 交换机
// @Summary 查询交换机列表
// @Produce  json
// @Security ApiKeyAuth
// @Param switch_id query string false "交换机ID"
// @Param switch_ip query string false "交换机IP"
// @Param switch_name query string false "交换机名称"
// @Param host_mac query string false "服务器MAC"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/switches [get]
func QuerySwitches(c *gin.Context) {
	//查询交换机资源池
	var (
		JsonData        cmdb_conf.QuerySwitch
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPort []databases.AssetSwitchPort
		Response        = common.Response{C: c}
		ExtData         = map[string]string{}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Where("switch_id != ?", "none").Order("switch_ip")
		// 参数匹配
		if JsonData.HostMac != "" {
			mac := strings.Split(JsonData.HostMac, ":")
			if len(mac) == 6 {
				db.Where("mac_address=?", strings.Join(mac[0:2], "")+"-"+
					strings.Join(mac[2:4], "")+"-"+strings.Join(mac[4:6], "")).First(&AssetSwitchPort)
				if len(AssetSwitchPort) > 0 {
					ExtData["mac_address"] = JsonData.HostMac
					ExtData["port_name"] = AssetSwitchPort[0].PortName
					ExtData["vlan"] = "vlan:" + strconv.Itoa(int(AssetSwitchPort[0].SwitchVlan))
					tx = tx.Where("switch_id = ?", AssetSwitchPort[0].SwitchId)
				} else {
					tx = tx.Where("switch_id = ?", 1)
				}
			} else {
				tx = tx.Where("switch_id = ?", 1)
			}
		}
		if JsonData.SwitchId != "" {
			tx = tx.Where("switch_id = ?", JsonData.SwitchId)
		}
		if JsonData.SwitchName != "" {
			tx = tx.Where("switch_name like ?", "%"+JsonData.SwitchName+"%")
		}
		if JsonData.SwitchIp != "" {
			tx = tx.Where("switch_ip = ?", JsonData.SwitchIp)
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, _ = p.Paging(&AssetSwitch)
		if len(AssetSwitch) > 0 {
			var (
				pc   int64
				vc   int64
				Data []map[string]interface{}
			)
			for _, v := range AssetSwitch {
				db.Model(&databases.AssetSwitchPort{}).Where("switch_id = ?", v.SwitchId).Count(&pc)
				db.Model(&databases.AssetSwitchVlan{}).Where("switch_id = ?", v.SwitchId).Count(&vc)
				Data = append(Data, map[string]interface{}{"switch_pool_id": v.SwitchPoolId, "switch_name": v.SwitchName,
					"switch_ip": v.SwitchIp, "switch_id": v.SwitchId, "switch_brand": v.SwitchBrand, "port_count": pc,
					"vlan_count": vc, "switch_version": v.SwitchVersion, "idc_id": v.IdcId, "status": v.Status, "sync_time": v.SyncTime,
					"ExtData": ExtData})
			}
			Response.Data = Data
		}
	}
}

// @Tags 交换机
// @Summary 查询交换机端口
// @Produce  json
// @Security ApiKeyAuth
// @Param switch_id query string true "交换机ID"
// @Param port_name query string false "端口名称"
// @Param mac_address query string false "mac地址"
// @Param page query integer false "页码"
// @Param pre_page query integer false "每页行数"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/switch/port [get]
func QuerySwitchPort(c *gin.Context) {
	//查询交换机资源池
	var (
		JsonData        = cmdb_conf.QuerySwitchPort{}
		AssetSwitchPort []databases.AssetSwitchPort
		Response        = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if JsonData.Page == 0 {
			JsonData.Page = 1
		}
		if JsonData.PerPage == 0 {
			JsonData.PerPage = 15
		}
		tx := db.Where("switch_id = ?", JsonData.SwitchId)
		if JsonData.MacAddress != "" {
			JsonData.Page = 1
			mac := strings.Split(JsonData.MacAddress, ":")
			if len(mac) == 6 {
				tx = tx.Where("mac_address=?", strings.Join(mac[0:2], "")+"-"+
					strings.Join(mac[2:4], "")+"-"+strings.Join(mac[4:6], ""))
			}
		}
		if JsonData.PortName != "" {
			JsonData.Page = 1
			tx = tx.Where("port_name like ?", "%"+JsonData.PortName+"%")
		}
		p := databases.Pagination{DB: tx, Page: JsonData.Page, PerPage: JsonData.PerPage}
		Response.Pages, Response.Data = p.Paging(&AssetSwitchPort)
	}
}

// @Tags 交换机
// @Summary 查询交换机Vlan
// @Produce  json
// @Security ApiKeyAuth
// @Param switch_id query string true "交换机ID"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/switch/vlan [get]
func QuerySwitchVlan(c *gin.Context) {
	//查询交换机资源池
	var (
		JsonData        = cmdb_conf.QuerySwitchVlan{}
		AssetSwitchVlan []databases.AssetSwitchVlan
		Response        = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id = ?", JsonData.SwitchId).Find(&AssetSwitchVlan)
		if len(AssetSwitchVlan) > 0 {
			Response.Data = AssetSwitchVlan
		}
	}
}

// @Tags 交换机
// @Summary 查询交换机级联
// @Produce  json
// @Security ApiKeyAuth
// @Param switch_id query string false "交换机ID"
// @Success 200 {} json "{pages:{},success:true,message:"ok",data:[]}"
// @Router /api/v1/cmdb/switch/relation [get]
func QuerySwitchesRelation(c *gin.Context) {
	//查询交换机级联
	var (
		JsonData            = cmdb_conf.QuerySwitchRelation{}
		AssetSwitchRelation []databases.AssetSwitchRelation
		Response            = common.Response{C: c}
	)
	err := c.ShouldBindQuery(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=?", JsonData.SwitchId).Find(&AssetSwitchRelation)
		if len(AssetSwitchRelation) > 0 {
			Response.Data = AssetSwitchRelation
		}
	}
}

// @Tags 交换机
// @Summary 新增交换机VLAN
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.AddSwitchVlan true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch/vlan [put]
func AddSwitchVlan(c *gin.Context) {
	//新增交换机VLAN
	var (
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		JsonData        = cmdb_conf.AddSwitchVlan{}
		Response        = common.Response{C: c}
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		if len(JsonData.SwitchIps) == 0 {
			err = errors.New("交换机IP不能为空")
		}
		vlan := kits.FormListFormat([]string{JsonData.Vlan})
		for _, v := range vlan {
			_, err = strconv.Atoi(v)
			if err != nil {
				err = errors.New("无效vlan信息:" + v)
				break
			}
		}
		if err == nil {
			for _, ip := range JsonData.SwitchIps {
				db.Where("switch_ip=? and switch_id!=?", ip, "none").First(&AssetSwitch)
				if len(AssetSwitch) == 0 {
					err = errors.New("无效IP信息:" + ip)
					break
				}
				switchId := AssetSwitch[0].SwitchId
				db.Where("id=?", AssetSwitch[0].SwitchPoolId).First(&AssetSwitchPool)
				if len(AssetSwitchPool) > 0 {
					pwd, e := Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
					if e == nil {
						go func(ip, port, user, pwd string) {
							session, e := ssh.NewSSHSession(user, pwd, ip+":"+port)
							if e == nil {
								defer func() {
									if r := recover(); r != nil {
										e = errors.New(fmt.Sprint(r))
									}
									if e != nil {
										Log.Error(e)
									}
									session.Close()
								}()
								if session.GetSSHBrand() != "" {
									session.ClearChannel()
									session.WriteChannel("system")
									for _, v := range vlan {
										session.WriteChannel("vlan "+v, "quit")
										sv, err := strconv.Atoi(v)
										if err == nil {
											asv := databases.AssetSwitchVlan{SwitchId: switchId, SwitchVlan: uint32(sv), LastTime: time.Now()}
											err = db.Create(&asv).Error
										} else {
											Log.Error(err)
											break
										}
									}
									session.WriteChannel("save force", "quit")
								}
							} else {
								err = e
							}
						}(ip, strconv.Itoa(AssetSwitchPool[0].SwitchPort), AssetSwitchPool[0].SwitchUser, string(pwd))
					} else {
						err = e
					}
				} else {
					err = errors.New("未找到相关" + ip + "交换机配置信息")
					break
				}
			}
		}
	}
}

// @Tags 交换机
// @Summary 端口变更VLAN
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.ChangeSwitchPortVlan true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch/port/vlan [post]
func ChangeSwitchPortVlan(c *gin.Context) {
	//端口变更VLAN
	var (
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		AssetSwitchVlan []databases.AssetSwitchVlan
		AssetSwitchPort []databases.AssetSwitchPort
		JsonData        = cmdb_conf.ChangeSwitchPortVlan{}
		Response        = common.Response{C: c}
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=?", JsonData.SwitchId).First(&AssetSwitch)
		if len(AssetSwitch) == 0 {
			err = errors.New("无效ID信息:" + JsonData.SwitchId)
		} else {
			db.Where("switch_id=? and switch_vlan=?", JsonData.SwitchId, JsonData.NewVlan).First(&AssetSwitchVlan)
			if len(AssetSwitchVlan) > 0 {
				db.Where("switch_id=? and port_name=?", JsonData.SwitchId, JsonData.PortName).First(&AssetSwitchPort)
				if len(AssetSwitchPort) > 0 {
					if AssetSwitchPort[0].PortType == "Access" {
						db.Where("id=?", AssetSwitch[0].SwitchPoolId).First(&AssetSwitchPool)
						if len(AssetSwitchPool) > 0 {
							pwd, e := Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
							if e == nil {
								go func(ip, port, user, pwd string) {
									session, e := ssh.NewSSHSession(user, pwd, ip+":"+port)
									defer func() {
										if r := recover(); r != nil {
											e = errors.New(fmt.Sprint(r))
										}
										if e != nil {
											Log.Error(e)
										}
										session.Close()
									}()
									if e == nil {
										//变更端口Vlan
										if session.GetSSHBrand() != "" {
											session.ClearChannel()
											session.WriteChannel(
												"system",
												"interface "+JsonData.PortName,
												"port access vlan "+JsonData.NewVlan,
												"save force",
												"quit", "quit")
											session.ReadChannelTiming(10 * time.Second)
											v, _ := strconv.Atoi(JsonData.NewVlan)
											err = db.Model(&AssetSwitchPort).Where("switch_id=? and port_name=?",
												JsonData.SwitchId, JsonData.PortName).Updates(
												databases.AssetSwitchPort{SwitchVlan: uint32(v)}).Error
										}
									} else {
										err = e
									}
								}(AssetSwitch[0].SwitchIp, strconv.Itoa(AssetSwitchPool[0].SwitchPort),
									AssetSwitchPool[0].SwitchUser, string(pwd))
							} else {
								err = e
							}
						} else {
							err = errors.New("未找到" + JsonData.SwitchId + "交换机配置信息")
						}
					} else {
						err = errors.New("端口非access类型不支持变更vlan")
					}
				} else {
					err = errors.New("无效端口信息:" + JsonData.PortName)
				}
			} else {
				err = errors.New("无效Vlan信息:" + JsonData.NewVlan)
			}
		}
	}
}

// @Tags 交换机
// @Summary 开启/关闭端口
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.SwitchPortOperate true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch/port/operate [post]
func SwitchPortOperate(c *gin.Context) {
	//开启/关闭端口
	var (
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		AssetSwitchPort []databases.AssetSwitchPort
		JsonData        = cmdb_conf.SwitchPortOperate{}
		Response        = common.Response{C: c}
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=?", JsonData.SwitchId).First(&AssetSwitch)
		if len(AssetSwitch) == 0 {
			err = errors.New("无效ID信息:" + JsonData.SwitchId)
		} else {
			db.Where("switch_id=? and port_name=?", JsonData.SwitchId, JsonData.PortName).First(&AssetSwitchPort)
			if len(AssetSwitchPort) > 0 {
				if AssetSwitchPort[0].PortStat != JsonData.Operate {
					db.Where("id=?", AssetSwitch[0].SwitchPoolId).First(&AssetSwitchPool)
					if len(AssetSwitchPool) > 0 {
						pwd, e := Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
						if e == nil {
							action := "shutdown"
							stats := "DOWN"
							if JsonData.Operate == "UP" {
								action = "no shutdown"
								stats = "UP"
							}
							go func(ip, port, user, pwd, action string) {
								session, e := ssh.NewSSHSession(user, pwd, ip+":"+port)
								defer func() {
									if r := recover(); r != nil {
										e = errors.New(fmt.Sprint(r))
									}
									if e != nil {
										Log.Error(e)
									}
									session.Close()
								}()
								if e == nil {
									//变更端口Vlan
									if session.GetSSHBrand() != "" {
										session.ClearChannel()
										session.WriteChannel(
											"system",
											"interface "+JsonData.PortName,
											action,
											"save force",
											"quit", "quit")
										err = db.Model(&AssetSwitchPort).Where("switch_id=? and port_name=?",
											JsonData.SwitchId, JsonData.PortName).Updates(
											databases.AssetSwitchPort{PortStat: stats}).Error
									}
								} else {
									err = e
								}
							}(AssetSwitch[0].SwitchIp, strconv.Itoa(AssetSwitchPool[0].SwitchPort),
								AssetSwitchPool[0].SwitchUser, string(pwd), action)
						} else {
							err = e
						}
					} else {
						err = errors.New("未找到" + JsonData.SwitchId + "交换机配置信息")
					}
				}
			} else {
				err = errors.New("无效端口信息:" + JsonData.PortName)
			}
		}
	}
}

// @Tags 交换机
// @Summary 执行命令
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.SwitchOperate true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch/operate [post]
func SwitchOperate(c *gin.Context) {
	//执行命令
	var (
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		JsonData        = cmdb_conf.SwitchOperate{}
		Response        = common.Response{C: c}
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=?", JsonData.SwitchId).First(&AssetSwitch)
		if len(AssetSwitch) == 0 {
			err = errors.New("无效ID信息:" + JsonData.SwitchId)
		} else {
			db.Where("id=?", AssetSwitch[0].SwitchPoolId).First(&AssetSwitchPool)
			if len(AssetSwitchPool) > 0 {
				pwd, e := Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
				if e == nil {
					go func(ip, port, user, pwd, commands string) {
						session, e := ssh.NewSSHSession(user, pwd, ip+":"+port)
						defer func() {
							if r := recover(); r != nil {
								e = errors.New(fmt.Sprint(r))
							}
							if e != nil {
								Log.Error(e)
							}
							session.Close()
						}()
						if e == nil {
							//执行命令
							if session.GetSSHBrand() != "" {
								session.ClearChannel()
								session.WriteChannel("system")
								for _, command := range strings.Split(JsonData.Commands, ";") {
									session.WriteChannel(command)
								}
								session.WriteChannel("save force", "quit", "quit")
							}
						} else {
							err = e
						}
					}(AssetSwitch[0].SwitchIp, strconv.Itoa(AssetSwitchPool[0].SwitchPort),
						AssetSwitchPool[0].SwitchUser, string(pwd), JsonData.Commands)
				} else {
					err = e
				}
			} else {
				err = errors.New("未找到" + JsonData.SwitchId + "交换机配置信息")
			}
		}
	}
}

// @Tags 交换机
// @Summary 修改交换机名称
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.SwitchName true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch/name [post]
func SwitchName(c *gin.Context) {
	//修改交换机名称
	var (
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		JsonData        = cmdb_conf.SwitchName{}
		Response        = common.Response{C: c}
		Encrypt         = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=?", JsonData.SwitchId).First(&AssetSwitch)
		if len(AssetSwitch) == 0 {
			err = errors.New("无效ID信息:" + JsonData.SwitchId)
		} else {
			db.Where("id=?", AssetSwitch[0].SwitchPoolId).First(&AssetSwitchPool)
			if len(AssetSwitchPool) > 0 {
				pwd, e := Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
				if e == nil {
					go func(ip, port, user, pwd, name string) {
						session, e := ssh.NewSSHSession(user, pwd, ip+":"+port)
						defer func() {
							if r := recover(); r != nil {
								e = errors.New(fmt.Sprint(r))
							}
							if e != nil {
								Log.Error(e)
							}
							session.Close()
						}()
						if err == nil {
							//执行命令
							if session.GetSSHBrand() != "" {
								session.WriteChannel("system", "hostname "+name)
								session.WriteChannel("save force", "quit", "quit")
							}
						} else {
							err = e
						}
					}(AssetSwitch[0].SwitchIp, strconv.Itoa(AssetSwitchPool[0].SwitchPort),
						AssetSwitchPool[0].SwitchUser, string(pwd), JsonData.Name)
				} else {
					err = e
				}
			} else {
				err = errors.New("未找到" + JsonData.SwitchId + "交换机配置信息")
			}
		}
	}
}

// @Tags 交换机
// @Summary 删除交换机
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body  cmdb_conf.SwitchName true "json数据"
// @Success 200 {} json "{success:true,message:"ok",data:null}"
// @Router /api/v1/cmdb/switch [delete]
func DeleteSwitch(c *gin.Context) {
	//修改交换机名称
	var (
		AssetSwitch []databases.AssetSwitch
		JsonData    = cmdb_conf.SwitchName{}
		Response    = common.Response{C: c}
	)
	err := c.ShouldBindJSON(&JsonData)
	// 接口请求返回
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			Log.Error(err)
		}
		Response.Err = err
		Response.Send()
	}()
	if err == nil {
		db.Where("switch_id=? and status=?", JsonData.SwitchId, JsonData.Name, "offline").Find(&AssetSwitch)
		if len(AssetSwitch) == 0 {
			err = errors.New(JsonData.Name + "交换机在线无法删除")
		} else {
			err = db.Transaction(func(tx *gorm.DB) error {
				sqlErr := tx.Where("switch_id=?", JsonData.SwitchId).Delete(&AssetSwitch).Error
				sqlErr = tx.Where("switch_id=?", JsonData.SwitchId).Delete(&databases.AssetSwitchPort{}).Error
				sqlErr = tx.Where("switch_id=?", JsonData.SwitchId).Delete(&databases.AssetSwitchVlan{}).Error
				sqlErr = tx.Where("switch_id=?", JsonData.SwitchId).Delete(&databases.AssetSwitchRelation{}).Error
				return sqlErr
			})
		}
	}
}
