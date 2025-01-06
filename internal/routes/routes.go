package routes

import (
	"fmt"
	"net/http"

	handlers "agentkube.com/agent-kube-operator/internal/handlers"
	cors "github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	runtime "k8s.io/apimachinery/pkg/runtime"
	rest "k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("routes")

type Router struct {
	router  *gin.Engine
	handler *handlers.Handler
}

func NewRouter(client client.Client, scheme *runtime.Scheme, config *rest.Config) (*Router, error) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	handler, err := handlers.NewHandler(client, scheme, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %v", err)
	}

	r := &Router{
		router:  router,
		handler: handler,
	}

	r.setupRoutes()
	return r, nil
}

func (r *Router) setupRoutes() {
	r.router.GET("/health", r.handler.HealthCheck)
	r.router.GET("/ready", r.handler.ReadyCheck)

	// need a prometheus metrics to get all the network details
	v1 := r.router.Group("/api/v1")
	{

		metrics := v1.Group("/metrics")
		{
			metrics.GET("/pods", r.handler.GetPodMetrics)
			metrics.GET("/pods/history", r.handler.GetHistoricalPodMetrics)
			metrics.GET("/pods/comprehensive", r.handler.GetComprehensivePodMetrics)
		}

		cluster := v1.Group("/cluster")
		{
			cluster.GET("/info", r.handler.GetClusterInfo)
			cluster.GET("/metrics", r.handler.GetClusterMetrics)
			cluster.GET("/namespace-metrics", r.handler.GetNamespaceMetrics)
			cluster.POST("/resources", r.handler.ListResources)
			cluster.GET("/nodes", r.handler.GetNodes)
		}

		v1.GET("/resources/:group/:version/:resource_type/:resource_name", r.handler.GetK8sResource)
		v1.GET("/namespaces/:namespace/resources/:group/:version/:resource_type/:resource_name", r.handler.GetK8sResource)
		v1.GET("/resources", r.handler.ListAPIResources)

		v1.PUT("/namespaces/:namespace/resources/:group/:version/:resource_type/:resource_name", r.handler.ApplyK8sResource)
		v1.PUT("/resources/:group/:version/:resource_type/:resource_name", r.handler.ApplyK8sResource)
	}

	v1.POST("/kubectl", r.handler.ExecuteKubectl)
}

func (r *Router) StartServer(addr string) error {
	go func() {
		if err := r.router.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Error(err, "Failed to start HTTP server")
		}
	}()

	log.Info("Started HTTP server", "address", addr)
	return nil
}
