package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/welibekov/cluster-api-provider-ptc/api/v1alpha1"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth"
	ptcclient "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
	ptccloud "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/cloud"
	tasktypes "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/task/types"
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

// Core VM creation logic inside PTCMachineReconciler
func (r *PTCMachineReconciler) createPTCInstance(
	ctx context.Context,
	ptcMachine *infrav1.PTCMachine,
	ptcCluster *infrav1.PTCCluster,
	allocatedIP string,
	userData string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ptcHost := "10.220.112.90:42099"
	ptcScheme := "http"
	ptcBasepath := ""

	// Create a new client instance with config
	cfg := ptcclient.DefaultTransportConfig().
		WithHost(ptcHost).
		WithBasePath(ptcBasepath).
		WithSchemes([]string{ptcScheme})

	apiClient := ptcclient.NewHTTPClientWithConfig(nil, cfg)

	// 1. Get allocated IP from CAPI IPAM
	allocatedIP, err := r.ensureIPAddress(ctx, ptcMachine)
	if err != nil {
		logger.Info("IP allocation in progress", "reason", err.Error())
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 2. Map Machine Spec to PTC API CreateVMParams
	params := operations.NewCreateVMParams()
	params.Name = ptcMachine.Name
	params.InstanceType = ptcMachine.Spec.InstanceType // e.g. "m1.medium"
	params.ImageName = ptcMachine.Spec.Image           // e.g. "ubuntu-22.04"
	params.Network = ptcCluster.Spec.Network.Name      // "vlan-2415"

	if ptcCluster.Spec.Network.Subnet != "" {
		params.Subnet = &ptcCluster.Spec.Network.Subnet
	}
	if ptcMachine.Spec.BootDiskSize > 0 {
		diskSize := int64(ptcMachine.Spec.BootDiskSize)
		params.BootDiskSize = &diskSize
	}

	params.SSHKey = util.Ptr(ptcMachine.Spec.SSHKey)
	params.IPAddress = &allocatedIP
	params.UserData = &userData

	// 3. Trigger VM creation call
	resp, err := apiClient.Operations.CreateVM(params, auth.NewAPIKeyAuthWriter(auth.NewTokenManager().LoadTokenMust()))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to call CreateVM API: %w", err)
	}

	// 4. Parse Task payload
	task, err := ptccloud.ToTask(resp.GetPayload())
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to parse task response: %w", err)
	}

	logger.Info("VM creation task initiated", "taskID", task.ID, "status", task.Status)

	err = task.Wait(ctx, func(target *tasktypes.Task) error {
		freshTask, err := ptccloud.DescribeTask(ctx, task.ID.String(), apiClient)
		if err != nil {
			return err
		}
		*target = *freshTask
		return nil
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error waiting for VM creation task: %w", err)
	}

	var taskOutput struct {
		InstanceID string `json:"instance-id"`
	}

	// Extract generated Instance ID from Task Output JSON payload
	if len(task.Output) > 0 {
		// Step 1: Decode the outer JSON string
		var rawJSONString string
		if err := json.Unmarshal(task.Output, &rawJSONString); err == nil {
			// Step 2: Decode the inner JSON object
			if err := json.Unmarshal([]byte(rawJSONString), &taskOutput); err != nil {
				logger.Error(err, "failed to unmarshal inner task output payload")
			}
		} else {
			// Fallback: Attempt single-step unmarshal if task output is direct JSON
			_ = json.Unmarshal(task.Output, &taskOutput)
		}
	}

	instanceID := taskOutput.InstanceID
	if instanceID == "" {
		instanceID = ptcMachine.Name // Fallback if API returns empty output
	}

	// 6. Update PTCMachine Status
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

func ptrString(s string) *string {
	return &s
}
