package pve

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/lz-wang/pvectl/internal/output"
)

func TestGuestServiceResolveExplicitNode(t *testing.T) {
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}, {Name: "pve2"}},
		vms: map[string]map[int]*fakeGuest{
			"pve2": {100: {row: output.GuestRow{Kind: "vm", VMID: 100, Node: "pve2"}}},
		},
	}
	svc := NewVMService(backend, TaskRunner{}, nil, false)

	row, err := svc.Get(context.Background(), 100, "pve2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if row.Node != "pve2" {
		t.Fatalf("node = %q", row.Node)
	}
	if backend.vmCalls != 1 {
		t.Fatalf("vm calls = %d", backend.vmCalls)
	}
}

func TestGuestServiceResolveByTraversal(t *testing.T) {
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}, {Name: "pve2"}},
		vms: map[string]map[int]*fakeGuest{
			"pve2": {100: {row: output.GuestRow{Kind: "vm", VMID: 100, Node: "pve2"}}},
		},
	}
	svc := NewVMService(backend, TaskRunner{}, nil, false)

	row, err := svc.Get(context.Background(), 100, "")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if row.Node != "pve2" {
		t.Fatalf("node = %q", row.Node)
	}
	if backend.vmCalls != 2 {
		t.Fatalf("vm calls = %d", backend.vmCalls)
	}
}

func TestGuestServiceListPartialFailure(t *testing.T) {
	backend := &fakeBackend{
		nodes:   []output.NodeRow{{Name: "pve1"}, {Name: "pve2"}},
		vmErrs:  map[string]error{"pve1": errors.New("forbidden")},
		vmRows:  map[string][]output.GuestRow{"pve2": {{Kind: "vm", VMID: 100, Node: "pve2"}}},
		vms:     map[string]map[int]*fakeGuest{},
		lxcs:    map[string]map[int]*fakeGuest{},
		lxcRows: map[string][]output.GuestRow{},
	}
	svc := NewVMService(backend, TaskRunner{}, nil, false)

	rows, err := svc.List(context.Background(), "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].Node != "pve2" {
		t.Fatalf("rows = %#v", rows)
	}
}

func TestTaskRunnerWaitsAndReportsFailure(t *testing.T) {
	var stderr bytes.Buffer
	task := &fakeTask{upid: "UPID:pve1:1", failed: true, exitStatus: "ERROR"}
	runner := TaskRunner{Wait: true, ErrWriter: &stderr}

	err := runner.Handle(context.Background(), task)
	if err == nil {
		t.Fatal("expected task failure")
	}
	if !task.waited {
		t.Fatal("expected task to be waited")
	}
	if stderr.String() == "" {
		t.Fatal("expected task output on stderr")
	}
}

func TestGuestServiceCloneReturnsNewIDAndWaits(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:clone"}
	guest := &fakeGuest{
		row:       output.GuestRow{Kind: "vm", VMID: 9000, Node: "pve1", Name: "tmpl"},
		task:      task,
		cloneID:   101,
		cloneName: "app-vm",
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {9000: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	result, err := svc.Clone(context.Background(), 9000, "", CloneOptions{Name: "app-vm", Target: "pve2"})
	if err != nil {
		t.Fatalf("clone: %v", err)
	}
	if result.NewVMID != 101 || result.Task != "UPID:pve1:clone" {
		t.Fatalf("result = %#v", result)
	}
	if !task.waited {
		t.Fatal("expected clone task to be waited")
	}
	if guest.cloneOptions.Target != "pve2" || guest.cloneOptions.Name != "app-vm" {
		t.Fatalf("clone options = %#v", guest.cloneOptions)
	}
}

func TestGuestServiceConfigPassesValuesAndWaits(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:config"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.Config(context.Background(), 101, "", map[string]string{"memory": "4096", "cores": "4"})
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if !task.waited {
		t.Fatal("expected config task to be waited")
	}
	if guest.configValues["memory"] != "4096" || guest.configValues["cores"] != "4" {
		t.Fatalf("config values = %#v", guest.configValues)
	}
}

type fakeBackend struct {
	nodes   []output.NodeRow
	vms     map[string]map[int]*fakeGuest
	lxcs    map[string]map[int]*fakeGuest
	vmRows  map[string][]output.GuestRow
	lxcRows map[string][]output.GuestRow
	vmErrs  map[string]error
	lxcErrs map[string]error
	vmCalls int
}

func (b *fakeBackend) Nodes(context.Context) ([]output.NodeRow, error) {
	return b.nodes, nil
}

func (b *fakeBackend) VMs(_ context.Context, node string) ([]output.GuestRow, error) {
	if err := b.vmErrs[node]; err != nil {
		return nil, err
	}
	return b.vmRows[node], nil
}

func (b *fakeBackend) VM(_ context.Context, node string, vmid int) (Guest, error) {
	b.vmCalls++
	if guest := b.vms[node][vmid]; guest != nil {
		return guest, nil
	}
	return nil, ErrNotFound
}

func (b *fakeBackend) LXCs(_ context.Context, node string) ([]output.GuestRow, error) {
	if err := b.lxcErrs[node]; err != nil {
		return nil, err
	}
	return b.lxcRows[node], nil
}

func (b *fakeBackend) LXC(_ context.Context, node string, vmid int) (Guest, error) {
	if guest := b.lxcs[node][vmid]; guest != nil {
		return guest, nil
	}
	return nil, ErrNotFound
}

type fakeGuest struct {
	row          output.GuestRow
	task         Task
	cloneID      int
	cloneName    string
	cloneOptions CloneOptions
	configValues map[string]string
}

func (g *fakeGuest) Row() output.GuestRow {
	return g.row
}

func (g *fakeGuest) Start(context.Context) (Task, error) {
	return g.task, nil
}

func (g *fakeGuest) Shutdown(context.Context) (Task, error) {
	return g.task, nil
}

func (g *fakeGuest) Stop(context.Context) (Task, error) {
	return g.task, nil
}

func (g *fakeGuest) Clone(_ context.Context, options CloneOptions) (CloneResult, Task, error) {
	g.cloneOptions = options
	name := options.Name
	if name == "" {
		name = options.Hostname
	}
	if name == "" {
		name = g.cloneName
	}
	return CloneResult{
		Kind:       g.row.Kind,
		SourceVMID: g.row.VMID,
		NewVMID:    uint64(g.cloneID),
		SourceNode: g.row.Node,
		TargetNode: options.Target,
		Name:       name,
		Task:       g.task.UPID(),
	}, g.task, nil
}

func (g *fakeGuest) Config(_ context.Context, values map[string]string) (Task, error) {
	g.configValues = values
	return g.task, nil
}

type fakeTask struct {
	upid       string
	waited     bool
	failed     bool
	exitStatus string
}

func (t *fakeTask) UPID() string {
	return t.upid
}

func (t *fakeTask) WaitFor(context.Context, int) error {
	t.waited = true
	return nil
}

func (t *fakeTask) ExitStatus() string {
	return t.exitStatus
}

func (t *fakeTask) Failed() bool {
	return t.failed
}
