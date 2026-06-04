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

func TestGuestServiceDeleteWaits(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:delete"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.Delete(context.Background(), 101, "")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !guest.deleted {
		t.Fatal("expected guest to be deleted")
	}
	if !task.waited {
		t.Fatal("expected delete task to be waited")
	}
}

func TestGuestServiceMigratePassesOptions(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:migrate"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.Migrate(context.Background(), 101, "", MigrateOptions{Target: "pve2", Online: true})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if guest.migrateOptions.Target != "pve2" || !guest.migrateOptions.Online {
		t.Fatalf("migrate options = %#v", guest.migrateOptions)
	}
}

func TestGuestServiceResizePassesDiskAndSize(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:resize"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.Resize(context.Background(), 101, "", "scsi0", "+20G")
	if err != nil {
		t.Fatalf("resize: %v", err)
	}
	if guest.resizeDisk != "scsi0" || guest.resizeSize != "+20G" {
		t.Fatalf("resize = %s/%s", guest.resizeDisk, guest.resizeSize)
	}
}

func TestGuestServiceOperationTaskFailure(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:delete", failed: true, exitStatus: "ERROR"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.Delete(context.Background(), 101, "")
	if err == nil {
		t.Fatal("expected task failure")
	}
	if got := err.Error(); got != "task UPID:pve1:delete failed: ERROR" {
		t.Fatalf("error = %q", got)
	}
}

func TestGuestServiceListSnapshots(t *testing.T) {
	guest := &fakeGuest{
		row: output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		snapshots: []output.SnapshotRow{
			{Kind: "vm", VMID: 101, Node: "pve1", Name: "before-upgrade"},
		},
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{}, nil, false)

	rows, err := svc.ListSnapshots(context.Background(), 101, "")
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "before-upgrade" {
		t.Fatalf("snapshots = %#v", rows)
	}
}

func TestGuestServiceCreateSnapshotWaits(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:snapshot"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.CreateSnapshot(context.Background(), 101, "", " before-upgrade ")
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if guest.createdSnapshot != "before-upgrade" {
		t.Fatalf("created snapshot = %q", guest.createdSnapshot)
	}
	if !task.waited {
		t.Fatal("expected create task to be waited")
	}
}

func TestGuestServiceRollbackSnapshotWaits(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:rollback"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.RollbackSnapshot(context.Background(), 101, "", "before-upgrade")
	if err != nil {
		t.Fatalf("rollback snapshot: %v", err)
	}
	if guest.rollbackSnapshot != "before-upgrade" {
		t.Fatalf("rollback snapshot = %q", guest.rollbackSnapshot)
	}
	if !task.waited {
		t.Fatal("expected rollback task to be waited")
	}
}

func TestGuestServiceSnapshotNameRequired(t *testing.T) {
	svc := NewVMService(&fakeBackend{}, TaskRunner{}, nil, false)

	if err := svc.CreateSnapshot(context.Background(), 101, "", " "); err == nil {
		t.Fatal("expected empty create snapshot name error")
	}
	if err := svc.RollbackSnapshot(context.Background(), 101, "", " "); err == nil {
		t.Fatal("expected empty rollback snapshot name error")
	}
}

func TestGuestServiceSnapshotTaskFailure(t *testing.T) {
	task := &fakeTask{upid: "UPID:pve1:rollback", failed: true, exitStatus: "ERROR"}
	guest := &fakeGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &fakeBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string]map[int]*fakeGuest{"pve1": {101: guest}},
	}
	svc := NewVMService(backend, TaskRunner{Wait: true}, nil, false)

	err := svc.RollbackSnapshot(context.Background(), 101, "", "before-upgrade")
	if err == nil {
		t.Fatal("expected task failure")
	}
	if got := err.Error(); got != "task UPID:pve1:rollback failed: ERROR" {
		t.Fatalf("error = %q", got)
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
	row              output.GuestRow
	task             Task
	cloneID          int
	cloneName        string
	cloneOptions     CloneOptions
	configValues     map[string]string
	deleted          bool
	migrateOptions   MigrateOptions
	resizeDisk       string
	resizeSize       string
	snapshots        []output.SnapshotRow
	createdSnapshot  string
	rollbackSnapshot string
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

func (g *fakeGuest) Delete(context.Context) (Task, error) {
	g.deleted = true
	return g.task, nil
}

func (g *fakeGuest) Migrate(_ context.Context, options MigrateOptions) (Task, error) {
	g.migrateOptions = options
	return g.task, nil
}

func (g *fakeGuest) Resize(_ context.Context, disk, size string) (Task, error) {
	g.resizeDisk = disk
	g.resizeSize = size
	return g.task, nil
}

func (g *fakeGuest) Snapshots(context.Context) ([]output.SnapshotRow, error) {
	return g.snapshots, nil
}

func (g *fakeGuest) CreateSnapshot(_ context.Context, name string) (Task, error) {
	g.createdSnapshot = name
	return g.task, nil
}

func (g *fakeGuest) RollbackSnapshot(_ context.Context, name string) (Task, error) {
	g.rollbackSnapshot = name
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
