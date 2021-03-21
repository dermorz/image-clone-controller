package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/crane"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("deployment", req.NamespacedName)

	// Ignore Deployments in namespace "kube-system"
	if req.Namespace == "kube-system" {
		r.Log.Info(`Ignoring Deployment in namespace "kube-system"`,
			"deployment", req.NamespacedName,
		)
		return ctrl.Result{}, nil
	}

	// Get Deployment
	depl := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, depl)
	// Don't requeue if Deployment does not exist
	if errors.IsNotFound(err) {
		r.Log.Error(err, "Could not find Deployment")
		return ctrl.Result{}, nil
	}

	// Requeue on error
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Reconciling Deployment",
		"container name", depl.Spec.Template.Spec.Containers[0].Name,
		"container image", depl.Spec.Template.Spec.Containers[0].Image,
	)

	for i, container := range depl.Spec.Template.Spec.Containers {
		// Skip containers already using cloned images
		if strings.HasPrefix(container.Image, "imageclone/") {
			continue
		}

		// Clone image in own repo
		clone := fmt.Sprintf("%s/%s",
			"imageclone",
			strings.ReplaceAll(container.Image, "/", "_"),
		)
		err := crane.Copy(container.Image, clone)
		if err != nil {
			r.Log.Error(err, "Unable to clone image")
			return ctrl.Result{}, err
		}

		depl.Spec.Template.Spec.Containers[i].Image = clone
	}

	// Update the Deployment
	err = r.Update(ctx, depl)
	if err != nil {
		r.Log.Error(err, "Unable to update Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
