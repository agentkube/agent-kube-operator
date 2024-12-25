package controllers

import (
	"context"
	"fmt"

	dto "agentkube.com/agent-kube-operator/internal/dto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodeController struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewNodeController(client client.Client, scheme *runtime.Scheme) *NodeController {
	return &NodeController{
		client: client,
		scheme: scheme,
	}
}

func (c *NodeController) GetNodes(ctx context.Context) ([]dto.NodeInfo, error) {
	var nodes corev1.NodeList
	if err := c.client.List(ctx, &nodes); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeInfos := make([]dto.NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		taints := make([]string, 0, len(node.Spec.Taints))
		for _, taint := range node.Spec.Taints {
			taints = append(taints, fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect))
		}

		roles := make([]string, 0)
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/control-plane" || label == "node-role.kubernetes.io/master" {
				roles = append(roles, "master")
			} else if label == "node-role.kubernetes.io/worker" {
				roles = append(roles, "worker")
			}
		}

		conditions := make([]dto.NodeCondition, 0, len(node.Status.Conditions))
		for _, condition := range node.Status.Conditions {
			conditions = append(conditions, dto.NodeCondition{
				Type:    string(condition.Type),
				Status:  string(condition.Status),
				Message: condition.Message,
			})
		}

		cpu := node.Status.Capacity.Cpu().String()
		memory := node.Status.Capacity.Memory().String()
		diskSize := "N/A"
		storage := node.Status.Capacity.StorageEphemeral()
		if !storage.IsZero() {
			diskSize = storage.String()
		}

		nodeInfo := dto.NodeInfo{
			Name:       node.Name,
			CPU:        cpu,
			Memory:     memory,
			Disk:       diskSize,
			Taints:     taints,
			Roles:      roles,
			Version:    node.Status.NodeInfo.KubeletVersion,
			Age:        node.CreationTimestamp.String(),
			Conditions: conditions,
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	return nodeInfos, nil
}
