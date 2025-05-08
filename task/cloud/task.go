package cloud

import (
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/baidubce/bce-sdk-go/services/bos"
	bosAPi "github.com/baidubce/bce-sdk-go/services/bos/api"
	common2 "github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/golang-module/carbon"
	volcEcs "github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/databases"
	"strings"
	"sync"
	"time"
)

func SyncAliYunEcs() {
	lock := common.SyncMutex{LockKey: "AliYun_ecs_sync_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				err         error
				CloudServer []databases.CloudServer
				CloudKeys   []databases.CloudKeys
				InstanceIds = map[string]interface{}{}
			)
			//同步阿里云ECS信息
			Log.Info("AliYunEcsSync start working ......")
			defer func() {
				lock.UnLock(true)
			}()
			db.Where("cloud=? and key_type=?", "aliyun", "ecs").Find(&CloudKeys)
			if len(CloudKeys) > 0 {
				wg := sync.WaitGroup{}
				for _, v := range CloudKeys {
					wg.Add(1)
					go func(v databases.CloudKeys) {
						defer wg.Done()
						defer func() {
							if r := recover(); r != nil {
								err = errors.New(fmt.Sprint(r))
							}
							if err != nil {
								Log.Error(err)
							}
						}()
						client := ecs.NewClient(v.KeyId, v.KeySecret)
						regionId := strings.Split(v.EndPoint, ".")
						if len(regionId) >= 2 {
							client.WithRegionID(common2.Region(regionId[1]))
						}
						client.WithEndpoint("https://" + v.EndPoint)
						PageNumber := 1
						PageSize := 50
						for {
							Args := ecs.DescribeInstancesArgs{Pagination: common2.Pagination{PageNumber: PageNumber, PageSize: PageSize}}
							instances, pagination, err := client.DescribeInstances(&Args)
							if err == nil && len(instances) > 0 {
								if PageNumber > 1 {
									if pagination.PageNumber*PageSize > pagination.TotalCount {
										break
									} else {
										PageNumber = pagination.PageNumber
									}
								}
								for _, i := range instances {
									InstanceIds[i.InstanceId] = struct{}{}
									IpAddress := strings.Join(i.InnerIpAddress.IpAddress, ",")
									if len(i.NetworkInterfaces.NetworkInterface) > 0 {
										IpAddress = i.NetworkInterfaces.NetworkInterface[0].PrimaryIpAddress
									}
									db.Where("cloud=? and instance_id=?", "aliyun", i.InstanceId).Find(&CloudServer)
									if len(CloudServer) == 0 {
										ae := databases.CloudServer{Cloud: "aliyun", InstanceId: i.InstanceId,
											InstanceName: i.InstanceName,
											InstanceType: i.InstanceType, Description: i.Description,
											HostName: i.HostName,
											Sn:       i.SerialNumber, RegionId: fmt.Sprint(i.RegionId),
											ZoneId: i.ZoneId,
											Cpu:    i.Cpu, Memory: i.Memory,
											PublicIpAddress: strings.Join(i.PublicIpAddress.IpAddress, ","),
											InnerIpAddress:  IpAddress,
											Status:          fmt.Sprint(i.Status),
											CreationDate:    carbon.Parse(i.CreationTime.String()).Carbon2Time(),
											ExpiredTime:     carbon.Parse(i.ExpiredTime.String()).Carbon2Time(),
											KeyId:           v.KeyId, SyncTime: time.Now()}
										err = db.Create(&ae).Error
									} else {
										err = db.Model(&CloudServer).Where("cloud=? and instance_id=?",
											"aliyun", i.InstanceId).Updates(
											databases.CloudServer{Description: i.Description, HostName: i.HostName,
												Sn: i.SerialNumber, Cpu: i.Cpu, Memory: i.Memory,
												PublicIpAddress: strings.Join(i.PublicIpAddress.IpAddress, ","),
												InnerIpAddress:  IpAddress,
												Status:          fmt.Sprint(i.Status),
												ExpiredTime:     carbon.Parse(i.ExpiredTime.String()).Carbon2Time(),
												KeyId:           v.KeyId, SyncTime: time.Now()}).Error
									}
								}
								if PageNumber == 1 {
									PageNumber++
								}
							} else {
								break
							}
							if err != nil {
								Log.Error(err)
							}
						}
					}(v)
				}
				wg.Wait()
			}
			if len(InstanceIds) > 0 {
				db.Select("instance_id").Where("cloud=?", "aliyun").Find(&CloudServer)
				for _, v := range CloudServer {
					_, ok := InstanceIds[v.InstanceId]
					if !ok {
						db.Where("cloud=? and instance_id=?", "aliyun", v.InstanceId).Delete(&CloudServer)
					}
				}
			}
		}()
	}
}

func SyncAliYunOss() {
	lock := common.SyncMutex{LockKey: "AliYun_oss_sync_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				CloudKeys []databases.CloudKeys
				CloudOss  []databases.CloudOss
				NewOss    = map[string]struct{}{}
			)
			//同步阿里云OSS信息
			Log.Info("AliYunOssSync start working ......")
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
				lock.UnLock(true)
			}()
			db.Where("cloud=? and key_type=?", "aliyun", "oss").Find(&CloudKeys)
			if len(CloudKeys) > 0 {
				for _, v := range CloudKeys {
					client, err := oss.New("https://"+v.EndPoint, v.KeyId, v.KeySecret)
					if err != nil {
						Log.Error(err)
					} else {
						lsRes, err := client.ListBuckets(oss.MaxKeys(100))
						if err != nil {
							Log.Error(err)
						} else {
							for _, bucket := range lsRes.Buckets {
								Stat, _ := client.GetBucketStat(bucket.Name)
								NewOss[bucket.Name] = struct{}{}
								db.Where("cloud=? and bucket=?", "aliyun", bucket.Name).First(&CloudOss)
								if len(CloudOss) > 0 {
									err = db.Model(&CloudOss).Where("cloud=? and bucket=?", "aliyun", bucket.Name).Updates(
										map[string]interface{}{
											"Location":     bucket.Location,
											"StorageClass": bucket.StorageClass,
											"CreationDate": bucket.CreationDate,
											"Storage":      Stat.Storage,
											"ObjectCount":  Stat.ObjectCount,
											"sync_time":    time.Now()}).Error
								} else {
									ao := databases.CloudOss{
										Cloud:        "aliyun",
										KeyId:        v.KeyId,
										Bucket:       bucket.Name,
										Location:     bucket.Location,
										StorageClass: bucket.StorageClass,
										CreationDate: bucket.CreationDate,
										Storage:      Stat.Storage,
										ObjectCount:  Stat.ObjectCount,
										SyncTime:     time.Now()}
									err = db.Create(&ao).Error
								}
							}
							if err != nil {
								Log.Error(err)
							}
						}
					}
					if err != nil {
						Log.Error(err)
					}
				}
				if len(NewOss) > 0 {
					db.Where("cloud=?", "aliyun").Find(&CloudOss)
					if len(CloudOss) > 0 {
						for _, v := range CloudOss {
							_, ok := NewOss[v.Bucket]
							if !ok {
								db.Where("cloud=? and bucket=?", "aliyun", v.Bucket).Delete(&CloudOss)
							}
						}
					}
				}
			}
		}()
	}
}

func SyncBaiduBcc() {
	lock := common.SyncMutex{LockKey: "baidu_bcc_sync_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				err         error
				CloudServer []databases.CloudServer
				CloudKeys   []databases.CloudKeys
				InstanceIds = map[string]struct{}{}
			)
			//同步百度云BCC信息
			Log.Info("BaiduBccSync start working ......")
			defer func() {
				lock.UnLock(true)
			}()
			db.Where("cloud=? and key_type=?", "baidu", "bcc").Find(&CloudKeys)
			if len(CloudKeys) > 0 {
				wg := sync.WaitGroup{}
				for _, v := range CloudKeys {
					wg.Add(1)
					go func(v databases.CloudKeys) {
						defer wg.Done()
						defer func() {
							if r := recover(); r != nil {
								err = errors.New(fmt.Sprint(r))
							}
							if err != nil {
								Log.Error(err)
							}
						}()
						bccClient, err := bcc.NewClient(v.KeyId, v.KeySecret, v.EndPoint)
						if err == nil {
							for {
								args := &api.ListInstanceArgs{}
								var marker string
								if marker != "" {
									args.Marker = marker
								}
								result, err := bccClient.ListInstances(args)
								if err == nil {
									for _, i := range result.Instances {
										if rc.Exists(ctx, platform_conf.ServerSnKey+"_"+i.NicInfo.DeviceId).Val() == 0 {
											db.Where("cloud=? and instance_id=?", "baidu", i.InstanceId).First(&CloudServer)
											InstanceIds[i.InstanceId] = struct{}{}
											if len(CloudServer) == 0 && len(i.ExpireTime) > 0 {
												bb := databases.CloudServer{Cloud: "baidu", InstanceId: i.InstanceId,
													InstanceName: i.InstanceName,
													InstanceType: fmt.Sprint(i.InstanceType), Description: i.Description,
													HostName: i.InstanceName, Sn: i.NicInfo.DeviceId, RegionId: v.EndPoint,
													ZoneId: i.ZoneName, Cpu: i.CpuCount, Memory: i.MemoryCapacityInGB * 1000,
													PublicIpAddress: i.PublicIP, InnerIpAddress: i.InternalIP,
													Status:       fmt.Sprint(i.Status),
													CreationDate: carbon.Parse(i.CreationTime).Carbon2Time(),
													ExpiredTime:  carbon.Parse(i.ExpireTime).Carbon2Time(),
													KeyId:        v.KeyId, SyncTime: time.Now()}
												err = db.Create(&bb).Error
											} else {
												err = db.Model(&CloudServer).Where("cloud=? and instance_id=?",
													"baidu", i.InstanceId).Updates(
													databases.CloudServer{Description: i.Description,
														InstanceName: i.InstanceName,
														HostName:     i.InstanceName, Sn: i.NicInfo.DeviceId,
														Cpu: i.CpuCount, Memory: i.MemoryCapacityInGB * 1000,
														PublicIpAddress: i.PublicIP,
														InnerIpAddress:  i.InternalIP,
														Status:          fmt.Sprint(i.Status),
														ExpiredTime:     carbon.Parse(i.ExpireTime).Carbon2Time(),
														KeyId:           v.KeyId, SyncTime: time.Now()}).Error
											}
										}
									}
									if err != nil {
										Log.Error(err)
									}
									if !result.IsTruncated {
										break
									} else {
										marker = result.NextMarker
									}
								}
							}
						}
						if err != nil {
							Log.Error(err)
						}
					}(v)
				}
				wg.Wait()
			}
			if len(InstanceIds) > 0 {
				db.Select("instance_id").Where("cloud=?", "baidu").Find(&CloudServer)
				for _, v := range CloudServer {
					_, ok := InstanceIds[v.InstanceId]
					if !ok {
						db.Where("cloud=? and instance_id=?", "baidu", v.InstanceId).Delete(&CloudServer)
					}
				}
			}
		}()
	}
}

func SyncBaiduOss() {
	lock := common.SyncMutex{LockKey: "Baidu_oss_sync_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				CloudKeys []databases.CloudKeys
				CloudOss  []databases.CloudOss
				NewOss    = map[string]struct{}{}
			)
			//同步阿里云OSS信息
			Log.Info("BaiduOssSync start working ......")
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprint(r))
				}
				if err != nil {
					Log.Error(err)
				}
				lock.UnLock(true)
			}()
			db.Where("cloud=? and key_type=?", "baidu", "bos").Find(&CloudKeys)
			if len(CloudKeys) > 0 {
				for _, v := range CloudKeys {
					clientConfig := bos.BosClientConfiguration{
						Ak:               v.KeyId,
						Sk:               v.KeySecret,
						Endpoint:         v.EndPoint,
						RedirectDisabled: false,
					}
					// 初始化一个BosClient
					bosClient, err := bos.NewClientWithConfig(&clientConfig)
					if err == nil {
						res, err := bosClient.ListBuckets()
						if err == nil {
							for _, bucket := range res.Buckets {
								storageClass, _ := bosClient.GetBucketStorageclass(bucket.Name)
								args := &bosAPi.ListObjectsArgs{Delimiter: "", Marker: "", MaxKeys: 1000000, Prefix: ""}
								ObjectsResult, _ := bosClient.ListObjects(bucket.Name, args)
								var (
									Storage     int64
									ObjectCount int64
								)
								for _, v := range ObjectsResult.Contents {
									Storage = int64(v.Size) + Storage
									ObjectCount++
								}
								NewOss[bucket.Name] = struct{}{}
								db.Where("cloud=? and bucket=?", "baidu", bucket.Name).First(&CloudOss)
								if len(CloudOss) > 0 {
									err = db.Model(&CloudOss).Where("cloud=? and bucket=?", "baidu", bucket.Name).Updates(
										map[string]interface{}{
											"Location":     bucket.Location,
											"StorageClass": storageClass,
											"Storage":      Storage,
											"ObjectCount":  ObjectCount,
											"sync_time":    time.Now()}).Error
								} else {
									ao := databases.CloudOss{
										Cloud:        "baidu",
										KeyId:        v.KeyId,
										Bucket:       bucket.Name,
										Location:     bucket.Location,
										StorageClass: storageClass,
										CreationDate: carbon.Parse(bucket.CreationDate).Carbon2Time(),
										Storage:      Storage,
										ObjectCount:  ObjectCount,
										SyncTime:     time.Now()}
									err = db.Create(&ao).Error
								}
							}
						}
					}
					if err != nil {
						Log.Error(err)
					}
				}
				if len(NewOss) > 0 {
					db.Where("cloud=?", "baidu").Find(&CloudOss)
					if len(CloudOss) > 0 {
						for _, v := range CloudOss {
							_, ok := NewOss[v.Bucket]
							if !ok {
								db.Where("cloud=? and bucket=?", "baidu", v.Bucket).Delete(&CloudOss)
							}
						}
					}
				}
			}
		}()
	}
}

func SyncVolcengineEcs() {
	lock := common.SyncMutex{LockKey: "Volcengine_ecs_sync_lock"}
	//加锁
	if lock.Lock() {
		func() {
			var (
				err         error
				CloudServer []databases.CloudServer
				CloudKeys   []databases.CloudKeys
				InstanceIds = map[string]interface{}{}
			)
			//同步火山云ECS信息
			Log.Info("VolcengineEcsSync start working ......")
			defer func() {
				lock.UnLock(true)
			}()
			db.Where("cloud=? and key_type=?", "volcengine", "ecs").Find(&CloudKeys)
			if len(CloudKeys) > 0 {
				wg := sync.WaitGroup{}
				for _, v := range CloudKeys {
					wg.Add(1)
					go func(v databases.CloudKeys) {
						defer wg.Done()
						defer func() {
							if r := recover(); r != nil {
								err = errors.New(fmt.Sprint(r))
							}
							if err != nil {
								Log.Error(err)
							}
						}()
						regionId := "cn-beijing"
						regionIds := strings.Split(v.EndPoint, ".")
						if len(regionIds) >= 2 {
							regionId = regionIds[1]
						}
						config := volcengine.NewConfig().
							WithRegion(regionId).
							WithCredentials(credentials.NewStaticCredentials(v.KeyId, v.KeySecret, ""))
						sess, err := session.NewSession(config)
						if err == nil {
							client := volcEcs.New(sess)
							var NextToken string
							for {
								describeInstancesInput := &volcEcs.DescribeInstancesInput{NextToken: &NextToken}
								outPut, err := client.DescribeInstances(describeInstancesInput)
								if err == nil && len(outPut.Instances) > 0 {
									for _, i := range outPut.Instances {
										InstanceIds[*i.InstanceId] = struct{}{}
										db.Where("cloud=? and instance_id=?", "volcengine", *i.InstanceId).Find(&CloudServer)
										cpu, _ := convertor.ToInt(*i.Cpus)
										mem, _ := convertor.ToInt(*i.MemorySize)
										ipAddress := "0.0.0.0"
										if len(i.NetworkInterfaces) > 0 {
											ipAddress = *i.NetworkInterfaces[0].PrimaryIpAddress
										}
										createdAt := strings.Split(*i.CreatedAt, "T")
										expiredAt := strings.Split(*i.ExpiredAt, "T")
										ct := createdAt[0] + " " + strings.Split(createdAt[1], "+")[0]
										et := expiredAt[0] + " " + strings.Split(expiredAt[1], "+")[0]
										if len(CloudServer) > 0 {
											err = db.Model(&CloudServer).Where("cloud=? and instance_id=?",
												"volcengine", *i.InstanceId).Updates(
												databases.CloudServer{Description: *i.Description,
													HostName:        *i.Hostname,
													Sn:              *i.Uuid,
													Cpu:             int(cpu),
													Memory:          int(mem),
													PublicIpAddress: *i.EipAddress.IpAddress,
													InnerIpAddress:  ipAddress,
													Status:          *i.Status,
													ExpiredTime:     carbon.Parse(et).ToStdTime(),
													KeyId:           v.KeyId, SyncTime: time.Now()}).Error
										} else {
											cs := databases.CloudServer{Cloud: "volcengine", InstanceId: *i.InstanceId,
												InstanceName:    *i.InstanceName,
												InstanceType:    *i.InstanceTypeId,
												Description:     *i.Description,
												HostName:        *i.Hostname,
												Sn:              *i.Uuid,
												RegionId:        regionId,
												ZoneId:          *i.ZoneId,
												Cpu:             int(cpu),
												Memory:          int(mem),
												PublicIpAddress: *i.EipAddress.IpAddress,
												InnerIpAddress:  ipAddress,
												Status:          *i.Status,
												CreationDate:    carbon.Parse(ct).ToStdTime(),
												ExpiredTime:     carbon.Parse(et).ToStdTime(),
												KeyId:           v.KeyId, SyncTime: time.Now()}
											err = db.Create(&cs).Error
										}
									}
								} else {
									break
								}
								if NextToken != *outPut.NextToken {
									NextToken = *outPut.NextToken
								} else {
									break
								}
								if err != nil {
									Log.Error(err)
								}
							}
						}
						if err != nil {
							Log.Error(err)
						}
					}(v)
				}
				wg.Wait()
			}
			if len(InstanceIds) > 0 {
				db.Select("instance_id").Where("cloud=?", "volcengine").Find(&CloudServer)
				for _, v := range CloudServer {
					_, ok := InstanceIds[v.InstanceId]
					if !ok {
						db.Where("cloud=? and instance_id=?", "volcengine", v.InstanceId).Delete(&CloudServer)
					}
				}
			}
		}()
	}
}
