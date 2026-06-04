package pve

import (
	"context"
	"errors"

	"github.com/lz-wang/pvectl/internal/output"
)

var ErrNotFound = errors.New("not found")

type Backend interface {
	Nodes(ctx context.Context) ([]output.NodeRow, error)
	VMs(ctx context.Context, node string) ([]output.GuestRow, error)
	VM(ctx context.Context, node string, vmid int) (Guest, error)
	LXCs(ctx context.Context, node string) ([]output.GuestRow, error)
	LXC(ctx context.Context, node string, vmid int) (Guest, error)
}

type Guest interface {
	Row() output.GuestRow
	Start(ctx context.Context) (Task, error)
	Shutdown(ctx context.Context) (Task, error)
	Stop(ctx context.Context) (Task, error)
}

type Task interface {
	UPID() string
	WaitFor(ctx context.Context, seconds int) error
	ExitStatus() string
	Failed() bool
}
