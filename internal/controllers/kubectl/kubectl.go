package kubectl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("kubectl-controller")

	// Service account token path in Kubernetes
	tokenFile     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type Controller struct {
	kubeconfig string
}

type CommandResult struct {
	Command string `json:"command"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func NewController(config *rest.Config) (*Controller, error) {
	kubeconfigPath := filepath.Join(os.TempDir(), "kubeconfig")

	// Check if running in cluster by looking for service account token
	inCluster := fileExists(tokenFile)

	if inCluster {
		// In-cluster configuration
		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account token: %v", err)
		}

		namespace, err := ioutil.ReadFile(namespaceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read namespace: %v", err)
		}

		kubeconfig := clientcmdapi.Config{
			APIVersion: "v1",
			Kind:       "Config",
			Clusters: map[string]*clientcmdapi.Cluster{
				"default": {
					Server:                config.Host,
					CertificateAuthority:  rootCAFile,
					InsecureSkipTLSVerify: false,
				},
			},
			AuthInfos: map[string]*clientcmdapi.AuthInfo{
				"default": {
					Token: string(token),
				},
			},
			Contexts: map[string]*clientcmdapi.Context{
				"default": {
					Cluster:   "default",
					AuthInfo:  "default",
					Namespace: string(namespace),
				},
			},
			CurrentContext: "default",
		}

		err = clientcmd.WriteToFile(kubeconfig, kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to write kubeconfig: %v", err)
		}
	} else {
		// Local development configuration
		// First try KUBECONFIG environment variable
		kubeConfigEnv := os.Getenv("KUBECONFIG")
		if kubeConfigEnv != "" {
			kubeconfigPath = kubeConfigEnv
		} else {
			// Then try default location
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %v", err)
			}
			kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
		}

		// Verify the kubeconfig exists
		if !fileExists(kubeconfigPath) {
			return nil, fmt.Errorf("kubeconfig not found at %s", kubeconfigPath)
		}
	}

	return &Controller{
		kubeconfig: kubeconfigPath,
	}, nil
}

// ExecuteCommand executes a kubectl command using the local kubectl binary
func (c *Controller) ExecuteCommand(command ...string) (*CommandResult, error) {
	args := append([]string{"--kubeconfig", c.kubeconfig}, command...)

	cmd := exec.Command("kubectl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &CommandResult{
		Command: fmt.Sprintf("kubectl %s", strings.Join(command, " ")),
		Output:  stdout.String(),
	}

	if err != nil {
		result.Error = stderr.String()
		log.Error(err, "Failed to execute kubectl command",
			"command", result.Command,
			"error", result.Error)
		return result, fmt.Errorf("failed to execute kubectl command: %v", err)
	}

	log.Info("Successfully executed kubectl command",
		"command", result.Command)

	return result, nil
}

func (c *Controller) GetLogs(podName, namespace string, previousLogs bool) (*CommandResult, error) {
	args := []string{"logs", podName, "-n", namespace}
	if previousLogs {
		args = append(args, "-p")
	}
	return c.ExecuteCommand(args...)
}

func (c *Controller) Cleanup() {
	if c.kubeconfig != os.Getenv("KUBECONFIG") &&
		!strings.Contains(c.kubeconfig, ".kube/config") {
		os.Remove(c.kubeconfig)
	}
}

// Helper function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
