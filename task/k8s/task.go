package k8s

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/mathutil"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"time"
)

func GetNodeMetric() {
	lock := common.SyncMutex{LockKey: "k8s_node_metric_lock"}
	var (
		err        error
		K8sCluster []databases.K8sCluster
	)
	//加锁
	if lock.Lock() {
		defer lock.UnLock(true)
		Log.Info("k8s_node_metric task start working ......")
		db.Find(&K8sCluster)
		if len(K8sCluster) > 0 {
			influx := common.InfluxDb{Cli: Cli, Database: "opsone_k8s"}
			for _, d := range K8sCluster {
				go func(d databases.K8sCluster) {
					defer func() {
						if r := recover(); r != nil {
							err = errors.New(fmt.Sprint(r))
						}
						if err != nil {
							Log.Error(err)
						}
					}()
					k8s := common.K8sCluster{Name: d.K8sName, Config: d.K8sConfig}
					metric := k8s.ListNodeMetrics()
					for k, v := range metric {
						fields := map[string]interface{}{"cpu": float64(0), "memory": 0}
						tags := map[string]string{"k8s_id": d.K8sId, "k8s_name": d.K8sName, "node_name": k}
						cpu := v.(map[string]interface{})["cpu"].(string)
						memory := v.(map[string]interface{})["memory"].(string)
						i, err := strconv.Atoi(cpu)
						if err == nil {
							fields["cpu"] = float64(i)
						}
						if strings.Contains(cpu, "m") {
							i, err := strconv.Atoi(strings.Replace(cpu, "m", "", 1))
							if err == nil {
								fields["cpu"] = float64(i) / 1000
							}
						}
						if strings.Contains(cpu, "n") {
							i, err := strconv.Atoi(strings.Replace(cpu, "n", "", 1))
							if err == nil {
								fields["cpu"] = float64(i) / 1000 / 1000 / 1000
							}
						}
						fields["cpu"] = mathutil.RoundToFloat(fields["cpu"].(float64), 5)
						if strings.Contains(memory, "Ki") {
							memory = strings.Replace(memory, "Ki", "000", 1)
						}
						if strings.Contains(memory, "Mi") {
							memory = strings.Replace(memory, "Mi", "000000", 1)
						}
						if strings.Contains(memory, "Gi") {
							memory = strings.Replace(memory, "Gi", "000000000", 1)
						}
						i, err = strconv.Atoi(memory)
						if err == nil {
							fields["memory"] = i
						}
						err = influx.WritesPoints("node_1m", tags, fields)
					}
				}(d)
			}
		}
	}
}

func GetPodMetric() {
	lock := common.SyncMutex{LockKey: "k8s_pod_metric_lock"}
	var (
		err        error
		K8sCluster []databases.K8sCluster
	)
	//加锁
	if lock.Lock() {
		defer lock.UnLock(true)
		Log.Info("k8s_pod_metric task start working ......")
		db.Find(&K8sCluster)
		if len(K8sCluster) > 0 {
			influx := common.InfluxDb{Cli: Cli, Database: "opsone_k8s"}
			for _, d := range K8sCluster {
				go func(d databases.K8sCluster) {
					defer func() {
						if r := recover(); r != nil {
							err = errors.New(fmt.Sprint(r))
						}
						if err != nil {
							Log.Error(err)
						}
					}()
					k8s := common.K8sCluster{Name: d.K8sName, Config: d.K8sConfig}
					for _, k := range k8s.ListNameSpaces() {
						nameSpace := k["Name"].(string)
						metric := k8s.ListPodMetrics(nameSpace)
						for p, e := range metric {
							pod, _ := k8s.Client().CoreV1().Pods(nameSpace).Get(ctx, p, metav1.GetOptions{})
							var (
								podCpu float64
								podMem int64
							)
							for _, c := range pod.Spec.Containers {
								podCpu = podCpu + float64(c.Resources.Limits.Cpu().MilliValue())
							}
							podCpu = mathutil.RoundToFloat(podCpu/1000, 5)
							for _, c := range pod.Spec.Containers {
								podMem = podMem + c.Resources.Limits.Memory().MilliValue()
							}
							podMem = podMem / 1000
							fields := map[string]interface{}{"cpu": float64(0), "memory": int64(0)}
							tags := map[string]string{"k8s_id": d.K8sId, "k8s_name": d.K8sName, "name_space": nameSpace, "pod_name": p}
							for _, v := range e.([]map[string]interface{}) {
								cpu := v["cpu"].(string)
								memory := v["memory"].(string)
								i, err := strconv.Atoi(cpu)
								if err == nil {
									fields["cpu"] = fields["cpu"].(float64) + float64(i)
								}
								if strings.Contains(cpu, "m") {
									i, err := strconv.Atoi(strings.Replace(cpu, "m", "", 1))
									if err == nil {
										fields["cpu"] = fields["cpu"].(float64) + float64(i)/1000
									}
								}
								if strings.Contains(cpu, "n") {
									i, err := strconv.Atoi(strings.Replace(cpu, "n", "", 1))
									if err == nil {
										fields["cpu"] = fields["cpu"].(float64) + float64(i)/1000/1000/1000
									}
								}
								if strings.Contains(memory, "Ki") {
									memory = strings.Replace(memory, "Ki", "000", 1)
								}
								if strings.Contains(memory, "Mi") {
									memory = strings.Replace(memory, "Mi", "000000", 1)
								}
								if strings.Contains(memory, "Gi") {
									memory = strings.Replace(memory, "Gi", "000000000", 1)
								}
								i, err = strconv.Atoi(memory)
								if err == nil {
									fields["memory"] = fields["memory"].(int64) + int64(i)
								}
							}
							fields["cpu"] = mathutil.RoundToFloat(fields["cpu"].(float64), 5)
							err = influx.WritesPoints("pod_1m", tags, fields)
						}
					}
				}(d)
			}
		}
	}
}

func GetOverView() {
	lock := common.SyncMutex{LockKey: "k8s_overview_lock"}
	var (
		err        error
		Data       = map[string]any{}
		K8sCluster []databases.K8sCluster
	)
	// 接口请求返回结果
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
			Log.Error(err)
		}
	}()
	//加锁
	if lock.Lock() {
		defer lock.UnLock(true)
		Log.Info("k8s_overview task start working ......")
		db.Find(&K8sCluster)
		Data["clusters"] = len(K8sCluster)
		Data["clusterName"] = []string{}
		if len(K8sCluster) > 0 {
			for _, kc := range K8sCluster {
				Data[kc.K8sName] = map[string]string{}
				Data["clusterName"] = append(Data["clusterName"].([]string), kc.K8sName)
				k8s := common.K8sCluster{Name: kc.K8sName, Config: kc.K8sConfig}
				Data[kc.K8sName].(map[string]string)["nodes"] = strconv.Itoa(len(k8s.ListNodes()))
				var pods int
				for _, n := range k8s.ListNodes() {
					pods = pods + len(k8s.GetNodeDetail(n["Name"].(string))["pods"].([]map[string]interface{}))
				}
				Data[kc.K8sName].(map[string]string)["pods"] = strconv.Itoa(pods)
				Data[kc.K8sName].(map[string]string)["nameSpace"] = strconv.Itoa(len(k8s.ListNameSpaces()))
				Subject, Expires, DaysLeft := common.CheckAPIExpires(k8s.ApiHost)
				Data[kc.K8sName].(map[string]string)["subject"] = Subject
				Data[kc.K8sName].(map[string]string)["expires"] = Expires
				Data[kc.K8sName].(map[string]string)["daysLeft"] = strconv.Itoa(DaysLeft)
			}
		}
		rc.Set(ctx, "k8s_overview", kits.MapToJson(Data), 5*time.Minute)
	}
}
