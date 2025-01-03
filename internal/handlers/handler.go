package handlers

import (
	"fmt"
	"net/http"

	controllers "agentkube.com/agent-kube-operator/internal/controllers"
	kubectl "agentkube.com/agent-kube-operator/internal/controllers/kubectl"
	metrics "agentkube.com/agent-kube-operator/internal/controllers/metrics"
	monitor "agentkube.com/agent-kube-operator/internal/controllers/monitor"
	pod "agentkube.com/agent-kube-operator/internal/controllers/pod"
	resources "agentkube.com/agent-kube-operator/internal/controllers/resources"
	utils "agentkube.com/agent-kube-operator/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
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
