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
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/opt"
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

// Helper: Wait/Poll for PTC Task execution until terminal state
func (r *PTCMachineReconciler) waitForTaskCompletion(ctx context.Context, apiClient *ptcclient.Ptc, taskID string) (*tasktypes.Task, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Fetch task status from API
			params := operations.NewDescribeTaskParams().WithContext(ctx)
			params.TaskID = taskID
			resp, err := apiClient.Operations.DescribeTask(params, auth.NewAPIKeyAuthWriter(auth.NewTokenManager().LoadTokenMust()))
			if err != nil {
				return nil, fmt.Errorf("error querying task %s status: %w", taskID, err)
			}
			task := &tasktypes.Task{}
			if err := opt.Rest2Task(resp.GetPayload(), task); err != nil {
				return nil, err
			}

			switch task.Status {
			case "completed":
				return task, nil
			case "failed":
				errMsg := "unknown task failure"
				if task.ErrMessage != nil {
					errMsg = *task.ErrMessage
				}
				return nil, fmt.Errorf("ptc task %s failed: %s", taskID, errMsg)
			case "accepted", "running":
				// Continue polling loop
				continue
			default:
				return nil, fmt.Errorf("unhandled task status %q for task %s", task.Status, taskID)
			}
		}
	}
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

	//if ptcCluster.Spec.Network.Subnet != "" {
	//	params.Subnet = &ptcCluster.Spec.Network.Subnet
	//}
	logger.Info("subnet=", ptcCluster.Spec.Network.Subnet)
	params.Subnet = util.Ptr("255.255.255.0")
	if ptcMachine.Spec.BootDiskSize > 0 {
		diskSize := int64(ptcMachine.Spec.BootDiskSize)
		params.BootDiskSize = &diskSize
	}

	params.IPAddress = &allocatedIP
	params.UserData = &userData

	// 3. Trigger VM creation call
	resp, err := apiClient.Operations.CreateVM(params, auth.NewAPIKeyAuthWriter(auth.NewTokenManager().LoadTokenMust()))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to call CreateVM API: %w", err)
	}

	// 4. Parse Task payload
	task := &tasktypes.Task{}
	if err := opt.Rest2Task(resp.GetPayload(), task); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to parse task response: %w", err)
	}

	logger.Info("VM creation task initiated", "taskID", task.ID, "status", task.Status)

	// 5. Poll Task until completed
	completedTask, err := r.waitForTaskCompletion(ctx, apiClient, string(task.ID))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error waiting for VM creation task: %w", err)
	}

	// Extract generated Instance ID from Task Output JSON payload
	var taskOutput struct {
		InstanceID string `json:"instance-id"`
	}
	if len(completedTask.Output) > 0 {
		_ = json.Unmarshal(completedTask.Output, &taskOutput)
	}

	instanceID := taskOutput.InstanceID
	if instanceID == "" {
		instanceID = ptcMachine.Name // Fallback if API returns empty output
	}

	// 6. Update PTCMachine Status
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
