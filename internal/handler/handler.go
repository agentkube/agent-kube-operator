package handler

import (
	"net/http"

	"agentkube.com/agent-kube-operator/utils"
	"github.com/gin-gonic/gin"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("api-handler")

type Handler struct {
	router *gin.Engine
}

func NewHandler() *Handler {
	router := gin.New()
	router.Use(gin.Recovery())

	h := &Handler{
		router: router,
	}

	// Setup routes
	h.setupRoutes()

	return h
}

func (h *Handler) setupRoutes() {
	// Health check endpoints
	h.router.GET("/health", h.HealthCheck)
	h.router.GET("/ready", h.ReadyCheck)

	// API v1 routes
	v1 := h.router.Group("/api/v1")
	{
		// Agents endpoints
		agents := v1.Group("/agents")
		{
			agents.GET("", h.ListAgents)
			agents.GET("/:name", h.GetAgent)
			agents.POST("", h.CreateAgent)
			agents.PUT("/:name", h.UpdateAgent)
			agents.DELETE("/:name", h.DeleteAgent)
		}

		// Cluster info endpoints
		cluster := v1.Group("/cluster")
		{
			cluster.GET("/info", h.GetClusterInfo)
			cluster.GET("/metrics", h.GetClusterMetrics)
		}
	}
}

// StartServer starts the HTTP server in a goroutine
func (h *Handler) StartServer(addr string) error {
	go func() {
		if err := h.router.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Error(err, "Failed to start HTTP server")
		}
	}()

	log.Info("Started HTTP server", "address", addr)
	return nil
}

// Handler functions

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
	c.JSON(http.StatusOK, gin.H{
		"metrics": map[string]interface{}{
			"cpu_usage":    "0.5",
			"memory_usage": "1.2GB",
		},
	})
}
