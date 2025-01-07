package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	controllers "agentkube.com/agent-kube-operator/internal/controllers"
	kubectl "agentkube.com/agent-kube-operator/internal/controllers/kubectl"
	metrics "agentkube.com/agent-kube-operator/internal/controllers/metrics"
	monitor "agentkube.com/agent-kube-operator/internal/controllers/monitor"
	pod "agentkube.com/agent-kube-operator/internal/controllers/pod"
	resources "agentkube.com/agent-kube-operator/internal/controllers/resources"
	"agentkube.com/agent-kube-operator/internal/dto"
	utils "agentkube.com/agent-kube-operator/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// var log = ctrl.Log.WithName("handlers")

type Handler struct {
	kubectlController *kubectl.Controller
	k8sClient         client.Client
	scheme            *runtime.Scheme
	restConfig        *rest.Config
}

func NewHandler(client client.Client, scheme *runtime.Scheme, config *rest.Config) (*Handler, error) {
	kubectlController, err := kubectl.NewController(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubectl controller: %v", err)
	}

	return &Handler{
		kubectlController: kubectlController,
		k8sClient:         client,
		scheme:            scheme,
		restConfig:        config,
	}, nil
}

// Health and readiness handlers
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func (h *Handler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Cluster handlers
func (h *Handler) GetClusterInfo(c *gin.Context) {
	// TODO: Implement cluster info retrieval
	clusterName := utils.GetEnviron("CLUSTER_NAME")
	externalEndpoint := utils.GetEnviron("EXTERNAL_ENDPOTIN")

	c.JSON(http.StatusOK, gin.H{
		"clusterName":      clusterName,
		"externalEndpoint": externalEndpoint,
		"version":          "1.0.0",
	})
}

func (h *Handler) GetClusterMetrics(c *gin.Context) {
	metricsController := metrics.NewController(h.k8sClient, h.scheme)

	metrics, err := metricsController.GetClusterMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) ExecuteKubectl(c *gin.Context) {
	var request struct {
		Command []string `json:"command"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	command := request.Command
	result, err := h.kubectlController.ExecuteCommand(command...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetNamespaceMetrics(c *gin.Context) {
	namespace := c.Query("namespace")

	metricsController := metrics.NewController(h.k8sClient, h.scheme)
	metrics, err := metricsController.GetNamespaceMetrics(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) ListResources(c *gin.Context) {
	var req controllers.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	listController := controllers.NewListController(h.k8sClient, h.scheme)
	resources, err := listController.ListResources(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources)
}

func (h *Handler) GetNodes(c *gin.Context) {
	nodeController := controllers.NewNodeController(h.k8sClient, h.scheme)

	nodes, err := nodeController.GetNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *Handler) ListAPIResources(c *gin.Context) {
	resourcesController, err := resources.NewController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create controller: %v", err),
		})
		return
	}

	resources, err := resourcesController.ListAPIResources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources)
}

func (h *Handler) GetK8sResource(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	version := c.Param("version")
	resourceType := c.Param("resource_type")
	resourceName := c.Param("resource_name")

	if version == "" || resourceType == "" || resourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "version, resource_type, and resource_name are required",
		})
		return
	}

	resourcesController, err := resources.NewController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create controller: %v", err),
		})
		return
	}

	result, err := resourcesController.GetResource(
		c.Request.Context(),
		namespace,
		group,
		version,
		resourceType,
		resourceName,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if c.Query("output") == "yaml" {
		yamlData, err := yaml.Marshal(result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to convert to yaml: %v", err),
			})
			return
		}
		c.Header("Content-Type", "application/yaml")
		c.String(http.StatusOK, string(yamlData))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ApplyK8sResource(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	version := c.Param("version")
	resourceType := c.Param("resource_type")
	resourceName := c.Param("resource_name")

	var content map[string]interface{}
	if err := c.ShouldBindJSON(&content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid yaml content: %v", err),
		})
		return
	}

	resourcesController, err := resources.NewController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create controller: %v", err),
		})
		return
	}

	err = resourcesController.ApplyResource(
		c.Request.Context(),
		namespace,
		group,
		version,
		resourceType,
		resourceName,
		content,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Resource updated successfully",
	})
}

func (h *Handler) DeleteK8sResource(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	version := c.Param("version")
	resourceType := c.Param("resource_type")
	resourceName := c.Param("resource_name")

	if version == "" || resourceType == "" || resourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "version, resource_type, and resource_name are required",
		})
		return
	}

	resourcesController, err := resources.NewController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create controller: %v", err),
		})
		return
	}

	err = resourcesController.DeleteResource(
		c.Request.Context(),
		namespace,
		group,
		version,
		resourceType,
		resourceName,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Resource %s/%s deleted successfully", resourceType, resourceName),
	})
}

func (h *Handler) GetPodMetrics(c *gin.Context) {
	metricsController, err := pod.NewMetricsController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create metrics controller: %v", err),
		})
		return
	}

	metrics, err := metricsController.GetPodMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) GetHistoricalPodMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "1d")

	metricsController, err := monitor.NewMetricsController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create metrics controller: %v", err),
		})
		return
	}

	metrics, err := metricsController.GetHistoricalMetrics(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) GetComprehensivePodMetrics(c *gin.Context) {
	// Get query parameters
	timeRange := c.DefaultQuery("range", "1h")
	podName := c.Query("pod")
	namespace := c.DefaultQuery("namespace", "default")

	startTime, endTime, step, err := getTimeRangeParams(timeRange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create metrics controller
	metricsController, err := monitor.NewMetricsController(h.k8sClient, h.scheme, h.restConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create metrics controller: %v", err),
		})
		return
	}

	// Get list of pods with optional filtering
	var pods corev1.PodList
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}

	if err := h.k8sClient.List(c.Request.Context(), &pods, opts...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to list pods: %v", err),
		})
		return
	}

	// Filter pods by name if specified
	var filteredPods []corev1.Pod
	if podName != "" {
		for _, pod := range pods.Items {
			if pod.Name == podName {
				filteredPods = append(filteredPods, pod)
				break
			}
		}
		if len(filteredPods) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("pod %s not found in namespace %s", podName, namespace),
			})
			return
		}
		pods.Items = filteredPods
	}

	var allMetrics []dto.ComprehensivePodMetrics

	for _, pod := range pods.Items {
		queries := map[string]string{
			"cpu":    fmt.Sprintf(`rate(container_cpu_usage_seconds_total{pod="%s",namespace="%s"}[5m])`, pod.Name, pod.Namespace),
			"memory": fmt.Sprintf(`container_memory_working_set_bytes{pod="%s",namespace="%s"}`, pod.Name, pod.Namespace),
			"rx":     fmt.Sprintf(`rate(container_network_receive_bytes_total{pod="%s",namespace="%s"}[5m])`, pod.Name, pod.Namespace),
			"tx":     fmt.Sprintf(`rate(container_network_transmit_bytes_total{pod="%s",namespace="%s"}[5m])`, pod.Name, pod.Namespace),
			"rxErr":  fmt.Sprintf(`rate(container_network_receive_errors_total{pod="%s",namespace="%s"}[5m])`, pod.Name, pod.Namespace),
			"txErr":  fmt.Sprintf(`rate(container_network_transmit_errors_total{pod="%s",namespace="%s"}[5m])`, pod.Name, pod.Namespace),
		}

		metrics := dto.ComprehensivePodMetrics{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
		}

		if pod.Status.StartTime != nil {
			metrics.StartTime = pod.Status.StartTime.Format(time.RFC3339)
		}

		// Get all metrics for the time range
		responses := make(map[string]*monitor.PrometheusResponse)
		for metricName, query := range queries {
			response, err := metricsController.QueryPrometheus(query, startTime, endTime, step)
			if err != nil {
				log.Printf("Error querying metric %s for pod %s: %v", metricName, pod.Name, err)
				continue
			}
			responses[metricName] = response
		}

		// Process all time points
		if cpu, ok := responses["cpu"]; ok && len(cpu.Data.Result) > 0 {
			timePoints := cpu.Data.Result[0].Values
			for i, timePoint := range timePoints {
				timestamp := time.Unix(int64(timePoint[0].(float64)), 0)

				point := dto.PodMetricsPoint{
					Timestamp: timestamp,
				}

				// Get resource requests and limits (constant values)
				for _, container := range pod.Spec.Containers {
					if requests := container.Resources.Requests; requests != nil {
						if cpu := requests.Cpu(); cpu != nil {
							point.CPURequests += cpu.AsApproximateFloat64()
						}
						if memory := requests.Memory(); memory != nil {
							point.MemoryRequests += float64(memory.Value()) / (1024 * 1024) // Convert to MiB
						}
					}
					if limits := container.Resources.Limits; limits != nil {
						if cpu := limits.Cpu(); cpu != nil {
							point.CPULimits += cpu.AsApproximateFloat64()
						}
						if memory := limits.Memory(); memory != nil {
							point.MemoryLimits += float64(memory.Value()) / (1024 * 1024) // Convert to MiB
						}
					}
				}

				// Get container restart count
				for _, containerStatus := range pod.Status.ContainerStatuses {
					point.RestartCount += containerStatus.RestartCount
				}

				// Add metrics from Prometheus for this timestamp
				for metricName, response := range responses {
					if len(response.Data.Result) > 0 && i < len(response.Data.Result[0].Values) {
						value, err := parseFloat64(response.Data.Result[0].Values[i][1])
						if err != nil {
							log.Printf("Error parsing value for metric %s: %v", metricName, err)
							continue
						}

						switch metricName {
						case "cpu":
							point.CPUUsage = value
						case "memory":
							point.MemoryUsage = value / (1024 * 1024) // Convert to MiB
						case "rx":
							point.NetworkRxBytes = value
						case "tx":
							point.NetworkTxBytes = value
						case "rxErr":
							point.NetworkRxErrors = value
						case "txErr":
							point.NetworkTxErrors = value
						}
					}
				}

				metrics.TimePoints = append(metrics.TimePoints, point)
			}
		}

		allMetrics = append(allMetrics, metrics)
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":   allMetrics,
		"timeRange": timeRange,
		"startTime": startTime.Format(time.RFC3339),
		"endTime":   endTime.Format(time.RFC3339),
		"step":      step.String(),
		"query": gin.H{
			"namespace": namespace,
			"pod":       podName,
		},
	})
}

// Helpers
func getTimeRangeParams(timeRange string) (time.Time, time.Time, time.Duration, error) {
	endTime := time.Now()
	var startTime time.Time
	var step time.Duration

	switch timeRange {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
		step = 1 * time.Minute
	case "1d":
		startTime = endTime.Add(-24 * time.Hour)
		step = 15 * time.Minute
	case "3d":
		startTime = endTime.Add(-72 * time.Hour)
		step = 30 * time.Minute
	default:
		return time.Time{}, time.Time{}, 0, fmt.Errorf("invalid time range: %s. Supported values are: 1h, 1d, 3d", timeRange)
	}

	return startTime, endTime, step, nil
}

func parseFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported type for float64 conversion: %T", value)
	}
}
