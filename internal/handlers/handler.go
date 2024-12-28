package handlers

import (
	"fmt"
	"net/http"

	controllers "agentkube.com/agent-kube-operator/internal/controllers"
	kubectl "agentkube.com/agent-kube-operator/internal/controllers/kubectl"
	metrics "agentkube.com/agent-kube-operator/internal/controllers/metrics"
	raw "agentkube.com/agent-kube-operator/internal/controllers/raw"
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

func NewHandler(client client.Client, scheme *runtime.Scheme, config *rest.Config) *Handler {
	return &Handler{
		k8sClient:  client,
		scheme:     scheme,
		restConfig: config,
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

func (h *Handler) GetRawResource(c *gin.Context) {
	// Get the path from the URL
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "path parameter is required",
		})
		return
	}

	// Ensure path starts with /
	if path[0] != '/' {
		path = "/" + path
	}

	fmt.Println(path)

	rawController := raw.NewController(h.k8sClient, h.scheme, h.restConfig)
	result, err := rawController.GetRawResource(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if yaml format is requested
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

// TODO will be deprecated
func (h *Handler) GetNamespacedResource(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	version := c.Param("version")
	resourceType := c.Param("resource_type")
	resourceName := c.Param("resource_name")

	// Validate required parameters
	if namespace == "" || version == "" || resourceType == "" || resourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace, version, resource_type, and resource_name are required",
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

	result, err := resourcesController.GetNamespacedResource(
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

	// Check if yaml format is requested
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
