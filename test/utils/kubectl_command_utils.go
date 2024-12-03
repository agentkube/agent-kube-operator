package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// KubectlCommand wraps kubectl command execution
type KubectlCommand struct {
	Namespace string
}

// NewKubectlCommand creates a new KubectlCommand instance
func NewKubectlCommand(namespace string) *KubectlCommand {
	return &KubectlCommand{
		Namespace: namespace,
	}
}

// execute runs kubectl command and returns output
func (k *KubectlCommand) execute(args ...string) (string, error) {
	if k.Namespace != "" {
		args = append([]string{"-n", k.Namespace}, args...)
	}
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl command failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// Create creates a resource from file
func (k *KubectlCommand) Create(filename string) (string, error) {
	return k.execute("create", "-f", filename)
}

// Apply applies a resource from file
func (k *KubectlCommand) Apply(filename string) (string, error) {
	return k.execute("apply", "-f", filename)
}

// Delete deletes a resource
func (k *KubectlCommand) Delete(resourceType, name string) (string, error) {
	return k.execute("delete", resourceType, name)
}

// DeleteWithFile deletes resources from file
func (k *KubectlCommand) DeleteWithFile(filename string) (string, error) {
	return k.execute("delete", "-f", filename)
}

// Get retrieves resource information
func (k *KubectlCommand) Get(resourceType, name string) (string, error) {
	return k.execute("get", resourceType, name)
}

// GetAll retrieves all resources of a type
func (k *KubectlCommand) GetAll(resourceType string) (string, error) {
	return k.execute("get", resourceType)
}

// Watch watches a resource for changes
func (k *KubectlCommand) Watch(resourceType, name string) (string, error) {
	return k.execute("get", resourceType, name, "--watch")
}

// Describe gets detailed information about a resource
func (k *KubectlCommand) Describe(resourceType, name string) (string, error) {
	return k.execute("describe", resourceType, name)
}

// Logs retrieves container logs
func (k *KubectlCommand) Logs(podName string, container string, tail int) (string, error) {
	args := []string{"logs", podName}
	if container != "" {
		args = append(args, "-c", container)
	}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	return k.execute(args...)
}

// GetEvents retrieves events for a resource
func (k *KubectlCommand) GetEvents(resourceType, name string) (string, error) {
	return k.execute("get", "events", "--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", name, resourceType))
}

// List resources with label selector
func (k *KubectlCommand) List(resourceType string, labelSelector string) (string, error) {
	args := []string{"get", resourceType}
	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}
	return k.execute(args...)
}

// Scale changes the number of replicas
func (k *KubectlCommand) Scale(resourceType, name string, replicas int) (string, error) {
	return k.execute("scale", resourceType, name, "--replicas", fmt.Sprintf("%d", replicas))
}

// GetResourceStatus gets status of a resource
func (k *KubectlCommand) GetResourceStatus(resourceType, name string) (string, error) {
	output, err := k.execute("get", resourceType, name, "-o", "jsonpath={.status}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}
