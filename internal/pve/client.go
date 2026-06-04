package pve

import (
	"context"
	"fmt"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"

	"github.com/lz-wang/pvectl/internal/config"
	"github.com/lz-wang/pvectl/internal/output"
)

type ClientOptions struct {
	TokenSecret string
	Timeout     time.Duration
	Insecure    bool
}

func NewProxmoxBackend(ctxCfg config.Context, opts ClientOptions) (Backend, error) {
	if ctxCfg.Endpoint == "" {
		return nil, fmt.Errorf("pve endpoint is empty")
	}
	if ctxCfg.TokenID == "" {
		return nil, fmt.Errorf("pve token_id is empty")
	}
	if opts.TokenSecret == "" {
		return nil, fmt.Errorf("pve token secret is empty")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	clientOpts := []proxmox.Option{
		proxmox.WithAPIToken(ctxCfg.TokenID, opts.TokenSecret),
		proxmox.WithTimeout(opts.Timeout),
		proxmox.WithRetry(),
	}
	if opts.Insecure || ctxCfg.InsecureSkipVerify {
		clientOpts = append(clientOpts, proxmox.WithInsecureSkipVerify())
	}

	return &ProxmoxBackend{client: proxmox.NewClient(ctxCfg.Endpoint, clientOpts...)}, nil
}

type ProxmoxBackend struct {
	client *proxmox.Client
}

func (b *ProxmoxBackend) Nodes(ctx context.Context) ([]output.NodeRow, error) {
	nodes, err := b.client.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]output.NodeRow, 0, len(nodes))
	for _, node := range nodes {
		name := node.Node
		if name == "" {
			name = node.Name
		}
		status := node.Status
		if status == "" && node.Online == 1 {
			status = "online"
		}
		rows = append(rows, output.NodeRow{
			Name:    name,
			Status:  status,
			CPU:     node.CPU,
			Mem:     node.Mem,
			MaxMem:  node.MaxMem,
			Disk:    node.Disk,
			MaxDisk: node.MaxDisk,
			Uptime:  node.Uptime,
		})
	}
	return rows, nil
}

func (b *ProxmoxBackend) VMs(ctx context.Context, nodeName string) ([]output.GuestRow, error) {
	node, err := b.client.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	vms, err := node.VirtualMachines(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]output.GuestRow, 0, len(vms))
	for _, vm := range vms {
		rows = append(rows, vmRow(vm))
	}
	return rows, nil
}

func (b *ProxmoxBackend) VM(ctx context.Context, nodeName string, vmid int) (Guest, error) {
	node, err := b.client.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return nil, err
	}
	return vmGuest{vm: vm}, nil
}

func (b *ProxmoxBackend) LXCs(ctx context.Context, nodeName string) ([]output.GuestRow, error) {
	node, err := b.client.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	containers, err := node.Containers(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]output.GuestRow, 0, len(containers))
	for _, ct := range containers {
		rows = append(rows, lxcRow(ct))
	}
	return rows, nil
}

func (b *ProxmoxBackend) LXC(ctx context.Context, nodeName string, vmid int) (Guest, error) {
	node, err := b.client.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	ct, err := node.Container(ctx, vmid)
	if err != nil {
		return nil, err
	}
	return lxcGuest{ct: ct}, nil
}

type vmGuest struct {
	vm *proxmox.VirtualMachine
}

func (g vmGuest) Row() output.GuestRow {
	return vmRow(g.vm)
}

func (g vmGuest) Start(ctx context.Context) (Task, error) {
	task, err := g.vm.Start(ctx)
	return wrapTask(task), err
}

func (g vmGuest) Shutdown(ctx context.Context) (Task, error) {
	task, err := g.vm.Shutdown(ctx)
	return wrapTask(task), err
}

func (g vmGuest) Stop(ctx context.Context) (Task, error) {
	task, err := g.vm.Stop(ctx)
	return wrapTask(task), err
}

type lxcGuest struct {
	ct *proxmox.Container
}

func (g lxcGuest) Row() output.GuestRow {
	return lxcRow(g.ct)
}

func (g lxcGuest) Start(ctx context.Context) (Task, error) {
	task, err := g.ct.Start(ctx)
	return wrapTask(task), err
}

func (g lxcGuest) Shutdown(ctx context.Context) (Task, error) {
	task, err := g.ct.Shutdown(ctx, false, 0)
	return wrapTask(task), err
}

func (g lxcGuest) Stop(ctx context.Context) (Task, error) {
	task, err := g.ct.Stop(ctx)
	return wrapTask(task), err
}

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

func vmRow(vm *proxmox.VirtualMachine) output.GuestRow {
	if vm == nil {
		return output.GuestRow{Kind: "vm"}
	}
	return output.GuestRow{
		Kind:    "vm",
		VMID:    uint64(vm.VMID),
		Name:    vm.Name,
		Node:    vm.Node,
		Status:  vm.Status,
		CPUs:    vm.CPUs,
		CPU:     vm.CPU,
		Mem:     vm.Mem,
		MaxMem:  vm.MaxMem,
		MaxDisk: vm.MaxDisk,
		Uptime:  vm.Uptime,
		Tags:    vm.Tags,
	}
}

func lxcRow(ct *proxmox.Container) output.GuestRow {
	if ct == nil {
		return output.GuestRow{Kind: "lxc"}
	}
	return output.GuestRow{
		Kind:    "lxc",
		VMID:    uint64(ct.VMID),
		Name:    ct.Name,
		Node:    ct.Node,
		Status:  ct.Status,
		CPUs:    ct.CPUs,
		Mem:     0,
		MaxMem:  ct.MaxMem,
		MaxDisk: ct.MaxDisk,
		Uptime:  ct.Uptime,
		Tags:    ct.Tags,
	}
}
