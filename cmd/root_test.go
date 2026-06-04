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

func TestParseSetFlags(t *testing.T) {
	values, err := parseSetFlags([]string{"memory=4096", "cores=4"})
	if err != nil {
		t.Fatalf("parse set: %v", err)
	}
	if values["memory"] != "4096" || values["cores"] != "4" {
		t.Fatalf("values = %#v", values)
	}

	for _, values := range [][]string{nil, []string{"memory"}, []string{"=4096"}} {
		if _, err := parseSetFlags(values); err == nil {
			t.Fatalf("expected error for %#v", values)
		}
	}
}

func TestVMListCommandUsesDefaultOutputFromContext(t *testing.T) {
	cfgPath := writeTestConfig(t, "json")

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

func TestVMCloneCommandPrintsNewVMIDAndWaits(t *testing.T) {
	cfgPath := writeTestConfig(t, "json")
	task := &commandTask{upid: "UPID:pve1:clone"}
	guest := &commandGuest{
		row:     output.GuestRow{Kind: "vm", VMID: 9000, Node: "pve1"},
		task:    task,
		cloneID: 101,
	}
	backend := &commandBackend{
		nodes:    []output.NodeRow{{Name: "pve1"}},
		vmGuests: map[string]map[int]*commandGuest{"pve1": {9000: guest}},
	}
	var stdout bytes.Buffer

	err := RunWithDependencies([]string{
		"pvectl", "--config", cfgPath,
		"vm", "clone", "9000",
		"--newid", "101",
		"--name", "app-vm",
		"--target", "pve2",
		"--wait",
	}, "test", testDeps(&stdout, backend))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout.String(), `"new_vmid": 101`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if !task.waited {
		t.Fatal("expected clone task to be waited")
	}
	if guest.cloneOptions.Name != "app-vm" || guest.cloneOptions.Target != "pve2" || guest.cloneOptions.NewID != 101 {
		t.Fatalf("clone options = %#v", guest.cloneOptions)
	}
}

func TestLXCCloneCommandMapsHostname(t *testing.T) {
	cfgPath := writeTestConfig(t, "json")
	task := &commandTask{upid: "UPID:pve1:lxcclone"}
	guest := &commandGuest{
		row:     output.GuestRow{Kind: "lxc", VMID: 900, Node: "pve1"},
		task:    task,
		cloneID: 201,
	}
	backend := &commandBackend{
		nodes:     []output.NodeRow{{Name: "pve1"}},
		lxcGuests: map[string]map[int]*commandGuest{"pve1": {900: guest}},
	}
	var stdout bytes.Buffer

	err := RunWithDependencies([]string{
		"pvectl", "--config", cfgPath,
		"lxc", "clone", "900",
		"--newid", "201",
		"--hostname", "app-lxc",
		"--target", "pve1",
		"--wait",
	}, "test", testDeps(&stdout, backend))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout.String(), `"new_vmid": 201`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if !task.waited {
		t.Fatal("expected clone task to be waited")
	}
	if guest.cloneOptions.Hostname != "app-lxc" || guest.cloneOptions.Target != "pve1" {
		t.Fatalf("clone options = %#v", guest.cloneOptions)
	}
}

func TestVMConfigCommandPassesSetValuesAndWaits(t *testing.T) {
	cfgPath := writeTestConfig(t, "table")
	task := &commandTask{upid: "UPID:pve1:config"}
	guest := &commandGuest{
		row:  output.GuestRow{Kind: "vm", VMID: 101, Node: "pve1"},
		task: task,
	}
	backend := &commandBackend{
		nodes:    []output.NodeRow{{Name: "pve1"}},
		vmGuests: map[string]map[int]*commandGuest{"pve1": {101: guest}},
	}

	err := RunWithDependencies([]string{
		"pvectl", "--config", cfgPath,
		"vm", "config", "101",
		"--set", "memory=4096",
		"--set", "cores=4",
		"--wait",
	}, "test", testDeps(&bytes.Buffer{}, backend))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !task.waited {
		t.Fatal("expected config task to be waited")
	}
	if guest.configValues["memory"] != "4096" || guest.configValues["cores"] != "4" {
		t.Fatalf("config values = %#v", guest.configValues)
	}
}

func TestLXCConfigCommandPassesSetValues(t *testing.T) {
	cfgPath := writeTestConfig(t, "table")
	task := &commandTask{upid: "UPID:pve1:lxcconfig"}
	guest := &commandGuest{
		row:  output.GuestRow{Kind: "lxc", VMID: 201, Node: "pve1"},
		task: task,
	}
	backend := &commandBackend{
		nodes:     []output.NodeRow{{Name: "pve1"}},
		lxcGuests: map[string]map[int]*commandGuest{"pve1": {201: guest}},
	}

	err := RunWithDependencies([]string{
		"pvectl", "--config", cfgPath,
		"lxc", "config", "201",
		"--set", "memory=2048",
		"--set", "cores=2",
		"--wait",
	}, "test", testDeps(&bytes.Buffer{}, backend))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !task.waited {
		t.Fatal("expected config task to be waited")
	}
	if guest.configValues["memory"] != "2048" || guest.configValues["cores"] != "2" {
		t.Fatalf("config values = %#v", guest.configValues)
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
	nodes     []output.NodeRow
	vms       map[string][]output.GuestRow
	lxcs      map[string][]output.GuestRow
	vmGuests  map[string]map[int]*commandGuest
	lxcGuests map[string]map[int]*commandGuest
}

func (b *commandBackend) Nodes(context.Context) ([]output.NodeRow, error) {
	return b.nodes, nil
}

func (b *commandBackend) VMs(_ context.Context, node string) ([]output.GuestRow, error) {
	return b.vms[node], nil
}

func (b *commandBackend) VM(_ context.Context, node string, vmid int) (pve.Guest, error) {
	if guest := b.vmGuests[node][vmid]; guest != nil {
		return guest, nil
	}
	return nil, pve.ErrNotFound
}

func (b *commandBackend) LXCs(_ context.Context, node string) ([]output.GuestRow, error) {
	return b.lxcs[node], nil
}

func (b *commandBackend) LXC(_ context.Context, node string, vmid int) (pve.Guest, error) {
	if guest := b.lxcGuests[node][vmid]; guest != nil {
		return guest, nil
	}
	return nil, pve.ErrNotFound
}

type commandGuest struct {
	row          output.GuestRow
	task         pve.Task
	cloneID      int
	cloneOptions pve.CloneOptions
	configValues map[string]string
}

func (g *commandGuest) Row() output.GuestRow {
	return g.row
}

func (g *commandGuest) Start(context.Context) (pve.Task, error) {
	return g.task, nil
}

func (g *commandGuest) Shutdown(context.Context) (pve.Task, error) {
	return g.task, nil
}

func (g *commandGuest) Stop(context.Context) (pve.Task, error) {
	return g.task, nil
}

func (g *commandGuest) Clone(_ context.Context, options pve.CloneOptions) (pve.CloneResult, pve.Task, error) {
	g.cloneOptions = options
	name := options.Name
	if name == "" {
		name = options.Hostname
	}
	return pve.CloneResult{
		Kind:       g.row.Kind,
		SourceVMID: g.row.VMID,
		NewVMID:    uint64(g.cloneID),
		SourceNode: g.row.Node,
		TargetNode: options.Target,
		Name:       name,
		Task:       g.task.UPID(),
	}, g.task, nil
}

func (g *commandGuest) Config(_ context.Context, values map[string]string) (pve.Task, error) {
	g.configValues = values
	return g.task, nil
}

type commandTask struct {
	upid   string
	waited bool
}

func (t *commandTask) UPID() string {
	return t.upid
}

func (t *commandTask) WaitFor(context.Context, int) error {
	t.waited = true
	return nil
}

func (t *commandTask) ExitStatus() string {
	return ""
}

func (t *commandTask) Failed() bool {
	return false
}

func writeTestConfig(t *testing.T, defaultOutput string) string {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PVECTL_TOKEN", "secret")
	cfg := config.Empty()
	if err := cfg.SetContext("home", config.Context{
		Endpoint:       "https://pve.example:8006/api2/json",
		TokenID:        "root@pam!test",
		TokenSecretEnv: "PVECTL_TOKEN",
		DefaultOutput:  defaultOutput,
	}); err != nil {
		t.Fatalf("set context: %v", err)
	}
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	return cfgPath
}

func testDeps(stdout *bytes.Buffer, backend pve.Backend) Dependencies {
	return Dependencies{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		BackendFactory: func(config.Context, pve.ClientOptions) (pve.Backend, error) {
			return backend, nil
		},
	}
}
