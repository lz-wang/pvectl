package pve

import (
	"context"

	proxmox "github.com/luthermonson/go-proxmox"
)

type proxmoxTask struct {
	task *proxmox.Task
}

func wrapTask(task *proxmox.Task) Task {
	if task == nil {
		return nil
	}
	return proxmoxTask{task: task}
}

func (t proxmoxTask) UPID() string {
	return string(t.task.UPID)
}

func (t proxmoxTask) WaitFor(ctx context.Context, seconds int) error {
	return t.task.WaitFor(ctx, seconds)
}

func (t proxmoxTask) ExitStatus() string {
	return t.task.ExitStatus
}

func (t proxmoxTask) Failed() bool {
	return t.task.IsFailed
}
