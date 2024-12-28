package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	dto "agentkube.com/agent-kube-operator/internal/dto"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	autoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListController struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewListController(client client.Client, scheme *runtime.Scheme) *ListController {
	return &ListController{
		client: client,
		scheme: scheme,
	}
}

type ResourceRequest struct {
	Namespaces []string `json:"namespaces"`
	Resource   string   `json:"resource"`
}

func calculateAge(timestamp metav1.Time) string {
	return time.Since(timestamp.Time).Round(time.Second).String()
}

func (c *ListController) ListResources(ctx context.Context, req ResourceRequest) (interface{}, error) {
	switch req.Resource {
	/* Workloads */
	case "deployments", "deployment", "deploy":
		return c.listDeployments(ctx, req.Namespaces)
	case "pods", "pod", "po":
		return c.listPods(ctx, req.Namespaces)
	case "daemonsets", "ds":
		return c.listDaemonSets(ctx, req.Namespaces)
	case "statefulsets", "sts":
		return c.listStatefulSets(ctx, req.Namespaces)
	case "replicasets", "rs":
		return c.listReplicaSets(ctx, req.Namespaces)
	case "replicationcontrollers", "replicationcontroller", "rc":
		return c.listReplicationControllers(ctx, req.Namespaces)
	case "cronjobs", "cronjob":
		return c.listCronJobs(ctx, req.Namespaces)
	case "jobs", "job":
		return c.listJobs(ctx, req.Namespaces)

	/* Config */
	case "configmaps", "cm":
		return c.listConfigMaps(ctx, req.Namespaces)
	case "secrets", "secret":
		return c.listSecrets(ctx, req.Namespaces)
	case "horizontalpodautoscaler", "hpa":
		return c.listHPA(ctx, req.Namespaces)
	case "resourcequotas", "quota":
		return c.listResourceQuotas(ctx, req.Namespaces)
	case "limitranges":
		return c.listLimitRanges(ctx, req.Namespaces)
	case "verticalpodautoscaler", "vpa":
		return c.listVPA(ctx, req.Namespaces)
	case "poddisruptionbudget", "pdbs", "pdb":
		return c.listPDBs(ctx, req.Namespaces)
	case "priorityclasses", "priorityclass":
		return c.listPriorityClasses(ctx)
	case "runtimeclasses":
		return c.listRuntimeClasses(ctx)
	case "leases":
		return c.listLeases(ctx, req.Namespaces)
	case "mutatingwebhookconfigurations", "mutatingwebhook":
		return c.listMutatingWebhooks(ctx)
	case "validatingwebhookconfigurations", "validatingwebhook":
		return c.listValidatingWebhooks(ctx)

	/* Network */
	case "services", "svc":
		return c.listServices(ctx, req.Namespaces)
	case "endpoints", "endpoint", "ep":
		return c.listEndpoints(ctx, req.Namespaces)
	case "ingresses", "ingress", "ing":
		return c.listIngresses(ctx, req.Namespaces)
	case "ingressclasses", "ingressclass":
		return c.listIngressClasses(ctx)
	case "networkpolicies", "networkpolicy", "netpol":
		return c.listNetworkPolicies(ctx, req.Namespaces)
	/* Storage */
	case "persistentvolumeclaims", "pvcs", "pvc":
		return c.listPVCs(ctx, req.Namespaces)
	case "persistentvolume", "pvs", "pv":
		return c.listPVs(ctx)
	case "storageclasses", "storageclass", "sc":
		return c.listStorageClasses(ctx)

	/* Access Control */
	case "serviceaccounts", "sa":
		return c.listServiceAccounts(ctx, req.Namespaces)
	case "clusterroles", "clusterrole", "cr":
		return c.listClusterRoles(ctx)
	case "roles", "role":
		return c.listRoles(ctx, req.Namespaces)
	case "clusterrolebindings", "clusterrolebinding", "crb":
		return c.listClusterRoleBindings(ctx)
	case "rolebindings", "rolebinding", "rb":
		return c.listRoleBindings(ctx, req.Namespaces)

	case "events", "ev":
		return c.listEvents(ctx, req.Namespaces)
	case "namespaces", "ns":
		return c.listNamespaces(ctx)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", req.Resource)
	}
}

func (c *ListController) listDeployments(ctx context.Context, namespaces []string) ([]dto.DeploymentInfo, error) {
	var deployments appsv1.DeploymentList
	var result []dto.DeploymentInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &deployments, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, d := range deployments.Items {
			info := dto.DeploymentInfo{
				Name:      d.Name,
				Namespace: d.Namespace,
				Pods:      fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas),
				Replicas:  d.Status.Replicas,
				Age:       calculateAge(d.CreationTimestamp),
			}

			for _, c := range d.Status.Conditions {
				info.Conditions = append(info.Conditions, string(c.Type))
			}

			result = append(result, info)
		}
	}

	return result, nil
}

func (c *ListController) listPods(ctx context.Context, namespaces []string) ([]dto.PodInfo, error) {
	var pods corev1.PodList
	var result []dto.PodInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &pods, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			var totalRestarts int32
			for _, containerStatus := range pod.Status.ContainerStatuses {
				totalRestarts += containerStatus.RestartCount
			}

			var controlledBy string
			if len(pod.OwnerReferences) > 0 {
				controlledBy = fmt.Sprintf("%s/%s", pod.OwnerReferences[0].Kind, pod.OwnerReferences[0].Name)
			}

			info := dto.PodInfo{
				Name:         pod.Name,
				Namespace:    pod.Namespace,
				Containers:   len(pod.Spec.Containers),
				CPU:          pod.Spec.Containers[0].Resources.Requests.Cpu().String(),
				Memory:       pod.Spec.Containers[0].Resources.Requests.Memory().String(),
				Restarts:     totalRestarts,
				ControlledBy: controlledBy,
				Node:         pod.Spec.NodeName,
				QoS:          string(pod.Status.QOSClass),
				Age:          calculateAge(pod.CreationTimestamp),
				Status:       string(pod.Status.Phase),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listDaemonSets(ctx context.Context, namespaces []string) ([]dto.DaemonSetInfo, error) {
	var daemonsets appsv1.DaemonSetList
	var result []dto.DaemonSetInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &daemonsets, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, ds := range daemonsets.Items {
			var nodeSelectors []string
			for key, value := range ds.Spec.Template.Spec.NodeSelector {
				nodeSelectors = append(nodeSelectors, fmt.Sprintf("%s=%s", key, value))
			}

			info := dto.DaemonSetInfo{
				Name:         ds.Name,
				Namespace:    ds.Namespace,
				Pods:         fmt.Sprintf("%d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled),
				NodeSelector: nodeSelectors,
				Age:          calculateAge(ds.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listStatefulSets(ctx context.Context, namespaces []string) ([]dto.StatefulSetInfo, error) {
	var statefulsets appsv1.StatefulSetList
	var result []dto.StatefulSetInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &statefulsets, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, sts := range statefulsets.Items {
			info := dto.StatefulSetInfo{
				Name:      sts.Name,
				Namespace: sts.Namespace,
				Pods:      fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, sts.Status.Replicas),
				Replicas:  sts.Status.Replicas,
				Age:       calculateAge(sts.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listReplicaSets(ctx context.Context, namespaces []string) ([]dto.ReplicaSetInfo, error) {
	var replicasets appsv1.ReplicaSetList
	var result []dto.ReplicaSetInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &replicasets, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, rs := range replicasets.Items {
			info := dto.ReplicaSetInfo{
				Name:      rs.Name,
				Namespace: rs.Namespace,
				Desired:   *rs.Spec.Replicas,
				Current:   rs.Status.Replicas,
				Ready:     rs.Status.ReadyReplicas,
				Age:       calculateAge(rs.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listReplicationControllers(ctx context.Context, namespaces []string) ([]dto.ReplicationControllerInfo, error) {
	var rcs corev1.ReplicationControllerList
	var result []dto.ReplicationControllerInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &rcs, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, rc := range rcs.Items {
			var selectorStrings []string
			for key, value := range rc.Spec.Selector {
				selectorStrings = append(selectorStrings, fmt.Sprintf("%s=%s", key, value))
			}

			info := dto.ReplicationControllerInfo{
				Name:           rc.Name,
				Namespace:      rc.Namespace,
				Replica:        rc.Status.Replicas,
				DesiredReplica: *rc.Spec.Replicas,
				Selector:       selectorStrings,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listJobs(ctx context.Context, namespaces []string) ([]dto.JobInfo, error) {
	var jobs batchv1.JobList
	var result []dto.JobInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &jobs, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, job := range jobs.Items {
			completion := fmt.Sprintf("%d/%d", job.Status.Succeeded, *job.Spec.Completions)

			info := dto.JobInfo{
				Name:       job.Name,
				Namespace:  job.Namespace,
				Completion: completion,
				Age:        calculateAge(job.CreationTimestamp),
			}

			for _, condition := range job.Status.Conditions {
				info.Conditions = append(info.Conditions, string(condition.Type))
			}

			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listConfigMaps(ctx context.Context, namespaces []string) ([]dto.ConfigMapInfo, error) {
	var cms corev1.ConfigMapList
	var result []dto.ConfigMapInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &cms, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, cm := range cms.Items {
			var keys []string
			for k := range cm.Data {
				keys = append(keys, k)
			}

			info := dto.ConfigMapInfo{
				Name:      cm.Name,
				Namespace: cm.Namespace,
				Keys:      keys,
				Age:       calculateAge(cm.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listSecrets(ctx context.Context, namespaces []string) ([]dto.SecretInfo, error) {
	var secrets corev1.SecretList
	var result []dto.SecretInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &secrets, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, secret := range secrets.Items {
			var keys []string
			for k := range secret.Data {
				keys = append(keys, k)
			}

			info := dto.SecretInfo{
				Name:      secret.Name,
				Namespace: secret.Namespace,
				Labels:    secret.Labels,
				Keys:      keys,
				Type:      string(secret.Type),
				Age:       calculateAge(secret.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listHPA(ctx context.Context, namespaces []string) ([]dto.HPAInfo, error) {
	var hpas autoscalingv2.HorizontalPodAutoscalerList
	var result []dto.HPAInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &hpas, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, hpa := range hpas.Items {
			var metrics []string
			for _, metric := range hpa.Spec.Metrics {
				metrics = append(metrics, string(metric.Type))
			}

			status := "Unknown"
			if len(hpa.Status.Conditions) > 0 {
				status = string(hpa.Status.Conditions[0].Type)
			}

			var minPods int32
			if hpa.Spec.MinReplicas != nil {
				minPods = *hpa.Spec.MinReplicas
			}

			info := dto.HPAInfo{
				Name:      hpa.Name,
				Namespace: hpa.Namespace,
				Metrics:   metrics,
				MinPods:   minPods,
				MaxPods:   hpa.Spec.MaxReplicas,
				Replicas:  hpa.Status.CurrentReplicas,
				Age:       calculateAge(hpa.CreationTimestamp),
				Status:    status,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listCronJobs(ctx context.Context, namespaces []string) ([]dto.CronJobInfo, error) {
	var cronjobs batchv1.CronJobList
	var result []dto.CronJobInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &cronjobs, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, cronjob := range cronjobs.Items {
			lastSchedule := ""
			if cronjob.Status.LastScheduleTime != nil {
				lastSchedule = calculateAge(*cronjob.Status.LastScheduleTime)
			}

			info := dto.CronJobInfo{
				Name:         cronjob.Name,
				Namespace:    cronjob.Namespace,
				Schedule:     cronjob.Spec.Schedule,
				Suspend:      cronjob.Spec.Suspend != nil && *cronjob.Spec.Suspend,
				Active:       len(cronjob.Status.Active),
				LastSchedule: lastSchedule,
				Age:          calculateAge(cronjob.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listResourceQuotas(ctx context.Context, namespaces []string) ([]dto.ResourceQuotaInfo, error) {
	var quotas corev1.ResourceQuotaList
	var result []dto.ResourceQuotaInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &quotas, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, quota := range quotas.Items {
			info := dto.ResourceQuotaInfo{
				Name:      quota.Name,
				Namespace: quota.Namespace,
				Age:       calculateAge(quota.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listLimitRanges(ctx context.Context, namespaces []string) ([]dto.LimitRangeInfo, error) {
	var limits corev1.LimitRangeList
	var result []dto.LimitRangeInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &limits, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, limit := range limits.Items {
			info := dto.LimitRangeInfo{
				Name:      limit.Name,
				Namespace: limit.Namespace,
				Age:       calculateAge(limit.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listVPA(ctx context.Context, namespaces []string) ([]dto.VPAInfo, error) {
	var vpas autoscalingv1.VerticalPodAutoscalerList
	var result []dto.VPAInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &vpas, client.InNamespace(ns)); err != nil {
			fmt.Println(err)
			return nil, err
		}

		for _, vpa := range vpas.Items {
			// Handle UpdateMode safely
			mode := "Unknown"
			if vpa.Spec.UpdatePolicy != nil && vpa.Spec.UpdatePolicy.UpdateMode != nil {
				mode = string(*vpa.Spec.UpdatePolicy.UpdateMode)
			}

			// Initialize default values
			cpu := "No recommendation"
			memory := "No recommendation"

			// Handle Recommendation status safely
			if vpa.Status.Recommendation != nil && len(vpa.Status.Recommendation.ContainerRecommendations) > 0 {
				rec := vpa.Status.Recommendation.ContainerRecommendations[0]
				cpu = rec.Target.Cpu().String()
				memory = rec.Target.Memory().String()
			}

			info := dto.VPAInfo{
				Name:      vpa.Name,
				Namespace: vpa.Namespace,
				Age:       calculateAge(vpa.CreationTimestamp),
				Mode:      mode,
				CPU:       cpu,
				Memory:    memory,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listPDBs(ctx context.Context, namespaces []string) ([]dto.PDBInfo, error) {
	var pdbs policyv1.PodDisruptionBudgetList
	var result []dto.PDBInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &pdbs, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, pdb := range pdbs.Items {
			info := dto.PDBInfo{
				Name:           pdb.Name,
				Namespace:      pdb.Namespace,
				Age:            calculateAge(pdb.CreationTimestamp),
				MinAvailable:   pdb.Spec.MinAvailable.String(),
				MaxUnavailable: pdb.Spec.MaxUnavailable.String(),
				CurrentHealthy: pdb.Status.CurrentHealthy,
				DesiredHealthy: pdb.Status.DesiredHealthy,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listPriorityClasses(ctx context.Context) ([]dto.PriorityClassInfo, error) {
	var pcs schedulingv1.PriorityClassList
	var result []dto.PriorityClassInfo

	if err := c.client.List(ctx, &pcs); err != nil {
		return nil, err
	}

	for _, pc := range pcs.Items {
		info := dto.PriorityClassInfo{
			Name:          pc.Name,
			Age:           calculateAge(pc.CreationTimestamp),
			Value:         pc.Value,
			GlobalDefault: pc.GlobalDefault,
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listRuntimeClasses(ctx context.Context) ([]dto.RuntimeClassInfo, error) {
	var rcs nodev1.RuntimeClassList
	var result []dto.RuntimeClassInfo

	if err := c.client.List(ctx, &rcs); err != nil {
		return nil, err
	}

	for _, rc := range rcs.Items {
		info := dto.RuntimeClassInfo{
			Name:    rc.Name,
			Handler: rc.Handler,
			Age:     calculateAge(rc.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listLeases(ctx context.Context, namespaces []string) ([]dto.LeaseInfo, error) {
	var leases coordinationv1.LeaseList
	var result []dto.LeaseInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &leases, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, lease := range leases.Items {
			info := dto.LeaseInfo{
				Name:      lease.Name,
				Namespace: lease.Namespace,
				Age:       calculateAge(lease.CreationTimestamp),
				Holder:    *lease.Spec.HolderIdentity,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listServices(ctx context.Context, namespaces []string) ([]dto.ServiceInfo, error) {
	var services corev1.ServiceList
	var result []dto.ServiceInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &services, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, svc := range services.Items {
			var ports []string
			for _, port := range svc.Spec.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			}

			var externalIPs []string
			for _, ing := range svc.Status.LoadBalancer.Ingress {
				if ing.IP != "" {
					externalIPs = append(externalIPs, ing.IP)
				}
				if ing.Hostname != "" {
					externalIPs = append(externalIPs, ing.Hostname)
				}
			}

			info := dto.ServiceInfo{
				Name:       svc.Name,
				Namespace:  svc.Namespace,
				Age:        calculateAge(svc.CreationTimestamp),
				Type:       string(svc.Spec.Type),
				ClusterIP:  svc.Spec.ClusterIP,
				Ports:      strings.Join(ports, ","),
				ExternalIP: externalIPs,
				Selector:   svc.Spec.Selector,
				Status:     getServiceStatus(&svc),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listEndpoints(ctx context.Context, namespaces []string) ([]dto.EndpointInfo, error) {
	var endpoints corev1.EndpointsList
	var result []dto.EndpointInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &endpoints, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, ep := range endpoints.Items {
			var endpointAddresses []string
			for _, subset := range ep.Subsets {
				for _, addr := range subset.Addresses {
					endpointAddresses = append(endpointAddresses, addr.IP)
				}
			}

			info := dto.EndpointInfo{
				Name:      ep.Name,
				Namespace: ep.Namespace,
				Age:       calculateAge(ep.CreationTimestamp),
				Endpoints: endpointAddresses,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listIngresses(ctx context.Context, namespaces []string) ([]dto.IngressInfo, error) {
	var ingresses networkingv1.IngressList
	var result []dto.IngressInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &ingresses, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, ing := range ingresses.Items {
			var rules []string
			for _, rule := range ing.Spec.Rules {
				rules = append(rules, rule.Host)
			}

			var loadBalancers []string
			for _, lb := range ing.Status.LoadBalancer.Ingress {
				if lb.IP != "" {
					loadBalancers = append(loadBalancers, lb.IP)
				}
				if lb.Hostname != "" {
					loadBalancers = append(loadBalancers, lb.Hostname)
				}
			}

			info := dto.IngressInfo{
				Name:          ing.Name,
				Namespace:     ing.Namespace,
				Age:           calculateAge(ing.CreationTimestamp),
				LoadBalancers: loadBalancers,
				Rules:         rules,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listIngressClasses(ctx context.Context) ([]dto.IngressClassInfo, error) {
	var ingressClasses networkingv1.IngressClassList
	var result []dto.IngressClassInfo

	if err := c.client.List(ctx, &ingressClasses); err != nil {
		return nil, err
	}

	for _, ic := range ingressClasses.Items {
		info := dto.IngressClassInfo{
			Name:       ic.Name,
			Namespace:  ic.Namespace, // IngressClass is cluster-scoped, so this will be empty
			Controller: ic.Spec.Controller,
			APIGroup:   "",
			Kind:       "",
			Scope:      "",
		}

		// Safely handle Parameters which might be nil
		if ic.Spec.Parameters != nil {
			if ic.Spec.Parameters.APIGroup != nil {
				info.APIGroup = *ic.Spec.Parameters.APIGroup
			}
			info.Kind = ic.Spec.Parameters.Kind
			if ic.Spec.Parameters.Scope != nil {
				info.Scope = *ic.Spec.Parameters.Scope
			}
		}

		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listNetworkPolicies(ctx context.Context, namespaces []string) ([]dto.NetworkPolicyInfo, error) {
	var netpols networkingv1.NetworkPolicyList
	var result []dto.NetworkPolicyInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &netpols, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, np := range netpols.Items {
			var policyTypes []string
			for _, pt := range np.Spec.PolicyTypes {
				policyTypes = append(policyTypes, string(pt))
			}

			info := dto.NetworkPolicyInfo{
				Name:        np.Name,
				Namespace:   np.Namespace,
				Age:         calculateAge(np.CreationTimestamp),
				PolicyTypes: policyTypes,
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listPVCs(ctx context.Context, namespaces []string) ([]dto.PVCInfo, error) {
	var pvcs corev1.PersistentVolumeClaimList
	var result []dto.PVCInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &pvcs, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, pvc := range pvcs.Items {
			storageClass := ""
			if pvc.Spec.StorageClassName != nil {
				storageClass = *pvc.Spec.StorageClassName
			}

			info := dto.PVCInfo{
				Name:         pvc.Name,
				Namespace:    pvc.Namespace,
				StorageClass: storageClass,
				Size:         pvc.Spec.Resources.Requests.Storage().String(),
				Pods:         c.getPVCPods(&pvc),
				Age:          calculateAge(pvc.CreationTimestamp),
				Status:       string(pvc.Status.Phase),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listPVs(ctx context.Context) ([]dto.PVInfo, error) {
	var pvs corev1.PersistentVolumeList
	var result []dto.PVInfo

	if err := c.client.List(ctx, &pvs); err != nil {
		return nil, err
	}

	for _, pv := range pvs.Items {
		var claim string
		if pv.Spec.ClaimRef != nil {
			claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		info := dto.PVInfo{
			Name:         pv.Name,
			StorageClass: pv.Spec.StorageClassName,
			Capacity:     pv.Spec.Capacity.Storage().String(),
			Claim:        claim,
			Age:          calculateAge(pv.CreationTimestamp),
			Status:       string(pv.Status.Phase),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listStorageClasses(ctx context.Context) ([]dto.StorageClassInfo, error) {
	var scs storagev1.StorageClassList
	var result []dto.StorageClassInfo

	if err := c.client.List(ctx, &scs); err != nil {
		return nil, err
	}

	for _, sc := range scs.Items {
		info := dto.StorageClassInfo{
			Name:          sc.Name,
			Provisioner:   sc.Provisioner,
			ReclaimPolicy: string(*sc.ReclaimPolicy),
			Default:       sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true",
			Age:           calculateAge(sc.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listEvents(ctx context.Context, namespaces []string) ([]dto.EventInfo, error) {
	var events corev1.EventList
	var result []dto.EventInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &events, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, event := range events.Items {
			info := dto.EventInfo{
				Type:           event.Type,
				Message:        event.Message,
				Namespace:      event.Namespace,
				InvolvedObject: fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
				Source:         event.Source.Component,
				Count:          event.Count,
				Age:            calculateAge(event.CreationTimestamp),
				LastSeen:       calculateAge(event.LastTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listNamespaces(ctx context.Context) ([]dto.NamespaceInfo, error) {
	var namespaces corev1.NamespaceList
	var result []dto.NamespaceInfo

	if err := c.client.List(ctx, &namespaces); err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		info := dto.NamespaceInfo{
			Name:   ns.Name,
			Labels: ns.Labels,
			Age:    calculateAge(ns.CreationTimestamp),
			Status: string(ns.Status.Phase),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listServiceAccounts(ctx context.Context, namespaces []string) ([]dto.ServiceAccountInfo, error) {
	var serviceaccounts corev1.ServiceAccountList
	var result []dto.ServiceAccountInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &serviceaccounts, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, sa := range serviceaccounts.Items {
			info := dto.ServiceAccountInfo{
				Name:      sa.Name,
				Namespace: sa.Namespace,
				Age:       calculateAge(sa.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listClusterRoles(ctx context.Context) ([]dto.ClusterRoleInfo, error) {
	var clusterRoles rbacv1.ClusterRoleList
	var result []dto.ClusterRoleInfo

	if err := c.client.List(ctx, &clusterRoles); err != nil {
		return nil, err
	}

	for _, cr := range clusterRoles.Items {
		info := dto.ClusterRoleInfo{
			Name: cr.Name,
			Age:  calculateAge(cr.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listRoles(ctx context.Context, namespaces []string) ([]dto.RoleInfo, error) {
	var roles rbacv1.RoleList
	var result []dto.RoleInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &roles, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, role := range roles.Items {
			info := dto.RoleInfo{
				Name:      role.Name,
				Namespace: role.Namespace,
				Age:       calculateAge(role.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) listClusterRoleBindings(ctx context.Context) ([]dto.ClusterRoleBindingInfo, error) {
	var clusterRoleBindings rbacv1.ClusterRoleBindingList
	var result []dto.ClusterRoleBindingInfo

	if err := c.client.List(ctx, &clusterRoleBindings); err != nil {
		return nil, err
	}

	for _, crb := range clusterRoleBindings.Items {
		var bindings []string
		for _, subject := range crb.Subjects {
			binding := fmt.Sprintf("%s/%s/%s", subject.Kind, subject.Namespace, subject.Name)
			bindings = append(bindings, binding)
		}

		info := dto.ClusterRoleBindingInfo{
			Name:     crb.Name,
			Bindings: bindings,
			Age:      calculateAge(crb.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listRoleBindings(ctx context.Context, namespaces []string) ([]dto.RoleBindingInfo, error) {
	var roleBindings rbacv1.RoleBindingList
	var result []dto.RoleBindingInfo

	for _, ns := range namespaces {
		if err := c.client.List(ctx, &roleBindings, client.InNamespace(ns)); err != nil {
			return nil, err
		}

		for _, rb := range roleBindings.Items {
			var bindings []string
			for _, subject := range rb.Subjects {
				binding := fmt.Sprintf("%s/%s/%s", subject.Kind, subject.Namespace, subject.Name)
				bindings = append(bindings, binding)
			}

			info := dto.RoleBindingInfo{
				Name:      rb.Name,
				Namespace: rb.Namespace,
				Bindings:  bindings,
				Age:       calculateAge(rb.CreationTimestamp),
			}
			result = append(result, info)
		}
	}
	return result, nil
}

func (c *ListController) getPVCPods(pvc *corev1.PersistentVolumeClaim) string {
	var pods corev1.PodList
	if err := c.client.List(context.Background(), &pods, client.InNamespace(pvc.Namespace)); err != nil {
		return "0"
	}

	count := 0
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {
				count++
			}
		}
	}
	return fmt.Sprintf("%d", count)
}

func (c *ListController) listMutatingWebhooks(ctx context.Context) ([]dto.MutatingWebhookInfo, error) {
	var webhooks admissionv1.MutatingWebhookConfigurationList
	var result []dto.MutatingWebhookInfo

	if err := c.client.List(ctx, &webhooks); err != nil {
		return nil, err
	}

	for _, webhook := range webhooks.Items {
		info := dto.MutatingWebhookInfo{
			Name:         webhook.Name,
			WebhookCount: len(webhook.Webhooks),
			Age:          calculateAge(webhook.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *ListController) listValidatingWebhooks(ctx context.Context) ([]dto.ValidatingWebhookInfo, error) {
	var webhooks admissionv1.ValidatingWebhookConfigurationList
	var result []dto.ValidatingWebhookInfo

	if err := c.client.List(ctx, &webhooks); err != nil {
		return nil, err
	}

	for _, webhook := range webhooks.Items {
		info := dto.ValidatingWebhookInfo{
			Name:         webhook.Name,
			WebhookCount: len(webhook.Webhooks),
			Age:          calculateAge(webhook.CreationTimestamp),
		}
		result = append(result, info)
	}
	return result, nil
}

// Helper function for service status
func getServiceStatus(svc *corev1.Service) string {
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return "Active"
		}
		return "Pending"
	}
	return "Active"
}

// func ptrToString(s *string) string {
// 	if s == nil {
// 		return ""
// 	}
// 	return *s
// }
