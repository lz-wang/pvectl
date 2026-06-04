package pve

import (
	"context"

	proxmox "github.com/luthermonson/go-proxmox"

	"github.com/lz-wang/pvectl/internal/output"
)

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
		rows = append(rows, nodeRow(node))
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
