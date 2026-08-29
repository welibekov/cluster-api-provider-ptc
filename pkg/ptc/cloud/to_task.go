package cloud

import (
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/opt"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/task/types"
)

func ToTask(input any) (*types.Task, error) {
	task := types.Task{}
	if err := opt.Rest2Task(input, &task); err != nil {
		return nil, err
	}

	return &task, nil
}
