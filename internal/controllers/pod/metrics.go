package pod

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MetricsController struct {
	client        client.Client
	metricsClient *metrics.Clientset
	scheme        *runtime.Scheme
}

type PodMetrics struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	CPUUsage          float64           `json:"cpuUsage"`         // in cores
	MemoryUsage       float64           `json:"memoryUsage"`      // in MiB
	CPURequests       float64           `json:"cpuRequests"`      // in cores
	MemoryRequests    float64           `json:"memoryRequests"`   // in MiB
	CPULimits         float64           `json:"cpuLimits"`        // in cores
	MemoryLimits      float64           `json:"memoryLimits"`     // in MiB
	NetworkRxBytes    float64           `json:"networkRxBytes"`   // bytes received
	NetworkTxBytes    float64           `json:"networkTxBytes"`   // bytes transmitted
	NetworkRxPackets  float64           `json:"networkRxPackets"` // packets received
	NetworkTxPackets  float64           `json:"networkTxPackets"` // packets transmitted
	NetworkRxErrors   float64           `json:"networkRxErrors"`  // receive errors
	NetworkTxErrors   float64           `json:"networkTxErrors"`  // transmit errors
	NetworkRxDropped  float64           `json:"networkRxDropped"` // receive packets dropped
	NetworkTxDropped  float64           `json:"networkTxDropped"` // transmit packets dropped
	RestartCount      int32             `json:"restartCount"`
	LastRestartTime   string            `json:"lastRestartTime,omitempty"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses"`
}

type ContainerStatus struct {
	Name           string `json:"name"`
	Ready          bool   `json:"ready"`
	RestartCount   int32  `json:"restartCount"`
	ContainerState string `json:"containerState"`
}

func NewMetricsController(client client.Client, scheme *runtime.Scheme, config *rest.Config) (*MetricsController, error) {
	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	return &MetricsController{
		client:        client,
		metricsClient: metricsClient,
		scheme:        scheme,
	}, nil
}

func (c *MetricsController) GetPodMetrics(ctx context.Context) ([]PodMetrics, error) {
	var pods corev1.PodList
	if err := c.client.List(ctx, &pods); err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	podMetricsList, err := c.metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %v", err)
	}

	podMetricsMap := make(map[string]v1beta1.PodMetrics)
	for _, pm := range podMetricsList.Items {
		key := fmt.Sprintf("%s/%s", pm.Namespace, pm.Name)
		podMetricsMap[key] = pm
	}

	var result []PodMetrics
	for _, pod := range pods.Items {
		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		metrics := PodMetrics{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}

		if pm, exists := podMetricsMap[key]; exists {
			for _, container := range pm.Containers {
				metrics.CPUUsage += container.Usage.Cpu().AsApproximateFloat64()
				memoryBytes := container.Usage.Memory().AsApproximateFloat64()
				metrics.MemoryUsage += memoryBytes / (1024 * 1024) // Convert to MiB
			}
		}

		// Get container resource requests and limits
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				metrics.CPURequests += container.Resources.Requests.Cpu().AsApproximateFloat64()
				memoryReqBytes := container.Resources.Requests.Memory().AsApproximateFloat64()
				metrics.MemoryRequests += memoryReqBytes / (1024 * 1024)
			}
			if container.Resources.Limits != nil {
				metrics.CPULimits += container.Resources.Limits.Cpu().AsApproximateFloat64()
				memoryLimBytes := container.Resources.Limits.Memory().AsApproximateFloat64()
				metrics.MemoryLimits += memoryLimBytes / (1024 * 1024)
			}
		}

		// Get container statuses
		for _, status := range pod.Status.ContainerStatuses {
			containerStatus := ContainerStatus{
				Name:         status.Name,
				Ready:        status.Ready,
				RestartCount: status.RestartCount,
			}

			if status.State.Running != nil {
				containerStatus.ContainerState = "Running"
			} else if status.State.Waiting != nil {
				containerStatus.ContainerState = "Waiting"
			} else if status.State.Terminated != nil {
				containerStatus.ContainerState = "Terminated"
			}

			metrics.ContainerStatuses = append(metrics.ContainerStatuses, containerStatus)
			metrics.RestartCount += status.RestartCount

			if status.LastTerminationState.Terminated != nil {
				metrics.LastRestartTime = status.LastTerminationState.Terminated.FinishedAt.Format(time.RFC3339)
			}
		}

		// Get network metrics from pod status
		if pod.Status.PodIP != "" {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.Ready {
							fmt.Println("")
							// Network stats would be collected here if available
							// Note: This requires cAdvisor metrics or a CNI that exposes these metrics
						}
					}
				}
			}
		}

		result = append(result, metrics)
	}

	return result, nil
}
