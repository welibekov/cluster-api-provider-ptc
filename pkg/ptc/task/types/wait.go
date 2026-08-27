package types

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
)

// Wait blocks until the asynchronous task completes or fails, polling the task state periodically.
//
// It uses the provided describeTaskFn to fetch the current state of the task
// at regular intervals (currently set to every second). The function returns
// nil if the task completes successfully, or an error if the task fails, the
// context is canceled, or an error occurs while fetching the task state.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - describeTaskFn: A function that retrieves the current state of the task.
//
// Returns:
//   - error: Returns nil on successful task completion, or an error if the task fails,
//     the context is canceled, or an error occurs during task state retrieval.
func (t *Task) Wait(ctx context.Context, describTaskFn func(*Task) error) error {
	// TODO: The ticker interval can be taken from a config parameter.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := describTaskFn(t); err != nil {
				return err
			}

			if t.Done() {
				return nil
			}

			if t.Failed() {
				return fmt.Errorf("failed: %v", util.Deref(t.ErrMessage))
			}
		}
	}
}

func (t Task) String() string {
	taskBytes, _ := json.MarshalIndent(t, "", "  ")

	return string(taskBytes)
}
