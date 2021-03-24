package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/crane"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DaemonSetReconciler reconciles a DaemonSet object
type DaemonSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("daemonset", req.NamespacedName)

	// Ignore DaemonSets in namespace "kube-system"
	if req.Namespace == "kube-system" {
		r.Log.Info(`Ignoring DaemonSet in namespace "kube-system"`,
			"daemonset", req.NamespacedName,
		)
		return ctrl.Result{}, nil
	}

	// Get DaemonSet
	ds := &appsv1.DaemonSet{}
	err := r.Get(ctx, req.NamespacedName, ds)
	// Don't requeue if DaemonSet does not exist
	if errors.IsNotFound(err) {
		r.Log.Error(err, "Could not find DaemonSet")
		return ctrl.Result{}, nil
	}

	// Requeue on error
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get keychain for dockerhub repo
	k8sc, err := k8schain.NewInCluster(ctx, k8schain.Options{
		Namespace:        "image-clone-controller-system",
		ImagePullSecrets: []string{"image-clone-controller-regcred"},
	})
	if err != nil {
		r.Log.Error(err, "Unable to create keychain")
		return ctrl.Result{}, err
	}

	for i, container := range ds.Spec.Template.Spec.Containers {
		// Skip containers already using cloned images
		if strings.HasPrefix(container.Image, "imageclone/") {
			r.Log.Info("Already using cloned image",
				"daemonset", req.NamespacedName,
				"image", container.Image,
			)
			continue
		}

		clone := fmt.Sprintf("%s/%s",
			"imageclone",
			strings.ReplaceAll(container.Image, "/", "_"),
		)
		r.Log.Info("Exchanging container image",
			"namespace", req.Namespace,
			"container name", container.Name,
			"before", container.Image,
			"after", clone,
		)

		// Clone image in own repo
		err = crane.Copy(container.Image, clone, crane.WithAuthFromKeychain(k8sc))
		if err != nil {
			r.Log.Error(err, "Unable to clone image")
			return ctrl.Result{}, err
		}

		// Modify Container Image to use our cloned image
		ds.Spec.Template.Spec.Containers[i].Image = clone
	}

	// Update the DaemonSet
	err = r.Update(ctx, ds)
	if errors.IsConflict(err) {
		// Optimistic concurrency
		r.Log.Info("DaemonSet has been updated since getting it")
		return ctrl.Result{Requeue: true}, nil
	}
	if err != nil {
		r.Log.Error(err, "Unable to update DaemonSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Complete(r)
}
