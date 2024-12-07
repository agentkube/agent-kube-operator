package handlers

import (
	"net/http"

	kubectl "agentkube.com/agent-kube-operator/internal/controllers/kubectl"
	"agentkube.com/agent-kube-operator/utils"
	"github.com/gin-gonic/gin"
)

// var log = ctrl.Log.WithName("handlers")

type Handler struct {
	kubectlController *kubectl.Controller
}

func NewHandler() *Handler {
	return &Handler{}
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

// Agent handlers
func (h *Handler) ListAgents(c *gin.Context) {
	// TODO: Implement listing agents from the cluster
	c.JSON(http.StatusOK, gin.H{
		"agents": []string{},
	})
}

func (h *Handler) GetAgent(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement getting specific agent details
	c.JSON(http.StatusOK, gin.H{
		"agent": name,
	})
}

func (h *Handler) CreateAgent(c *gin.Context) {
	// TODO: Implement agent creation
	c.JSON(http.StatusCreated, gin.H{
		"message": "Agent created successfully",
	})
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement agent update
	c.JSON(http.StatusOK, gin.H{
		"message": "Agent " + name + " updated successfully",
	})
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	name := c.Param("name")
	// TODO: Implement agent deletion
	c.JSON(http.StatusOK, gin.H{
		"message": "Agent " + name + " deleted successfully",
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
	// TODO: Implement cluster metrics retrieval
	// Get total namespaces
	// Get total Deployments
	// Get total replicaset
	// Get total statefulset
	// Get total pods
	// Get total nodes
	// Get total Jobs (Running/Completed)
	// Get total CronJobs
	// Get total daemonsets

	// Network
	// Get total services
	// Get total endpoing
	// Get total Ingress

	// Volumes
	// Get total persistent volume
	// Get total pvc
	// Get total storage classes

	c.JSON(http.StatusOK, gin.H{
		"metrics": map[string]interface{}{
			"cpu_usage":    "0.5",
			"memory_usage": "1.2GB",
		},
	})
}

func (h *Handler) ExecuteKubectl(c *gin.Context) {
	var request struct {
		Args []string `json:"args"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	args := request.Args
	result, err := h.kubectlController.ExecuteCommand(args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
