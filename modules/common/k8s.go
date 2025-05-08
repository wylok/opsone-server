package common

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/fileutil"
	"github.com/duke-git/lancet/v2/slice"
	"inner/modules/kits"
	va1 "k8s.io/api/apps/v1"
	sv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

func CheckAPIExpires(apiServerURL string) (Subject, Expires string, DaysLeft int) {
	// 创建自定义TLS配置
	conf := &tls.Config{
		InsecureSkipVerify: true, // 跳过证书验证（仅用于获取证书）
	}
	// 建立TLS连接
	apiServerURL = strings.Replace(apiServerURL, "https://", "", -1)
	conn, err := tls.Dial("tcp", apiServerURL, conf)
	defer func(conn *tls.Conn) {
		err = conn.Close()
	}(conn)
	if err == nil {
		// 获取证书链
		certs := conn.ConnectionState().PeerCertificates
		if len(certs) == 0 {
			Log.Error(apiServerURL + " No certificates found")
		} else {
			cert := certs[0]
			expiry := cert.NotAfter
			DaysLeft = int(time.Until(expiry).Hours() / 24)
			Subject = cert.Subject.CommonName
			Expires = expiry.Format("2006-01-02")
		}
	} else {
		Log.Error(err)
	}
	return
}

type K8sCluster struct {
	Name    string
	Config  string
	ApiHost string
}

func (k8s *K8sCluster) RestConfig() *rest.Config {
	path := "/root/.kube/"
	file := path + "config-" + k8s.Name
	_ = os.MkdirAll(path, 0751)
	if !fileutil.IsExist(file) {
		err = os.WriteFile(file, []byte(k8s.Config), 0600)
	}
	if err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", file)
		if err == nil {
			k8s.ApiHost = config.Host
			return config
		}
	}
	return nil
}
func (k8s *K8sCluster) Client() *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(k8s.RestConfig())
	if err == nil {
		return client
	} else {
		Log.Error(err)
	}
	return nil
}
func (k8s *K8sCluster) ListNodes() []map[string]interface{} {
	var Nodes []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var InternalIP string
				var status string
				for _, addr := range n.Status.Addresses {
					if addr.Type == "InternalIP" {
						InternalIP = addr.Address
						break
					}
				}
				for _, s := range n.Status.Conditions {
					if s.Type == "Ready" {
						status = string(s.Status)
					}
				}
				_, master := n.Labels["node-role.kubernetes.io/master"]
				Nodes = append(Nodes, map[string]interface{}{"Name": n.Name, "CreationTimestamp": n.CreationTimestamp,
					"InternalIP": InternalIP, "nodeInfo": n.Status.NodeInfo, "Status": status, "master": master,
					"taints": n.Spec.Taints})
			}
		} else {
			Log.Error(err)
		}
	}
	return Nodes
}
func (k8s *K8sCluster) ListNameSpaces() []map[string]interface{} {
	var Namespaces []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Namespaces = append(Namespaces, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp, "Phase": n.Status.Phase})
			}
		} else {
			Log.Error(err)
		}
	}
	return Namespaces
}
func (k8s *K8sCluster) ListDeployments(nameSpace string) []map[string]interface{} {
	var Deployments []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().AppsV1().Deployments(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var container v1.Container
				for _, c := range n.Spec.Template.Spec.Containers {
					if strings.Contains(c.Name, n.Name) {
						container = c
						break
					}
				}
				Deployments = append(Deployments, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"UpdatedReplicas": n.Status.UpdatedReplicas, "Replicas": n.Status.Replicas,
					"AvailableReplicas": n.Status.AvailableReplicas, "ReadyReplicas": n.Status.ReadyReplicas,
					"Container": container})
			}
		} else {
			Log.Error(err)
		}
	}
	return Deployments
}
func (k8s *K8sCluster) ListReplicaSets(nameSpace string) []va1.ReplicaSet {
	object, _ := k8s.Client().AppsV1().ReplicaSets(nameSpace).List(ctx, metav1.ListOptions{})
	return object.Items
}
func (k8s *K8sCluster) ListDaemonSets(nameSpace string) []map[string]interface{} {
	var DaemonSets []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().AppsV1().DaemonSets(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var container v1.Container
				for _, c := range n.Spec.Template.Spec.Containers {
					if strings.Contains(c.Name, n.Name) {
						container = c
						break
					}
				}
				DaemonSets = append(DaemonSets, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"AvailableReplicas": n.Status.NumberAvailable, "ReadyReplicas": n.Status.NumberReady,
					"Replicas": n.Status.CurrentNumberScheduled, "Container": container,
					"UpdatedReplicas": n.Status.UpdatedNumberScheduled})
			}
		} else {
			Log.Error(err)
		}
	}
	return DaemonSets
}
func (k8s *K8sCluster) ListStatefulSets(nameSpace string) []map[string]interface{} {
	var StatefulSets []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().AppsV1().StatefulSets(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var container v1.Container
				for _, c := range n.Spec.Template.Spec.Containers {
					if strings.Contains(c.Name, n.Name) {
						container = c
						break
					}
				}
				StatefulSets = append(StatefulSets, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"AvailableReplicas": n.Status.AvailableReplicas, "Replicas": n.Status.Replicas,
					"ReadyReplicas": n.Status.ReadyReplicas, "UpdatedReplicas": n.Status.UpdatedReplicas,
					"Container": container})
			}
		} else {
			Log.Error(err)
		}
	}
	return StatefulSets
}
func (k8s *K8sCluster) ListServices(nameSpace string) []map[string]interface{} {
	var Services []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Services(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Services = append(Services, map[string]interface{}{"name": n.Name, "CreationTime": n.CreationTimestamp,
					"ClusterIP": n.Spec.ClusterIP, "Selector": n.Spec.Selector, "TargetPort": n.Spec.Ports[0].TargetPort,
					"NodePort": n.Spec.Ports[0].NodePort, "Protocol": n.Spec.Ports[0].Protocol, "type": n.Spec.Type})
			}
		} else {
			Log.Error(err)
		}
	}
	return Services
}
func (k8s *K8sCluster) ListIngress(nameSpace string) []map[string]interface{} {
	var Ingresses []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().ExtensionsV1beta1().Ingresses(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Ingresses = append(Ingresses, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"LoadBalancer": n.Status.LoadBalancer, "Rules": n.Spec.Rules})
			}
		} else {
			Log.Error(err)
		}
	}
	return Ingresses
}
func (k8s *K8sCluster) ListEndpoints(nameSpace string) []map[string]interface{} {
	var Endpoints []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Endpoints(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var address []string
				for _, sub := range n.Subsets {
					for _, addr := range sub.Addresses {
						address = append(address, addr.IP+":"+fmt.Sprint(sub.Ports[0].Port))
					}
				}
				Endpoints = append(Endpoints, map[string]interface{}{"name": n.Name, "CreationTime": n.CreationTimestamp,
					"address": address})
			}
		} else {
			Log.Error(err)
		}
	}
	return Endpoints
}
func (k8s *K8sCluster) ListEvents(nameSpace string) []map[string]interface{} {
	var Events []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Events(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Events = append(Events, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Message": n.Message, "Reason": n.Reason})
			}
		} else {
			Log.Error(err)
		}
	}
	return Events
}
func (k8s *K8sCluster) ListConfigMaps(nameSpace string) []map[string]interface{} {
	var ConfigMaps []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().ConfigMaps(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var stages []map[string]string
				for k, v := range n.Data {
					stages = append(stages, map[string]string{"key": k, "value": v})
				}
				ConfigMaps = append(ConfigMaps, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Data": n.Data, "Stages": stages})
			}
		} else {
			Log.Error(err)
		}
	}
	return ConfigMaps
}
func (k8s *K8sCluster) ListSecrets(nameSpace string) []map[string]interface{} {
	var Secrets []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().Secrets(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Secrets = append(Secrets, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Data": n.Data})
			}
		} else {
			Log.Error(err)
		}
	}
	return Secrets
}
func (k8s *K8sCluster) ListJobs(nameSpace string) []map[string]interface{} {
	var Jobs []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().BatchV1().Jobs(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				Jobs = append(Jobs, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Active": n.Status.Active, "StartTime": n.Status.StartTime, "CompletionTime": n.Status.CompletionTime,
					"Succeeded": n.Status.Succeeded})
			}
		} else {
			Log.Error(err)
		}
	}
	return Jobs
}
func (k8s *K8sCluster) ListCronJobs(nameSpace string) []map[string]interface{} {
	var CronJobs []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().BatchV1().CronJobs(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				CronJobs = append(CronJobs, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"LastScheduleTime": n.Status.LastScheduleTime, "LastSuccessfulTime": n.Status.LastSuccessfulTime,
					"Schedule": n.Spec.Schedule, "Suspend": n.Spec.Suspend})
			}
		} else {
			Log.Error(err)
		}
	}
	return CronJobs
}
func (k8s *K8sCluster) ListLimitRanges(nameSpace string) []map[string]interface{} {
	var LimitRanges []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().LimitRanges(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				LimitRanges = append(LimitRanges, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Limits": n.Spec.Limits})
			}
		} else {
			Log.Error(err)
		}
	}
	return LimitRanges
}
func (k8s *K8sCluster) ListClusterRoles() []map[string]interface{} {
	var ClusterRoles []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				ClusterRoles = append(ClusterRoles, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Rules": n.Rules, "Labels": n.Labels, "AggregationRule": n.AggregationRule.Size()})
			}
		} else {
			Log.Error(err)
		}
	}
	return ClusterRoles
}
func (k8s *K8sCluster) ListClusterRoleBindings() []map[string]interface{} {
	var ClusterRoleBindings []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				ClusterRoleBindings = append(ClusterRoleBindings, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Subjects": n.Subjects})
			}
		} else {
			Log.Error(err)
		}
	}
	return ClusterRoleBindings
}
func (k8s *K8sCluster) ListServiceAccounts(nameSpace string) []map[string]interface{} {
	var ServiceAccounts []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().ServiceAccounts(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				ServiceAccounts = append(ServiceAccounts, map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"Secrets": n.Secrets})
			}
		} else {
			Log.Error(err)
		}
	}
	return ServiceAccounts
}
func (k8s *K8sCluster) ListAutoscalers(nameSpace, deployment string) []map[string]interface{} {
	var Autoscalers []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().AutoscalingV1().HorizontalPodAutoscalers(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				m := map[string]interface{}{"Name": n.Name, "CreationTime": n.CreationTimestamp,
					"MinReplicas": n.Spec.MinReplicas, "MaxReplicas": n.Spec.MaxReplicas,
					"CPUUtilizationPercentage":        n.Spec.TargetCPUUtilizationPercentage,
					"CurrentCPUUtilizationPercentage": n.Status.CurrentCPUUtilizationPercentage,
					"CurrentReplicas":                 n.Status.CurrentReplicas, "DesiredReplicas": n.Status.DesiredReplicas,
					"LastScaleTime": n.Status.LastScaleTime}
				if deployment != "" && deployment == n.Name {
					Autoscalers = append(Autoscalers, m)
					break
				} else {
					Autoscalers = append(Autoscalers, m)
				}
			}
		} else {
			Log.Error(err)
		}
	}
	return Autoscalers
}
func (k8s *K8sCluster) ListPods(nameSpace, deployment, daemonSet, StatefulSet string) []map[string]interface{} {
	var Pods []map[string]interface{}
	if k8s.Client() != nil {
		var podNames []string
		if deployment != "" {
			list, err := k8s.Client().AppsV1().Deployments(nameSpace).List(ctx, metav1.ListOptions{})
			if err == nil {
				for _, n := range list.Items {
					if n.Name == deployment {
						for _, c := range n.Spec.Template.Spec.Containers {
							podNames = append(podNames, c.Name)
						}
					}
				}
			} else {
				Log.Error(err)
			}
		}
		if daemonSet != "" {
			list, err := k8s.Client().AppsV1().DaemonSets(nameSpace).List(ctx, metav1.ListOptions{})
			if err == nil {
				for _, n := range list.Items {
					if n.Name == daemonSet {
						for _, c := range n.Spec.Template.Spec.Containers {
							podNames = append(podNames, c.Name)
						}
					}
				}
			}
		}
		if StatefulSet != "" {
			list, err := k8s.Client().AppsV1().StatefulSets(nameSpace).List(ctx, metav1.ListOptions{})
			if err == nil {
				for _, n := range list.Items {
					if n.Name == StatefulSet {
						for _, c := range n.Spec.Template.Spec.Containers {
							podNames = append(podNames, c.Name)
						}
					}
				}
			} else {
				Log.Error(err)
			}
		}
		pod, err := k8s.Client().CoreV1().Pods(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, p := range pod.Items {
				for _, name := range podNames {
					if strings.Contains(p.Name, name+"-") {
						stat := "Ready"
						for _, c := range p.Status.ContainerStatuses {
							if !c.Ready {
								stat = "Not Ready"
								break
							}
						}
						m := map[string]interface{}{"Name": p.Name, "Stat": stat, "PodIP": p.Status.PodIP, "HostIP": p.Status.HostIP,
							"StartTime": p.Status.StartTime, "images": map[string][]string{}, "namespace": p.Namespace,
							"NodeName": p.Spec.NodeName, "resources": []interface{}{}}
						for _, c := range p.Status.ContainerStatuses {
							m["images"].(map[string][]string)[c.Name] = strings.Split(c.Image, ":")
						}
						for _, r := range p.Spec.Containers {
							m["resources"] = append(m["resources"].([]interface{}), r.Resources)
						}
						Pods = append(Pods, m)
					}
				}
			}
		} else {
			Log.Error(err)
		}
	}
	return Pods
}
func (k8s *K8sCluster) GetPodEvent(nameSpace, name string) interface{} {
	var podEvents []map[string]interface{}
	if k8s.Client() != nil {
		events, _ := k8s.Client().CoreV1().Events(nameSpace).List(ctx, metav1.ListOptions{})
		for _, item := range events.Items {
			if strings.Contains(item.Name, name+".") {
				var Time time.Time
				if !item.EventTime.IsZero() {
					Time = item.EventTime.Time
				} else {
					if !item.FirstTimestamp.IsZero() {
						Time = item.FirstTimestamp.Time
					}
					if !item.LastTimestamp.IsZero() {
						Time = item.LastTimestamp.Time
					}
				}
				podEvents = append(podEvents, map[string]interface{}{
					"Reason": item.Reason, "Message": item.Message, "Count": item.Count, "Time": Time, "Type": item.Type})
			}
		}
	}
	slice.Reverse(podEvents)
	return podEvents
}
func (k8s *K8sCluster) GetNodeDetail(NodeName string) map[string]interface{} {
	var (
		requestsCpu int64
		limitsCpu   int64
		requestsMem int64
		limitsMem   int64
		pods        []map[string]interface{}
	)
	node, _ := k8s.Client().CoreV1().Nodes().Get(ctx, NodeName, metav1.GetOptions{})
	for _, n := range k8s.ListNameSpaces() {
		_, ok := n["Name"]
		if ok {
			pod, err := k8s.Client().CoreV1().Pods(n["Name"].(string)).List(ctx, metav1.ListOptions{})
			if err == nil {
				for _, p := range pod.Items {
					if NodeName == p.Spec.NodeName {
						m := map[string]interface{}{"Name": p.Name, "Stat": p.Status.Phase, "PodIP": p.Status.PodIP,
							"StartTime": p.Status.StartTime, "images": []string{}, "namespace": p.Namespace,
							"RestartPolicy": p.Spec.RestartPolicy, "resources": []interface{}{}}
						for _, c := range p.Status.ContainerStatuses {
							m["images"] = append(m["images"].([]string), c.Image)
						}
						for _, r := range p.Spec.Containers {
							m["resources"] = append(m["resources"].([]interface{}), r.Resources)
							cpu, _ := r.Resources.Requests.Cpu().AsInt64()
							mem, _ := r.Resources.Requests.Memory().AsInt64()
							requestsCpu = requestsCpu + cpu
							requestsMem = requestsMem + mem
							cpu, _ = r.Resources.Limits.Cpu().AsInt64()
							mem, _ = r.Resources.Limits.Memory().AsInt64()
							limitsCpu = limitsCpu + cpu
							limitsMem = limitsMem + mem
						}
						pods = append(pods, m)
					}
				}
			} else {
				Log.Error(err)
			}
		}
	}
	return map[string]interface{}{"node": node, "pods": pods, "taints": node.Spec.Taints,
		"requests": map[string]int64{"cpu": requestsCpu, "mem": requestsMem / 1024 / 1000 / 1000},
		"limits":   map[string]int64{"cpu": limitsCpu, "mem": limitsMem / 1024 / 1000 / 1000}}
}
func (k8s *K8sCluster) GetNodeYaml(NodeName string) string {
	object, _ := k8s.Client().CoreV1().Nodes().Get(ctx, NodeName, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "v1", "kind": "Node",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return string(YAML)
}
func (k8s *K8sCluster) GetDeploymentYaml(nameSpace, deployment string) string {
	object, _ := k8s.Client().AppsV1().Deployments(nameSpace).Get(ctx, deployment, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) GetDaemonSetYaml(nameSpace, daemonSet string) string {
	object, _ := k8s.Client().AppsV1().DaemonSets(nameSpace).Get(ctx, daemonSet, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "apps/v1", "kind": "DaemonSet",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) GetStatefulSetYaml(nameSpace, statefulSet string) string {
	object, _ := k8s.Client().AppsV1().StatefulSets(nameSpace).Get(ctx, statefulSet, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "apps/v1", "kind": "StatefulSet",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) GetServicesYaml(nameSpace, service string) string {
	object, _ := k8s.Client().CoreV1().Services(nameSpace).Get(ctx, service, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "v1", "kind": "Service",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) GetConfigMapYaml(nameSpace, configMap string) string {
	object, _ := k8s.Client().CoreV1().ConfigMaps(nameSpace).Get(ctx, configMap, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "v1", "kind": "ConfigMap",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"data":     object.Data})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) GetPodYaml(nameSpace, pod string) string {
	object, _ := k8s.Client().CoreV1().Pods(nameSpace).Get(ctx, pod, metav1.GetOptions{})
	t, _ := json.Marshal(map[string]interface{}{"apiVersion": "v1", "kind": "Pod",
		"metadata": map[string]interface{}{"annotations": object.GetAnnotations(), "labels": object.GetLabels()},
		"spec":     object.Spec})
	YAML, _ := yaml.JSONToYAML(t)
	return strings.TrimSpace(string(YAML))
}
func (k8s *K8sCluster) RestartDeployment(nameSpace, name string) error {
	for _, p := range k8s.ListPods(nameSpace, name, "", "") {
		err = k8s.DelPod(nameSpace, p["Name"].(string))
	}
	return err
}
func (k8s *K8sCluster) RestartDaemonSet(nameSpace, name string) error {
	for _, p := range k8s.ListPods(nameSpace, "", name, "") {
		err = k8s.DelPod(nameSpace, p["Name"].(string))
	}
	return err
}
func (k8s *K8sCluster) RestartStatefulSet(nameSpace, name string) error {
	for _, p := range k8s.ListPods(nameSpace, "", "", name) {
		err = k8s.DelPod(nameSpace, p["Name"].(string))
	}
	return err
}
func (k8s *K8sCluster) UpdateAutoscaler(nameSpace, name string, sc map[string]*int32) error {
	scaler, err := k8s.Client().AutoscalingV1().HorizontalPodAutoscalers(nameSpace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		var hpa *sv1.HorizontalPodAutoscaler
		hpa.Name = name
		hpa.Spec.MinReplicas = sc["MinReplicas"]
		hpa.Spec.MaxReplicas = *sc["MaxReplicas"]
		hpa.Spec.TargetCPUUtilizationPercentage = sc["TargetCPUUtilizationPercentage"]
		scaler, err = k8s.Client().AutoscalingV1().HorizontalPodAutoscalers(nameSpace).Create(ctx, scaler, metav1.CreateOptions{})
		return err
	}
	scaler.Spec.MinReplicas = sc["MinReplicas"]
	scaler.Spec.MaxReplicas = *sc["MaxReplicas"]
	scaler.Spec.TargetCPUUtilizationPercentage = sc["TargetCPUUtilizationPercentage"]
	_, err = k8s.Client().AutoscalingV1().HorizontalPodAutoscalers(nameSpace).Update(ctx, scaler, metav1.UpdateOptions{})
	return err
}
func (k8s *K8sCluster) UpdateDeployment(nameSpace, name string, images map[string][]string) error {
	dep, err := k8s.Client().AppsV1().Deployments(nameSpace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		for index, c := range dep.Spec.Template.Spec.Containers {
			_, ok := images[c.Name]
			if ok {
				dep.Spec.Template.Spec.Containers[index].Image = strings.Join(images[c.Name], ":")
			}
		}
	}
	_, err = k8s.Client().AppsV1().Deployments(nameSpace).Update(ctx, dep, metav1.UpdateOptions{})
	return err
}
func (k8s *K8sCluster) UpdateDaemonSet(nameSpace, name string, images map[string][]string) error {
	dep, err := k8s.Client().AppsV1().DaemonSets(nameSpace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		for index, c := range dep.Spec.Template.Spec.Containers {
			_, ok := images[c.Name]
			if ok {
				dep.Spec.Template.Spec.Containers[index].Image = strings.Join(images[c.Name], ":")
			}
		}
	}
	_, err = k8s.Client().AppsV1().DaemonSets(nameSpace).Update(ctx, dep, metav1.UpdateOptions{})
	return err
}
func (k8s *K8sCluster) TaintNode(NodeName, Effect string) error {
	node, err := k8s.Client().CoreV1().Nodes().Get(ctx, NodeName, metav1.GetOptions{})
	var Taints []v1.Taint
	if err == nil {
		node.Spec.Unschedulable = false
		if Effect != "" {
			node.Spec.Unschedulable = true
			Taints = append(Taints, v1.Taint{Key: "node.kubernetes.io/unschedulable", Value: "", Effect: v1.TaintEffect(Effect)})
		}
		node.Spec.Taints = Taints
		_, err = k8s.Client().CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	}
	return err
}
func (k8s *K8sCluster) UpdateNode(NodeName string, Labels map[string]string) error {
	node, err := k8s.Client().CoreV1().Nodes().Get(ctx, NodeName, metav1.GetOptions{})
	if err == nil {
		if Labels != nil {
			node.Labels = Labels
		}
	}
	_, err = k8s.Client().CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}
func (k8s *K8sCluster) UpdateConfigMap(nameSpace, name string, data map[string]string) error {
	configMap, err := k8s.Client().CoreV1().ConfigMaps(nameSpace).Get(ctx, name, metav1.GetOptions{})
	configMap.Data = data
	_, err = k8s.Client().CoreV1().ConfigMaps(nameSpace).Update(ctx, configMap, metav1.UpdateOptions{})
	return err
}
func (k8s *K8sCluster) AddNameSpace(namespace string) error {
	_, err = k8s.Client().CoreV1().Namespaces().Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
	return err
}
func (k8s *K8sCluster) AddSecret(nameSpace, name, metaType string, data map[string]string) error {
	secret, err := k8s.Client().CoreV1().Secrets(nameSpace).Get(ctx, name, metav1.GetOptions{})
	secret.Name = name
	secret.Type = v1.SecretType(metaType)
	secret.Namespace = nameSpace
	secret.StringData = data
	_, err = k8s.Client().CoreV1().Secrets(nameSpace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}
func (k8s *K8sCluster) AddConfigMap(nameSpace, name string, data map[string]string) error {
	configMap, err := k8s.Client().CoreV1().ConfigMaps(nameSpace).Get(ctx, name, metav1.GetOptions{})
	configMap.Name = name
	configMap.Namespace = nameSpace
	configMap.Data = data
	_, err = k8s.Client().CoreV1().ConfigMaps(nameSpace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}
func (k8s *K8sCluster) DelPod(nameSpace, name string) error {
	err = k8s.Client().CoreV1().Pods(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelNode(NodeName string) error {
	err = k8s.Client().CoreV1().Nodes().Delete(ctx, NodeName, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelNameSpace(namespace string) error {
	err = k8s.Client().CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelDeployment(nameSpace, name string) error {
	err = k8s.Client().AppsV1().Deployments(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelDaemonSet(nameSpace, name string) error {
	err = k8s.Client().AppsV1().DaemonSets(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelService(nameSpace, name string) error {
	err = k8s.Client().CoreV1().Services(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelStatefulSet(nameSpace, name string) error {
	err = k8s.Client().AppsV1().StatefulSets(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DeleteAutoscaler(nameSpace, name string) error {
	err = k8s.Client().AutoscalingV1().HorizontalPodAutoscalers(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelConfigMap(nameSpace, name string) error {
	err = k8s.Client().CoreV1().ConfigMaps(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelSecret(nameSpace, name string) error {
	err = k8s.Client().CoreV1().Secrets(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelEndpoint(nameSpace, name string) error {
	err = k8s.Client().CoreV1().Endpoints(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelEvent(nameSpace, name string) error {
	err = k8s.Client().CoreV1().Events(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelJob(nameSpace, name string) error {
	err = k8s.Client().BatchV1().Jobs(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelCronJob(nameSpace, name string) error {
	err = k8s.Client().BatchV1().CronJobs(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelIngress(nameSpace, name string) error {
	err = k8s.Client().ExtensionsV1beta1().Ingresses(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelLimitRange(nameSpace, name string) error {
	err = k8s.Client().CoreV1().LimitRanges(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) DelServiceAccount(nameSpace, name string) error {
	err = k8s.Client().CoreV1().ServiceAccounts(nameSpace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *K8sCluster) ListPersistentVolumes() []map[string]interface{} {
	var PersistentVolumes []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				PersistentVolumes = append(PersistentVolumes, map[string]interface{}{"Name": n.Name,
					"CreationTimestamp": n.CreationTimestamp, "Status": n.Status,
					"Spec": n.Spec})
			}
		}
	}
	return PersistentVolumes
}
func (k8s *K8sCluster) ListPersistentVolumeClaims(nameSpace string) []map[string]interface{} {
	var PersistentVolumeClaims []map[string]interface{}
	if k8s.Client() != nil {
		list, err := k8s.Client().CoreV1().PersistentVolumeClaims(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				PersistentVolumeClaims = append(PersistentVolumeClaims, map[string]interface{}{"Name": n.Name,
					"CreationTimestamp": n.CreationTimestamp, "Status": n.Status,
					"Spec": n.Spec})
			}
		}
	}
	return PersistentVolumeClaims
}
func (k8s *K8sCluster) ListNodeMetrics() map[string]interface{} {
	items := map[string]interface{}{}
	mc, err := metrics.NewForConfig(k8s.RestConfig())
	if err == nil {
		list, err := mc.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				m, err := json.Marshal(n.Usage)
				if err == nil {
					items[n.Name] = kits.StringToMap(string(m))
				}
			}
		}
	}
	return items
}
func (k8s *K8sCluster) ListPodMetrics(nameSpace string) map[string]interface{} {
	items := map[string]interface{}{}
	mc, err := metrics.NewForConfig(k8s.RestConfig())
	if err == nil {
		list, err := mc.MetricsV1beta1().PodMetricses(nameSpace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, n := range list.Items {
				var con []map[string]interface{}
				for _, c := range n.Containers {
					m, err := json.Marshal(c.Usage)
					if err == nil {
						con = append(con, kits.StringToMap(string(m)))
					}
				}
				items[n.Name] = con
			}
		}
	}
	return items
}
