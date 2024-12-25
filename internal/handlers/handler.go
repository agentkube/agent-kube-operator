package handlers

import (
	"net/http"

	controllers "agentkube.com/agent-kube-operator/internal/controllers"
	kubectl "agentkube.com/agent-kube-operator/internal/controllers/kubectl"
	metrics "agentkube.com/agent-kube-operator/internal/controllers/metrics"
	"agentkube.com/agent-kube-operator/utils"
	"github.com/gin-gonic/gin"
	runtime "k8s.io/apimachinery/pkg/runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// var log = ctrl.Log.WithName("handlers")

type Handler struct {
	kubectlController *kubectl.Controller
	k8sClient         client.Client
	scheme            *runtime.Scheme
}

func NewHandler(client client.Client, scheme *runtime.Scheme) *Handler {
	return &Handler{
		k8sClient: client,
		scheme:    scheme,
	}
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

func (h *Handler) GetNamespaceResources(c *gin.Context) {
	var req metrics.NamespacedResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	// Validate the resource type
	validResources := map[string]bool{
		"pods":         true,
		"deployments":  true,
		"daemonsets":   true,
		"statefulsets": true,
		"services":     true,
	}

	if !validResources[req.Resource] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid resource type: " + req.Resource,
		})
		return
	}

	metricsController := metrics.NewController(h.k8sClient, h.scheme)
	resources, err := metricsController.GetNamespacedResources(c.Request.Context(), req.Namespace, req.Resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace": req.Namespace,
		"resource":  req.Resource,
		"items":     resources,
	})
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
