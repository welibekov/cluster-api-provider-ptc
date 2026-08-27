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
	"time"

	corev1 "k8s.io/api/core/v1"
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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
	ptcutil "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
)

const (
	MachineFinalizer = "infrastructure.cluster.x-k8s.io/ptcmachine"
)

// PTCMachineReconciler reconciles a PTCMachine object
type PTCMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=ptcmachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PTCMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *PTCMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// TODO(user): your logic here

	ptcMachine := &infrav1.PTCMachine{}
	if err := r.Get(ctx, req.NamespacedName, ptcMachine); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 1A. CHECK DELETION FIRST (Before checking CAPI owner)
	if !ptcMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, ptcMachine)
	}

	// 2A. Ensure finalizer exists for active resources
	if !controllerutil.ContainsFinalizer(ptcMachine, MachineFinalizer) {
		controllerutil.AddFinalizer(ptcMachine, MachineFinalizer)
		if err := r.Update(ctx, ptcMachine); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
	}

	// 2. Fetch the owning CAPI Machine resource
	capiMachine, err := util.GetOwnerMachine(ctx, r.Client, ptcMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get owner Machine: %w", err)
	}

	if capiMachine == nil {
		// Fallback: Check if CAPI Machine with matching name/namespace exists
		// (In CAPI, the infra capiMachine typically shares the exact same Name and Namespace as the core Machine)
		newCapiMachine := &clusterv1.Machine{}
		newCapiMachineKey := types.NamespacedName{
			Namespace: ptcMachine.Namespace,
			Name:      ptcMachine.Name,
		}

		if err := r.Get(ctx, newCapiMachineKey, newCapiMachine); err == nil {
			capiMachine = newCapiMachine
		} else if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to fetch CAPI Machine %s: %w", newCapiMachineKey, err)
		} else {
			machineList := &clusterv1.MachineList{}
			if err := r.List(ctx, machineList, client.InNamespace(ptcMachine.Namespace)); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to list CAPI machines: %w", err)
			}

			for i := range machineList.Items {
				m := &machineList.Items[i]
				ref := m.Spec.InfrastructureRef
				if ref.Kind == "PTCMachine" && ref.Name == ptcMachine.Name {
					capiMachine = m
					break
				}
			}

		}

		if capiMachine == nil {
			// Requeue to check again once CAPI creates the Machine resource
			logger.Info("CAPI Machine owner not yet created, waiting...")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

	}

	// 3. Fetch owning CAPI Cluster
	capiCluster, err := util.GetClusterFromMetadata(ctx, r.Client, ptcMachine.ObjectMeta)
	if err != nil && capiMachine != nil && capiMachine.Spec.ClusterName != "" {
		// Fallback: Fetch Cluster directly using capiMachine.Spec.ClusterName
		clusterKey := types.NamespacedName{
			Namespace: ptcMachine.Namespace,
			Name:      capiMachine.Spec.ClusterName,
		}
		capiCluster = &clusterv1.Cluster{}
		if err := r.Get(ctx, clusterKey, capiCluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get CAPI cluster %s via fallback: %w", clusterKey, err)
		}
		err = nil
	}
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get owner CAPI Cluster: %w", err)
	}

	if capiCluster == nil {
		logger.Info("CAPI Cluster owner not yet set, waiting...")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	if !capiCluster.Status.InfrastructureReady {
		logger.Info("CAPI Cluster infrastructure is not ready yet, waiting...")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 4. Check if paused
	if annotations.IsPaused(capiCluster, ptcMachine) {
		logger.Info("PTCMachine or owner CAPI Cluster is paused, skipping reconciliation")
		return ctrl.Result{}, nil
	}
	// 5. Setup Patch Helper
	patchHelper, err := patch.NewHelper(ptcMachine, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create patch helper: %w", err)
	}

	defer func() {
		if err := patchHelper.Patch(ctx, ptcMachine); err != nil {
			logger.Error(err, "Failed to patch PTCMachine status")
		}
	}()

	// 6. Handle Deletion
	if !ptcMachine.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, ptcMachine)
	}

	// 7. Handle Creation / Update
	return r.reconcileNormal(ctx, ptcMachine, capiMachine, capiCluster)
}

func (r *PTCMachineReconciler) reconcileNormal(
	ctx context.Context,
	ptcMachine *infrav1.PTCMachine,
	capiMachine *clusterv1.Machine,
	capiCluster *clusterv1.Cluster,
) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// If already created and marked ready, nothing left to do
	if ptcMachine.Status.Ready && ptcMachine.Status.InstanceID != "" {
		return ctrl.Result{}, nil
	}
	// Step Aa: Allocate IP address from CAPI IPAM
	allocatedIP, err := r.ensureIPAddress(ctx, ptcMachine)
	if err != nil {
		logger.Info("Waiting on IP allocation", "reason", err.Error())
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Step A: Register Finalizer
	if !controllerutil.ContainsFinalizer(ptcMachine, MachineFinalizer) {
		controllerutil.AddFinalizer(ptcMachine, MachineFinalizer)
		return ctrl.Result{}, nil
	}

	// Step B: Ensure Infrastructure Cluster is Ready
	ptcCluster := &infrav1.PTCCluster{}
	ptcClusterKey := types.NamespacedName{
		Namespace: ptcMachine.Namespace,
		Name:      capiCluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, ptcClusterKey, ptcCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch PTCCluster %s: %w", ptcClusterKey, err)
	}

	if !ptcCluster.Status.Ready {
		logger.Info("PTCCluster is not ready yet, waiting before provisioning machine...")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	// Step C: Check Bootstrap Data Readiness (Cloud-Init Secret)
	if capiMachine.Spec.Bootstrap.DataSecretName == nil {
		logger.Info("Waiting for CAPI core to set Bootstrap.DataSecretName...")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Step B: Ensure UserData (bootstrap cloud-init) is available from CAPI Machine
	userData, err := r.getBootstrapData(ctx, ptcMachine.Namespace, ptcutil.Deref(capiMachine.Spec.Bootstrap.DataSecretName))
	if err != nil || userData == "" {
		logger.Info("Waiting for bootstrap data secret to be ready")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	logger.Info("Creating VM instance on PTC Cloud", "machine", ptcMachine.Name)

	// Step C: Dispatch or check in-flight async VM creation task
	return r.createPTCInstance(ctx, ptcMachine, ptcCluster, allocatedIP, userData)

	//// 1. Resolve credentials via PTCCluster identityRef
	//creds, err := FetchCredentials(ctx, r.Client, ptcCluster.Spec.IdentityRef, ptcCluster.Namespace)
	//if err != nil {
	//	return ctrl.Result{}, fmt.Errorf("failed to fetch credentials: %w", err)
	//}

	//_ = creds // Construct PTC API client here

	//// 2. Call PTC API to provision VM with userData, CPU/RAM specs, and network settings
	//// instanceID, nodeIP, err := ptcClient.CreateVM(ctx, ...)
	//instanceID := "vm-demo-123" // Replace with real API call result
	//nodeIP := "10.220.112.160"  // Replace with assigned VM IP

	//ptcMachine.Status.InstanceID = instanceID
	//ptcMachine.Status.Addresses = []clusterv1.MachineAddress{
	//	{
	//		Type:    clusterv1.MachineInternalIP,
	//		Address: nodeIP,
	//	},
	//	{
	//		Type:    clusterv1.MachineHostName,
	//		Address: ptcMachine.Name,
	//	},
	//}

	//// Step E: Set Machine Status Ready
	//ptcMachine.Status.Ready = true

	//return ctrl.Result{}, nil
}

func (r *PTCMachineReconciler) reconcileDelete(ctx context.Context, ptcMachine *infrav1.PTCMachine) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling PTCMachine deletion", "instanceID", ptcMachine.Status.InstanceID)

	if ptcMachine.Status.InstanceID != "" {
		// Call PTC API to delete VM instance
		// err := ptcClient.DeleteVM(ctx, ptcMachine.Status.InstanceID)
	}

	// Remove finalizer to allow Kubernetes resource removal
	controllerutil.RemoveFinalizer(ptcMachine, MachineFinalizer)
	if err := r.Update(ctx, ptcMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}

	logger.Info("Successfully removed finalizer for PTCMachine")

	return ctrl.Result{}, nil
}

// Helper function to read base64 decoded cloud-init data from CAPI secret
func (r *PTCMachineReconciler) getBootstrapData(ctx context.Context, namespace, secretName string) (string, error) {
	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{Namespace: namespace, Name: secretName}

	if err := r.Get(ctx, secretKey, secret); err != nil {
		return "", fmt.Errorf("failed to get bootstrap secret %s: %w", secretKey, err)
	}

	value, ok := secret.Data["value"]
	if !ok {
		return "", fmt.Errorf("secret %s missing 'value' key", secretKey)
	}

	return string(value), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PTCMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.PTCMachine{}).
		Named("ptcmachine").
		Complete(r)
}
