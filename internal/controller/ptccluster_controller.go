/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
	ptccloud "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/cloud"
)

const (
	ClusterFinalizer = "infrastructure.cluster.x-k8s.io/ptccluster"
)

// PTCClusterReconciler reconciles a PTCCluster object
type PTCClusterReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	PtcTokenManager *ptccloud.TokenManager
}
// +kubebuilder:rbac:groups=ipam.cluster.x-k8s.io,resources=ipaddressclaims;ipaddressclaims/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipam.cluster.x-k8s.io,resources=ipaddresses;ipaddresses/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PTCCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *PTCClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// TODO(user): your logic here

	// 1. Fetch the PTCCluster instance
	ptcCluster := &infrav1.PTCCluster{}
	if err := r.Get(ctx, req.NamespacedName, ptcCluster); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Fetch the owning CAPI Cluster
	capiCluster, err := util.GetOwnerCluster(ctx, r.Client, ptcCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get owner CAPI Cluster: %w", err)
	}

	if capiCluster == nil {
		// Fallback: Check if CAPI Cluster with matching name/namespace exists
		cluster := &clusterv1.Cluster{}
		clusterKey := types.NamespacedName{
			Namespace: ptcCluster.Namespace,
			Name:      ptcCluster.Name,
		}
		if err := r.Get(ctx, clusterKey, cluster); err == nil {
			capiCluster = cluster
		} else if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to fetch CAPI cluster %s: %w", clusterKey, err)
		} else {
			logger.Info("CAPI Cluster owner not yet created, waiting...")
			return ctrl.Result{}, nil
		}
	}

	// 3. Handle paused clusters (CAPI standard annotation)
	if annotations.IsPaused(capiCluster, ptcCluster) {
		logger.Info("PTCCluster or owner CAPI Cluster is paused, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// 4. Initialize Patch Helper (tracks changes and patches object status efficiently)
	patchHelper, err := patch.NewHelper(ptcCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create patch helper: %w", err)
	}

	// Guarded patch helper defer execution
	defer func() {
		// CRITICAL: If finalizers are empty during/after deletion, the object no longer
		// exists in etcd, so attempting patchHelper.Patch() will fail with NotFound.
		if !ptcCluster.DeletionTimestamp.IsZero() && len(ptcCluster.Finalizers) == 0 {
			return
		}

		if err := patchHelper.Patch(ctx, ptcCluster); err != nil {
			// Ignore NotFound errors since the resource may have been deleted during reconciliation
			// Unwrap or check if the root cause is a NotFound error
			if !apierrors.IsNotFound(err) && !strings.Contains(err.Error(), "not found") {
				logger.Error(err, "Failed to patch PTCCluster status")
			}
		}
	}()

	// 5. Handle Deletion Lifecycle
	if !ptcCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, ptcCluster)
	}

	// 6. Handle Normal Creation/Update Lifecycle
	return r.reconcileNormal(ctx, ptcCluster, capiCluster)
}

func (r *PTCClusterReconciler) reconcileNormal(ctx context.Context, ptcCluster *infrav1.PTCCluster, capiCluster *clusterv1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Step A: Ensure Finalizer is registered
	if !controllerutil.ContainsFinalizer(ptcCluster, ClusterFinalizer) {
		controllerutil.AddFinalizer(ptcCluster, ClusterFinalizer)
		return ctrl.Result{}, nil
	}

	// Step B: Ensure Network Infrastructure (VLAN / Subnet / Security Rules)
	// Example: verify network specified in ptcCluster.Spec.Network.Name exists
	logger.Info("Reconciling network infrastructure", "network", ptcCluster.Spec.Network.Name)

	// Default port to 6443 if not explicitly specified
	if ptcCluster.Spec.ControlPlaneEndpoint.Port == 0 {
		ptcCluster.Spec.ControlPlaneEndpoint.Port = 6443
	}

	// Step C: Expose ControlPlaneEndpoint in Status and mark infrastructure Ready
	ptcCluster.Status.ControlPlaneEndpoint = clusterv1.APIEndpoint{
		Host: ptcCluster.Spec.ControlPlaneEndpoint.Host,
		Port: ptcCluster.Spec.ControlPlaneEndpoint.Port,
	}

	ptcCluster.Status.Ready = true
	logger.Info("PTCCluster infrastructure reconciliation ready",
		"endpoint", ptcCluster.Status.ControlPlaneEndpoint.String())

	// ---------------------------------------------------------------------
	// 3. Verify target network exists via PTC API
	// ---------------------------------------------------------------------
	networkName := ptcCluster.Spec.Network.Name
	if networkName == "" {
		return ctrl.Result{}, fmt.Errorf("network name is required in PTCCluster spec")
	}

	pc, err := GetClientForCluster(ctx, r.Client, ptcCluster, r.PtcTokenManager)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Describe network by name to check that it exists.
	if _, err := pc.DescribeNetworkByName(ctx, networkName); err != nil {
		// Failed
		ptcCluster.Status.Ready = false
		logger.Error(err, "Failed to describe network on PTC API", "network", networkName)

		// Requeue after a short delay so the reconciler retries if network creation is in progress
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Set the cluster status to ready.
	ptcCluster.Status.Ready = true

	if err := r.Status().Update(ctx, ptcCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update PTCCluster status: %w", err)
	}
	return ctrl.Result{}, nil

}

func (r *PTCClusterReconciler) reconcileDelete(ctx context.Context, ptcCluster *infrav1.PTCCluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling PTCCluster deletion")

	// 1. Teardown dynamic network resources or LB created specifically for this cluster
	// (Ensure all PTCMachines are deleted before wiping network components)

	// 2. Remove Finalizer to release object from Kubernetes API Server
	controllerutil.RemoveFinalizer(ptcCluster, ClusterFinalizer)
	if err := r.Update(ctx, ptcCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}

	logger.Info("Successfully removed finalizer for PTCCluster")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PTCClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.PTCCluster{}).
		Named("ptccluster").
		Complete(r)
}
