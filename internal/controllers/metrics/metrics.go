package metrics

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Controller handles metrics collection from the Kubernetes cluster
type Controller struct {
	client client.Client
	scheme *runtime.Scheme
}

// MetricsResponse represents the collected cluster metrics
type MetricsResponse struct {
	Workloads struct {
		Namespaces    int64 `json:"namespaces"`
		Deployments   int64 `json:"deployments"`
		ReplicaSets   int64 `json:"replicasets"`
		StatefulSets  int64 `json:"statefulsets"`
		Pods          int64 `json:"pods"`
		Nodes         int64 `json:"nodes"`
		Jobs          int64 `json:"jobs"`
		RunningJobs   int64 `json:"running_jobs"`
		CompletedJobs int64 `json:"completed_jobs"`
		CronJobs      int64 `json:"cronjobs"`
		DaemonSets    int64 `json:"daemonsets"`
	} `json:"workloads"`
	Network struct {
		Services  int64 `json:"services"`
		Endpoints int64 `json:"endpoints"`
		Ingresses int64 `json:"ingresses"`
	} `json:"network"`
	Storage struct {
		PersistentVolumes      int64 `json:"persistent_volumes"`
		PersistentVolumeClaims int64 `json:"persistent_volume_claims"`
		StorageClasses         int64 `json:"storage_classes"`
	} `json:"storage"`
}

// NewController creates a new metrics controller instance
func NewController(client client.Client, scheme *runtime.Scheme) *Controller {
	return &Controller{
		client: client,
		scheme: scheme,
	}
}

// GetClusterMetrics collects all metrics from the cluster
func (c *Controller) GetClusterMetrics(ctx context.Context) (*MetricsResponse, error) {
	metrics := &MetricsResponse{}

	// Collect workload metrics
	if err := c.collectWorkloadMetrics(ctx, metrics); err != nil {
		return nil, err
	}

	// Collect network metrics
	if err := c.collectNetworkMetrics(ctx, metrics); err != nil {
		return nil, err
	}

	// Collect storage metrics
	if err := c.collectStorageMetrics(ctx, metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *Controller) collectWorkloadMetrics(ctx context.Context, metrics *MetricsResponse) error {
	// Get namespaces
	var namespaces corev1.NamespaceList
	if err := c.client.List(ctx, &namespaces); err != nil {
		return err
	}
	metrics.Workloads.Namespaces = int64(len(namespaces.Items))

	// Get deployments
	var deployments appsv1.DeploymentList
	if err := c.client.List(ctx, &deployments); err != nil {
		return err
	}
	metrics.Workloads.Deployments = int64(len(deployments.Items))

	// Get replicasets
	var replicaSets appsv1.ReplicaSetList
	if err := c.client.List(ctx, &replicaSets); err != nil {
		return err
	}
	metrics.Workloads.ReplicaSets = int64(len(replicaSets.Items))

	// Get statefulsets
	var statefulSets appsv1.StatefulSetList
	if err := c.client.List(ctx, &statefulSets); err != nil {
		return err
	}
	metrics.Workloads.StatefulSets = int64(len(statefulSets.Items))

	// Get pods
	var pods corev1.PodList
	if err := c.client.List(ctx, &pods); err != nil {
		return err
	}
	metrics.Workloads.Pods = int64(len(pods.Items))

	// Get nodes
	var nodes corev1.NodeList
	if err := c.client.List(ctx, &nodes); err != nil {
		return err
	}
	metrics.Workloads.Nodes = int64(len(nodes.Items))

	// Get jobs and count running/completed
	var jobs batchv1.JobList
	if err := c.client.List(ctx, &jobs); err != nil {
		return err
	}
	metrics.Workloads.Jobs = int64(len(jobs.Items))
	for _, job := range jobs.Items {
		if job.Status.Active > 0 {
			metrics.Workloads.RunningJobs++
		}
		if job.Status.Succeeded > 0 {
			metrics.Workloads.CompletedJobs++
		}
	}

	// Get cronjobs
	var cronJobs batchv1.CronJobList
	if err := c.client.List(ctx, &cronJobs); err != nil {
		return err
	}
	metrics.Workloads.CronJobs = int64(len(cronJobs.Items))

	// Get daemonsets
	var daemonSets appsv1.DaemonSetList
	if err := c.client.List(ctx, &daemonSets); err != nil {
		return err
	}
	metrics.Workloads.DaemonSets = int64(len(daemonSets.Items))

	return nil
}

func (c *Controller) collectNetworkMetrics(ctx context.Context, metrics *MetricsResponse) error {
	// Get services
	var services corev1.ServiceList
	if err := c.client.List(ctx, &services); err != nil {
		return err
	}
	metrics.Network.Services = int64(len(services.Items))

	// Get endpoints
	var endpoints corev1.EndpointsList
	if err := c.client.List(ctx, &endpoints); err != nil {
		return err
	}
	metrics.Network.Endpoints = int64(len(endpoints.Items))

	// Get ingresses
	var ingresses networkingv1.IngressList
	if err := c.client.List(ctx, &ingresses); err != nil {
		return err
	}
	metrics.Network.Ingresses = int64(len(ingresses.Items))

	return nil
}

func (c *Controller) collectStorageMetrics(ctx context.Context, metrics *MetricsResponse) error {
	// Get persistent volumes
	var pvs corev1.PersistentVolumeList
	if err := c.client.List(ctx, &pvs); err != nil {
		return err
	}
	metrics.Storage.PersistentVolumes = int64(len(pvs.Items))

	// Get persistent volume claims
	var pvcs corev1.PersistentVolumeClaimList
	if err := c.client.List(ctx, &pvcs); err != nil {
		return err
	}
	metrics.Storage.PersistentVolumeClaims = int64(len(pvcs.Items))

	// Get storage classes
	var storageClasses storagev1.StorageClassList
	if err := c.client.List(ctx, &storageClasses); err != nil {
		return err
	}
	metrics.Storage.StorageClasses = int64(len(storageClasses.Items))

	return nil
}
