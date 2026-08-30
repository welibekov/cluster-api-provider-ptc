package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
)

func (c *Client) CreateInstance(ctx context.Context, params *operations.CreateVMParams) (*string, error) {
	authorizer, err := c.GetAuthorizer(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.Operations.CreateVM(params, authorizer)
	if err != nil {
		return nil, err
	}

	task, err := toTask(resp.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("failed to parse task response: %w", err)
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
		return nil, fmt.Errorf("error waiting for VM creation task: %w", err)
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
				return nil, fmt.Errorf("failed to unmarshal inner task output payload: %w", err)
			}
		} else {
			// Fallback: Attempt single-step unmarshal if task output is direct JSON
			if err := json.Unmarshal(task.Output, &taskOutput); err != nil {
				return nil, fmt.Errorf("failed to unmarshal single-step task if direct JSON: %w", err)
			}
		}
	}

	return util.Ptr(taskOutput.InstanceID), nil
}
