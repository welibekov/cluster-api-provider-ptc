package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1alpha1"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
)

// Helper: Ensure IPAddressClaim exists and retrieve assigned IP address
func (r *PTCMachineReconciler) ensureIPAddress(ctx context.Context, ptcMachine *infrav1.PTCMachine) (string, error) {
	claimName := fmt.Sprintf("%s-ip-claim", ptcMachine.Name)
	claim := &ipamv1.IPAddressClaim{}
	claimKey := types.NamespacedName{Namespace: ptcMachine.Namespace, Name: claimName}

	err := r.Get(ctx, claimKey, claim)
	if apierrors.IsNotFound(err) {
		// Create IPAddressClaim referencing the pool specified in PTCMachine spec
		newClaim := &ipamv1.IPAddressClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      claimName,
				Namespace: ptcMachine.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(ptcMachine, infrav1.GroupVersion.WithKind("PTCMachine")),
				},
			},
			Spec: ipamv1.IPAddressClaimSpec{
				PoolRef: util.Deref(ptcMachine.Spec.Network.IPFromPoolRef),
			},
		}
		if err := r.Create(ctx, newClaim); err != nil {
			return "", fmt.Errorf("failed to create IPAddressClaim: %w", err)
		}
		return "", fmt.Errorf("IPAddressClaim created, waiting for IPAM allocation")
	} else if err != nil {
		return "", err
	}

	// Claim exists; verify if IPAddress object has been bound
	if claim.Status.AddressRef.Name == "" {
		return "", fmt.Errorf("waiting for IPAM controller to allocate IPAddress")
	}

	ipObj := &ipamv1.IPAddress{}
	ipKey := types.NamespacedName{Namespace: ptcMachine.Namespace, Name: claim.Status.AddressRef.Name}
	if err := r.Get(ctx, ipKey, ipObj); err != nil {
		return "", fmt.Errorf("failed to fetch allocated IPAddress %s: %w", claim.Status.AddressRef.Name, err)
	}

	return ipObj.Spec.Address, nil
}
