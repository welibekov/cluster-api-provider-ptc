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
	// 1. Fetch the owner core CAPI Cluster (e.g., via ownerReferences/labels)
	cluster, err := util.GetOwnerCluster(ctx, r.Client, ptcMachine.ObjectMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner CAPI cluster for PTCMachine: %w", err)
	}
	if cluster == nil {
		return nil, fmt.Errorf("owner CAPI Cluster object not found on PTCMachine %s/%s", ptcMachine.Namespace, ptcMachine.Name)
	}

	// 2. Extract the name of the PTCCluster from spec.infrastructureRef
	if cluster.Spec.InfrastructureRef == nil {
		return nil, fmt.Errorf("cluster.Spec.InfrastructureRef is nil for cluster %s", cluster.Name)
	}

	ptcClusterKey := types.NamespacedName{
		Namespace: ptcMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}

	// 3. Fetch the PTCCluster CRD object from the API server
	ptcCluster := &infrav1.PTCCluster{}
	if err := r.Get(ctx, ptcClusterKey, ptcCluster); err != nil {
		return nil, fmt.Errorf("failed to fetch PTCCluster %s: %w", ptcClusterKey, err)
	}

	return ptcCluster, nil
}
