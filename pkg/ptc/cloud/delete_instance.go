package cloud

import (
	"context"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
)

func (c *Client) DeleteInstance(ctx context.Context, instanceID string) error {
	authorizer, err := c.GetAuthorizer(ctx)
	if err != nil {
		return err
	}

	// 1. Map Machine Spec to PTC API CreateVMParams
	params := operations.NewDeleteVMParams().WithContext(ctx)
	params.InstanceID = instanceID

	// 2. Trigger VM creation call
	resp, err := c.apiClient.Operations.DeleteVM(params, authorizer)
	if err != nil {
		return err
	}

	if _, err := c.waitForTaskComplete(ctx, resp.GetPayload()); err != nil {
		return err
	}

	return nil
}
