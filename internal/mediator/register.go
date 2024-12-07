package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"agentkube.com/agent-kube-operator/utils"
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

func RegisterCluster() error {
	serverEndpoint := utils.GetEnviron("AGENTKUBE_SERVER_ENDPOINT")
	apiKey := utils.GetEnviron("AGENTKUBE_API_KEY")
	clusterName := utils.GetEnviron("CLUSTER_NAME")
	accessType := utils.GetEnviron("ACCESS_TYPE")
	externalEndpoint := utils.GetEnviron("EXTERNAL_ENDPOINT")

	if serverEndpoint == "" || apiKey == "" || clusterName == "" || accessType == "" || externalEndpoint == "" {
		return fmt.Errorf("missing required environment variables")
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

	// Create request
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/register-cluster", serverEndpoint),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	// request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	setupLog.Info("cluster registered successfully")
	return nil
}
