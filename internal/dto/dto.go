package dto

/*
*********************
/ Workload **********
*********************
*/
type DeploymentInfo struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Pods       string   `json:"pods"`
	Replicas   int32    `json:"replicas"`
	Age        string   `json:"age"`
	Conditions []string `json:"conditions"`
}

type PodInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Containers   int    `json:"containers"`
	CPU          string `json:"cpu"`
	Memory       string `json:"memory"`
	Restarts     int32  `json:"restarts"`
	ControlledBy string `json:"controlledBy"`
	Node         string `json:"node"`
	QoS          string `json:"qos"`
	Age          string `json:"age"`
	Status       string `json:"status"`
}

type DaemonSetInfo struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	Pods         string   `json:"pods"`
	NodeSelector []string `json:"nodeSelector"`
	Age          string   `json:"age"`
}

type StatefulSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Pods      string `json:"pods"`
	Replicas  int32  `json:"replicas"`
	Age       string `json:"age"`
}

type ReplicaSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Desired   int32  `json:"desired"`
	Current   int32  `json:"current"`
	Ready     int32  `json:"ready"`
	Age       string `json:"age"`
}

type ReplicationControllerInfo struct {
	Name           string   `json:"name"`
	Namespace      string   `json:"namespace"`
	Replica        int32    `json:"replica"`
	DesiredReplica int32    `json:"desiredReplica"`
	Selector       []string `json:"selector"`
}

type CronJobInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Schedule     string `json:"schedule"`
	Suspend      bool   `json:"suspend"`
	Active       int    `json:"active"`
	LastSchedule string `json:"lastSchedule"`
	Age          string `json:"age"`
}

type JobInfo struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Completion string   `json:"completion"`
	Age        string   `json:"age"`
	Conditions []string `json:"conditions"`
}

/*
*********************
/ Config ************
*********************
*/
type ConfigMapInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

type SecretInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Keys      []string          `json:"keys"`
	Type      string            `json:"type"`
	Age       string            `json:"age"`
}

type HPAInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Metrics   []string `json:"metrics"`
	MinPods   int32    `json:"minPods"`
	MaxPods   int32    `json:"maxPods"`
	Replicas  int32    `json:"replicas"`
	Age       string   `json:"age"`
	Status    string   `json:"status"`
}

type ResourceQuotaInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
}

type LimitRangeInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
}

type VPAInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Mode      string `json:"mode"`
	Status    string `json:"status"`
}

type PDBInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Age            string `json:"age"`
	MinAvailable   string `json:"minAvailable"`
	MaxUnavailable string `json:"maxUnavailable"`
	CurrentHealthy int32  `json:"currentHealthy"`
	DesiredHealthy int32  `json:"desiredHealthy"`
}

type PriorityClassInfo struct {
	Name          string `json:"name"`
	Age           string `json:"age"`
	Value         int32  `json:"value"`
	GlobalDefault bool   `json:"globalDefault"`
}

type RuntimeClassInfo struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Age     string `json:"age"`
}

type LeaseInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Holder    string `json:"holder"`
}

/*
*********************
/ Network ***********
*********************
*/
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Age        string            `json:"age"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"clusterIP"`
	Ports      string            `json:"ports"`
	ExternalIP []string          `json:"externalIP"`
	Selector   map[string]string `json:"selector"`
	Status     string            `json:"status"`
}

type EndpointInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Age       string   `json:"age"`
	Endpoints []string `json:"endpoints"`
}

type IngressInfo struct {
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	Age           string   `json:"age"`
	LoadBalancers []string `json:"loadBalancers"`
	Rules         []string `json:"rules"`
}

type IngressClassInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Controller string `json:"controller"`
	APIGroup   string `json:"apiGroup"`
	Scope      string `json:"scope"`
	Kind       string `json:"kind"`
}

type NetworkPolicyInfo struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Age         string   `json:"age"`
	PolicyTypes []string `json:"policyTypes"`
}

/*
*********************
/ Storage ***********
*********************
*/
type PVCInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	StorageClass string `json:"storageClass"`
	Size         string `json:"size"`
	Pods         string `json:"pods"`
	Age          string `json:"age"`
	Status       string `json:"status"`
}

type PVInfo struct {
	Name         string `json:"name"`
	StorageClass string `json:"storageClass"`
	Capacity     string `json:"capacity"`
	Claim        string `json:"claim"`
	Age          string `json:"age"`
	Status       string `json:"status"`
}

type StorageClassInfo struct {
	Name          string `json:"name"`
	Provisioner   string `json:"provisioner"`
	ReclaimPolicy string `json:"reclaimPolicy"`
	Default       bool   `json:"default"`
	Age           string `json:"age"`
}

type EventInfo struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	Namespace      string `json:"namespace"`
	InvolvedObject string `json:"involvedObject"`
	Source         string `json:"source"`
	Count          int32  `json:"count"`
	Age            string `json:"age"`
	LastSeen       string `json:"lastSeen"`
}

type NamespaceInfo struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Age    string            `json:"age"`
	Status string            `json:"status"`
}

type NodeInfo struct {
	Name       string          `json:"name"`
	CPU        string          `json:"cpu"`
	Memory     string          `json:"memory"`
	Disk       string          `json:"disk"`
	Taints     []string        `json:"taints"`
	Roles      []string        `json:"roles"`
	Version    string          `json:"version"`
	Age        string          `json:"age"`
	Conditions []NodeCondition `json:"conditions"`
}

type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

/*
*********************
/ Access Control ****
*********************
*/

type ServiceAccountInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
}

type ClusterRoleInfo struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

type RoleInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
}

type ClusterRoleBindingInfo struct {
	Name     string   `json:"name"`
	Bindings []string `json:"bindings"`
	Age      string   `json:"age"`
}

type RoleBindingInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Bindings  []string `json:"bindings"`
	Age       string   `json:"age"`
}
