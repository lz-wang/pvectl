package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lz-wang/pvectl/internal/config"
	"github.com/lz-wang/pvectl/internal/output"
	"github.com/lz-wang/pvectl/internal/pve"
)

func TestParseVMID(t *testing.T) {
	if vmid, err := parseVMID("100"); err != nil || vmid != 100 {
		t.Fatalf("parse valid = %d/%v", vmid, err)
	}
	for _, value := range []string{"", "abc", "0", "-1"} {
		if _, err := parseVMID(value); err == nil {
			t.Fatalf("expected invalid vmid for %q", value)
		}
	}
}

func TestVMListCommandUsesDefaultOutputFromContext(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	t.Setenv("PVECTL_TOKEN", "secret")
	cfg := config.Empty()
	if err := cfg.SetContext("home", config.Context{
		Endpoint:       "https://pve.example:8006/api2/json",
		TokenID:        "root@pam!test",
		TokenSecretEnv: "PVECTL_TOKEN",
		DefaultOutput:  "json",
	}); err != nil {
		t.Fatalf("set context: %v", err)
	}
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var stdout bytes.Buffer
	backend := &commandBackend{
		nodes: []output.NodeRow{{Name: "pve1"}},
		vms:   map[string][]output.GuestRow{"pve1": {{Kind: "vm", VMID: 100, Name: "debian", Node: "pve1"}}},
	}
	deps := Dependencies{
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
		BackendFactory: func(config.Context, pve.ClientOptions) (pve.Backend, error) {
			return backend, nil
		},
	}

	err := RunWithDependencies([]string{"pvectl", "--config", cfgPath, "vm", "ls"}, "test", deps)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout.String(), `"vmid": 100`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestConfigSetContextCommandDoesNotRequireSecretEnv(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"config", "set-context", "home",
		"--endpoint", "https://pve.example:8006/api2/json",
		"--token-id", "root@pam!test",
		"--token-secret-env", "PVECTL_TOKEN",
		"--default-output", "yaml",
	}, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "token_secret:") || !strings.Contains(got, "token_secret_env: PVECTL_TOKEN") {
		t.Fatalf("config content = %s", got)
	}
}

type commandBackend struct {
	nodes []output.NodeRow
	vms   map[string][]output.GuestRow
	lxcs  map[string][]output.GuestRow
}

func (b *commandBackend) Nodes(context.Context) ([]output.NodeRow, error) {
	return b.nodes, nil
}

func (b *commandBackend) VMs(_ context.Context, node string) ([]output.GuestRow, error) {
	return b.vms[node], nil
}

func (b *commandBackend) VM(context.Context, string, int) (pve.Guest, error) {
	return nil, pve.ErrNotFound
}

func (b *commandBackend) LXCs(_ context.Context, node string) ([]output.GuestRow, error) {
	return b.lxcs[node], nil
}

func (b *commandBackend) LXC(context.Context, string, int) (pve.Guest, error) {
	return nil, pve.ErrNotFound
}
