package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"agentkube.com/agent-kube-operator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type ClusterRegistration struct {
	ClusterName      string `json:"clusterName"`
	AccessType       string `json:"accessType"`
	ExternalEndpoint string `json:"externalEndpoint"`
}

func getExternalEndpoint(clientset *kubernetes.Clientset) (string, error) {
	ingress, err := clientset.NetworkingV1().Ingresses("system").Get(context.Background(), "controller-manager-ingress", metav1.GetOptions{})
	if err != nil {
		return "http://controller-manager-service.system.svc.cluster.local:8082", nil
	}

	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
			return fmt.Sprintf("http://%s:8082", ingress.Status.LoadBalancer.Ingress[0].IP), nil
		}
		if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
			return fmt.Sprintf("http://%s:8082", ingress.Status.LoadBalancer.Ingress[0].Hostname), nil
		}
	}

	// Fall back to cluster-local endpoint if no LoadBalancer is available
	return "http://controller-manager-service.system.svc.cluster.local:8082", nil
}

func RegisterCluster(clientset *kubernetes.Clientset) error {
	serverEndpoint := utils.GetEnviron("AGENTKUBE_SERVER_ENDPOINT")
	apiKey := utils.GetEnviron("AGENTKUBE_API_KEY")
	clusterName := utils.GetEnviron("CLUSTER_NAME")
	accessType := utils.GetEnviron("ACCESS_TYPE")

	if serverEndpoint == "" || apiKey == "" || clusterName == "" || accessType == "" {
		return fmt.Errorf("missing required environment variables")
	}

	externalEndpoint, err := getExternalEndpoint(clientset)
	if err != nil {
		return fmt.Errorf("failed to get external endpoint: %v", err)
	}

	registration := ClusterRegistration{
		ClusterName:      clusterName,
		AccessType:       accessType,
		ExternalEndpoint: externalEndpoint,
	}

	jsonData, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal registration data: %v", err)
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/register-cluster", serverEndpoint),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	setupLog.Info("cluster registered successfully",
		"externalEndpoint", externalEndpoint)
	return nil
}
