package routes

import (
	"net/http"

	"agentkube.com/agent-kube-operator/internal/handlers"
	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("routes")

type Router struct {
	router  *gin.Engine
	handler *handlers.Handler
}

func NewRouter(client client.Client, scheme *runtime.Scheme) *Router {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	r := &Router{
		router:  router,
		handler: handlers.NewHandler(client, scheme),
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Health check endpoints
	r.router.GET("/health", r.handler.HealthCheck)
	r.router.GET("/ready", r.handler.ReadyCheck)

	// Requires a  middleware to verify if any write

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
			// Memory/CPU utilization and total Pods/Deployment/Daemonset/Statefulset running per namespace (return something like { namespace, metrcis: { cpu, memory }, workloads: { pods, deployment, .... }})
			// Get All namespaces
			// Get all kubernetes resources for every namespace (resources, namespaces(by default: default ns)) -> returns resources json the items: array[]
		}
	}
	v1.POST("/kubectl", r.handler.ExecuteKubectl)

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
