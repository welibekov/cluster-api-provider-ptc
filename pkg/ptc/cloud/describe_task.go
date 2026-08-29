package cloud

import (
	"context"
	"fmt"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/task/types"
)

func DescribeTask(ctx context.Context, taskID string, ptcclient *client.Ptc) (*types.Task, error) {
	params := operations.NewDescribeTaskParams().WithContext(ctx)
	params.TaskID = taskID

	resp, err := ptcclient.Operations.DescribeTask(params, auth.NewAPIKeyAuthWriter(auth.NewTokenManager().LoadTokenMust()))
	if err != nil {
		return nil, fmt.Errorf("error calling API with auth: %w", err)
	}

	task, err := ToTask(resp.GetPayload())
	if err != nil {
		return nil, err
	}

	return task, nil
}
