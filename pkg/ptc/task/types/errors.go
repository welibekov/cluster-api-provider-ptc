package types

import "errors"

var (
	ErrQueueEmpty   = errors.New("queue is empty")
	ErrQueueFull    = errors.New("queue is full")
	ErrTaskNotFound = errors.New("task not found")
)
