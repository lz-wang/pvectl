package pve

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"
)

type TaskRunner struct {
	Wait        bool
	WaitTimeout time.Duration
	ErrWriter   io.Writer
}

func (r TaskRunner) Handle(ctx context.Context, task Task) error {
	if task == nil {
		return nil
	}

	if r.ErrWriter != nil {
		fmt.Fprintf(r.ErrWriter, "task: %s\n", task.UPID())
	}
	if !r.Wait {
		return nil
	}

	timeout := r.WaitTimeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	seconds := int(math.Ceil(timeout.Seconds()))
	if seconds <= 0 {
		seconds = 300
	}

	if r.ErrWriter != nil {
		fmt.Fprintf(r.ErrWriter, "waiting for task: %s\n", task.UPID())
	}
	if err := task.WaitFor(ctx, seconds); err != nil {
		return fmt.Errorf("task %s wait failed: %w", task.UPID(), err)
	}
	if task.Failed() {
		status := task.ExitStatus()
		if status == "" {
			status = "unknown"
		}
		return fmt.Errorf("task %s failed: %s", task.UPID(), status)
	}
	if r.ErrWriter != nil {
		fmt.Fprintf(r.ErrWriter, "task completed: %s\n", task.UPID())
	}
	return nil
}
