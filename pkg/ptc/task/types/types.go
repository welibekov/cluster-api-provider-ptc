package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/types"
	internalutil "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/util"
)

// Task represents a unit of work with associated metadata.
type Task struct {
	ID         TaskID          `json:"task-id"`               // Unique identifier for the task
	Title      string          `json:"title,omitempty"`       // Title or description of the task
	ErrMessage *string         `json:"err_message,omitempty"` // Optional error message if the task fails
	Status     TaskStatus      `json:"status"`                // Current status of the task
	CreatedAt  time.Time       `json:"created_at,omitempty"`  // Timestamp when the task was created
	UpdatedAt  time.Time       `json:"updated_at,omitempty"`  // Timestamp when the task was last updated
	Handler    *TaskHandler    `json:"handler,omitempty"`     // Parameters associated with the task
	Output     json.RawMessage `json:"output,omitempty"`      // Output result of handler execution
}

type TaskID string

// String converts TaskID to a string.
func (t TaskID) String() string {
	return string(t)
}

// CompactHandlerOutput removes unnecessary whitespace from the handler output and returns it as a string.
func (t *Task) CompactHandlerOutput() (string, error) {
	if t.Output == nil {
		return "", nil
	}

	dst := new(bytes.Buffer)
	// json.Compact removes all unnecessary whitespace
	err := json.Compact(dst, t.Output)
	if err != nil {
		return "", err
	}
	return dst.String(), nil
}

// SetStatusAccepted sets the task status to accepted.
func (t *Task) SetStatusAccepted() {
	t.Status = TaskAccepted
}

// SetStatusCompleted sets the task status to completed.
func (t *Task) SetStatusCompleted() {
	t.Status = TaskCompleted
}

// SetStatusRunning sets the task status to running.
func (t *Task) SetStatusRunning() {
	t.Status = TaskRunning
}

// SetStatusFailed sets the task status to failed and logs the error message.
func (t *Task) SetStatusFailed(err error) {
	t.Status = TaskFailed

	if err != nil {
		t.ErrMessage = internalutil.Ptr(err.Error())
	}
}

// Done returns true if the task status is completed.
func (t *Task) Done() bool {
	return t.Status == TaskCompleted
}

// Failed returns true if the task status is failed.
func (t *Task) Failed() bool {
	return t.Status == TaskFailed
}

// Execute runs the task's handler function with parameters and updates the status accordingly.
func (t *Task) Execute(fnExec func(params json.RawMessage, principal *types.Principal) (json.RawMessage, error)) error {
	if fnExec == nil {
		return fmt.Errorf("uknown handler %s", t.GetHandler())
	}

	t.SetStatusRunning()
	output, err := fnExec(t.GetHandlerParams(), t.GetHandlerPrincipal())
	if err == nil && output != nil {
		t.Output = output
	}
	return err
}

// GetHandler returns the name of the handler associated with the task.
func (t *Task) GetHandler() string {
	return t.Handler.Name
}

// GetHandlerParams retrieves the parameters for the handler.
func (t *Task) GetHandlerParams() json.RawMessage {
	return t.Handler.Params
}

// GetHandlerPrincipal retrieves the principal details for the task handler.
func (t *Task) GetHandlerPrincipal() *types.Principal {
	return t.Handler.Principal
}

// Marshal marshals the Task into a JSON format.
func (t *Task) Marshal() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// IsOwnedBy checks if the task is owned by the specified principal.
func (t *Task) IsOwnedBy(principal *types.Principal) bool {
	return t.GetHandlerPrincipal().PtcTenantID == principal.PtcTenantID
}

// TaskHandler holds parameters for the task handler.
type TaskHandler struct {
	Name      string           `json:"name"`      // Name of the handler for the task
	Params    json.RawMessage  `json:"params"`    // Parameters to pass to the handler
	Principal *types.Principal `json:"principal"` // Principal details for the task
}

// TaskStatus defines the possible states of a task.
type TaskStatus string

const (
	TaskAccepted  TaskStatus = "accepted"  // Task has been accepted for processing
	TaskCompleted TaskStatus = "completed" // Task has been successfully completed
	TaskFailed    TaskStatus = "failed"    // Task has failed during processing
	TaskRunning   TaskStatus = "running"   // Task is currently being processed
)

// GenTaskID generates a new unique task identifier.
func GenTaskID() TaskID {
	return TaskID("tsk-" + internalutil.GenRandomSuffix()) // Returns a unique task ID prefixed with "tsk-"
}

// GetEmptyTaskID returns an empty TaskID.
func GetEmptyTaskID() TaskID {
	return TaskID("") // Returns an empty TaskID
}
