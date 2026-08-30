package controller

import (
	"context"
	"fmt"
	"time"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Core VM creation logic inside PTCMachineReconciler
func (r *PTCMachineReconciler) createPTCInstance(
	ctx context.Context,
	ptcMachine *infrav1.PTCMachine,
	ptcCluster *infrav1.PTCCluster,
	allocatedIP string,
	userData string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Get allocated IP from CAPI IPAM
	allocatedIP, err := r.ensureIPAddress(ctx, ptcMachine)
	if err != nil {
		logger.Info("IP allocation in progress", "reason", err.Error())
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 2. Map Machine Spec to PTC API CreateVMParams
	params := &operations.CreateVMParams{
		Name:         ptcMachine.Name,
		InstanceType: ptcMachine.Spec.InstanceType, // e.g. "m1.medium"
		ImageName:    ptcMachine.Spec.Image,        // e.g. "ubuntu-22.04"
		Network:      ptcCluster.Spec.Network.Name, // "vlan-2415"
		SSHKey:       util.Ptr(ptcMachine.Spec.SSHKey),
		IPAddress:    util.Ptr(allocatedIP),
		UserData:     util.Ptr(userData),
	}
	if ptcCluster.Spec.Network.Subnet != "" {
		params.Subnet = util.Ptr(ptcCluster.Spec.Network.Subnet)
	}
	if ptcMachine.Spec.BootDiskSize > 0 {
		diskSize := int64(ptcMachine.Spec.BootDiskSize)
		params.BootDiskSize = util.Ptr(diskSize)
	}

	// 3. Trigger VM creation call
	pc, err := GetClientForCluster(ctx, r.Client, ptcCluster, r.PtcTokenManager)
	if err != nil {
		return ctrl.Result{}, err
	}

	rawInstanceID, err := pc.CreateInstance(ctx, params)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to call CreateVM API: %w", err)
	}

	// 6. Update PTCMachine Status
	instanceID := util.Deref(rawInstanceID)

	ptcMachine.Spec.ProviderID = util.Ptr(fmt.Sprintf("ptc:///%s", instanceID))
	ptcMachine.Status.InstanceID = instanceID
	ptcMachine.Status.Addresses = []clusterv1.MachineAddress{
		{
			Type:    clusterv1.MachineInternalIP,
			Address: allocatedIP,
		},
		{
			Type:    clusterv1.MachineHostName,
			Address: ptcMachine.Name,
		},
	}
	ptcMachine.Status.Ready = true

	logger.Info("Successfully created PTC instance", "instanceID", instanceID, "ip", allocatedIP)
	return ctrl.Result{}, nil
}
