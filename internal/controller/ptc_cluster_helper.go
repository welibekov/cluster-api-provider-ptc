package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
)

// getPTCCluster retrieves the target PTCCluster resource associated with a PTCMachine.
func (r *PTCMachineReconciler) getPTCCluster(ctx context.Context, ptcMachine *infrav1.PTCMachine) (*infrav1.PTCCluster, error) {
	// 1. Fetch the owner core CAPI Machine object
	capiMachine, err := util.GetOwnerMachine(ctx, r.Client, ptcMachine.ObjectMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner CAPI Machine for PTCMachine: %w", err)
	}
	if capiMachine == nil {
		return nil, fmt.Errorf("owner CAPI Machine object not found on PTCMachine %s/%s", ptcMachine.Namespace, ptcMachine.Name)
	}

	// 2. Fetch the owner core CAPI Cluster using capiMachine.Spec.ClusterName
	capiCluster, err := util.GetClusterByName(ctx, r.Client, capiMachine.Namespace, capiMachine.Spec.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("owner CAPI Cluster %s/%s not found: %w", capiMachine.Namespace, capiMachine.Spec.ClusterName, err)
	}

	// 3. Verify InfrastructureRef exists on the CAPI Cluster
	if capiCluster.Spec.InfrastructureRef == nil {
		return nil, fmt.Errorf("cluster.Spec.InfrastructureRef is nil for cluster %s", capiCluster.Name)
	}

	ptcClusterNamespace := capiCluster.Spec.InfrastructureRef.Namespace
	if ptcClusterNamespace == "" {
		ptcClusterNamespace = capiCluster.Namespace
	}

	// 4. Fetch the PTCCluster resource
	ptcClusterKey := types.NamespacedName{
		Namespace: ptcClusterNamespace,
		Name:      capiCluster.Spec.InfrastructureRef.Name,
	}

	ptcCluster := &infrav1.PTCCluster{}
	if err := r.Get(ctx, ptcClusterKey, ptcCluster); err != nil {
		return nil, fmt.Errorf("failed to fetch PTCCluster %s: %w", ptcClusterKey, err)
	}

	return ptcCluster, nil
}
