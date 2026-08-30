package cloud

import (
	"context"
	"fmt"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
)

func (c *Client) DeleteInstance(ctx context.Context, instanceID string) error {
	authorizer, err := c.GetAuthorizer(ctx)
	if err != nil {
		return err
	}

	// 1. Map Machine Spec to PTC API CreateVMParams
	params := operations.NewDeleteVMParams()
	params.InstanceID = instanceID

	// 2. Trigger VM creation call
	resp, err := c.apiClient.Operations.DeleteVM(params, authorizer)
	if err != nil {
		return err
	}

	task, err := toTask(resp.GetPayload())
	if err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	err = task.Wait(ctx, func(target *Task) error {
		freshTask, err := c.DescribeTask(ctx, task.ID.String())
		if err != nil {
			return err
		}
		*target = *freshTask
		return nil
	})
	if err != nil {
		return fmt.Errorf("error waiting for VM deletion task: %w", err)
	}

	return nil
}
