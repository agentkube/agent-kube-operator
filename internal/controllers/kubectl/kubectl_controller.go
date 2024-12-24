package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("kubectl-controller")

type Controller struct{}

type CommandResult struct {
	Command string `json:"command"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func NewController() *Controller {
	return &Controller{}
}

// ExecuteCommand executes a kubectl command and returns the result
func (c *Controller) ExecuteCommand(command ...string) (*CommandResult, error) {
	// Construct the command
	cmd := exec.Command("kubectl", command...)

	// Create buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Prepare the result
	result := &CommandResult{
		Command: fmt.Sprintf("kubectl %s", strings.Join(command, " ")),
		Output:  stdout.String(),
	}

	// Handle any errors
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

// Some helper methods for common operations
func (c *Controller) GetPods(namespace string) (*CommandResult, error) {
	return c.ExecuteCommand("get", "pods", "-n", namespace)
}

func (c *Controller) GetNodes() (*CommandResult, error) {
	return c.ExecuteCommand("get", "nodes")
}

func (c *Controller) GetNamespaces() (*CommandResult, error) {
	return c.ExecuteCommand("get", "namespaces")
}

func (c *Controller) DescribeResource(resourceType, name, namespace string) (*CommandResult, error) {
	if namespace != "" {
		return c.ExecuteCommand("describe", resourceType, name, "-n", namespace)
	}
	return c.ExecuteCommand("describe", resourceType, name)
}

func (c *Controller) GetLogs(podName, namespace string, previousLogs bool) (*CommandResult, error) {
	args := []string{"logs", podName, "-n", namespace}
	if previousLogs {
		args = append(args, "-p")
	}
	return c.ExecuteCommand(args...)
}
