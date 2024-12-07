package routes

import (
	"net/http"

	"agentkube.com/agent-kube-operator/internal/handlers"
	"github.com/gin-gonic/gin"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("routes")

type Router struct {
	router  *gin.Engine
	handler *handlers.Handler
}

func NewRouter() *Router {
	router := gin.New()
	router.Use(gin.Recovery())

	r := &Router{
		router:  router,
		handler: handlers.NewHandler(),
	}

	// Setup routes
	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	// Health check endpoints
	r.router.GET("/health", r.handler.HealthCheck)
	r.router.GET("/ready", r.handler.ReadyCheck)

	// API v1 routes
	v1 := r.router.Group("/api/v1")
	{
		// Agents endpoints
		agents := v1.Group("/agents")
		{
			agents.GET("", r.handler.ListAgents)
			agents.GET("/:name", r.handler.GetAgent)
			agents.POST("", r.handler.CreateAgent)
			agents.PUT("/:name", r.handler.UpdateAgent)
			agents.DELETE("/:name", r.handler.DeleteAgent)
		}

		// Cluster info endpoints
		cluster := v1.Group("/cluster")
		{
			cluster.GET("/info", r.handler.GetClusterInfo)
			cluster.GET("/metrics", r.handler.GetClusterMetrics)
		}
	}
}

// StartServer starts the HTTP server in a goroutine
func (r *Router) StartServer(addr string) error {
	go func() {
		if err := r.router.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Error(err, "Failed to start HTTP server")
		}
	}()

	log.Info("Started HTTP server", "address", addr)
	return nil
}
