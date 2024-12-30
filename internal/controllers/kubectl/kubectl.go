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
	cmd := exec.Command("kubectl", command...)

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
