package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AgentDeploymentController reconciles agent deployments
type AgentDeploymentController struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles the agent deployment lifecycle
func (r *AgentDeploymentController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create ServiceAccount if doesn't exist
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agentkube-controller-sa",
			Namespace: req.Namespace,
		},
	}
	if err := r.createIfNotExists(ctx, sa); err != nil {
		log.Error(err, "Failed to ensure ServiceAccount")
		return ctrl.Result{}, err
	}

	// Create Role if doesn't exist
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agentkube-controller-role",
			Namespace: req.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "nodes", "events"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "statefulsets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	if err := r.createIfNotExists(ctx, role); err != nil {
		log.Error(err, "Failed to ensure Role")
		return ctrl.Result{}, err
	}

	// Create RoleBinding if doesn't exist
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agentkube-controller-rolebinding",
			Namespace: req.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: req.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		},
	}
	if err := r.createIfNotExists(ctx, rb); err != nil {
		log.Error(err, "Failed to ensure RoleBinding")
		return ctrl.Result{}, err
	}

	// Create or update the agent Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agentkube-controller",
			Namespace: req.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "agentkube-controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "agentkube-controller",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: sa.Name,
					Containers: []corev1.Container{
						{
							Name: "agent",
							// Image: "agentkube-controller:latest", // You might want to make this configurable
							Image:     "nginx:latest", // You might want to make this configurable
							Resources: corev1.ResourceRequirements{
								// Add resource limits/requests as needed
							},
						},
					},
				},
			},
		},
	}

	if err := r.createOrUpdate(ctx, deployment); err != nil {
		log.Error(err, "Failed to ensure Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createIfNotExists creates a resource if it doesn't exist
func (r *AgentDeploymentController) createIfNotExists(ctx context.Context, obj client.Object) error {
	err := r.Create(ctx, obj)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// createOrUpdate creates or updates a resource
func (r *AgentDeploymentController) createOrUpdate(ctx context.Context, obj client.Object) error {
	err := r.Create(ctx, obj)
	if err != nil && errors.IsAlreadyExists(err) {
		return r.Update(ctx, obj)
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AgentDeploymentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Watch deployments in case they get deleted or modified
		For(&appsv1.Deployment{}).
		Complete(r)
}
