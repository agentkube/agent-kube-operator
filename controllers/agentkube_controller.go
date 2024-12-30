package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AgentDeploymentController reconciles agent deployments
type AgentDeploymentController struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles the agent deployment lifecycle
func (r *AgentDeploymentController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// createIfNotExists creates a resource if it doesn't exist
func (r *AgentDeploymentController) CreateIfNotExists(ctx context.Context, obj client.Object) error {
	err := r.Create(ctx, obj)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// createOrUpdate creates or updates a resource
func (r *AgentDeploymentController) CreateOrUpdate(ctx context.Context, obj client.Object) error {
	err := r.Create(ctx, obj)
	if err != nil && errors.IsAlreadyExists(err) {
		return r.Update(ctx, obj)
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AgentDeploymentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
