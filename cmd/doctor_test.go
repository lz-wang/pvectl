package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lz-wang/pvectl/internal/config"
	"github.com/lz-wang/pvectl/internal/output"
)

func TestConfigInitCommandWritesContext(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")

	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"config", "init",
		"--endpoint", "https://pve.example:8006/api2/json",
		"--token-id", "root@pam!test",
		"--token-secret-env", "PVECTL_TOKEN",
		"--insecure",
	}, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.CurrentContext != "home" {
		t.Fatalf("current context = %q, want home", cfg.CurrentContext)
	}
	ctx := cfg.Contexts["home"]
	if ctx.Endpoint != "https://pve.example:8006/api2/json" || ctx.TokenSecretEnv != "PVECTL_TOKEN" {
		t.Fatalf("context = %#v", ctx)
	}
	if !ctx.InsecureSkipVerify || ctx.Timeout != "30s" || ctx.DefaultOutput != "table" {
		t.Fatalf("context defaults = %#v", ctx)
	}
}

func TestConfigInitNoUseLeavesCurrentEmpty(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")

	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"config", "init",
		"--endpoint", "https://pve.example:8006/api2/json",
		"--token-id", "root@pam!test",
		"--token-secret-env", "PVECTL_TOKEN",
		"--no-use",
	}, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.CurrentContext != "" {
		t.Fatalf("current context = %q, want empty", cfg.CurrentContext)
	}
	if _, ok := cfg.Contexts["home"]; !ok {
		t.Fatal("expected home context")
	}
}

func TestConfigInitOverwriteBehavior(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	baseArgs := []string{
		"pvectl",
		"--config", cfgPath,
		"config", "init",
		"--endpoint", "https://pve.example:8006/api2/json",
		"--token-id", "root@pam!test",
		"--token-secret-env", "PVECTL_TOKEN",
	}
	if err := RunWithDependencies(baseArgs, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err != nil {
		t.Fatalf("run initial init: %v", err)
	}

	replacementArgs := []string{
		"pvectl",
		"--config", cfgPath,
		"config", "init",
		"--endpoint", "https://other.example:8006/api2/json",
		"--token-id", "root@pam!test",
		"--token-secret-env", "PVECTL_TOKEN",
	}
	if err := RunWithDependencies(replacementArgs, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err == nil {
		t.Fatal("expected duplicate context error")
	}

	replacementArgs = append(replacementArgs, "--overwrite")
	if err := RunWithDependencies(replacementArgs, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err != nil {
		t.Fatalf("run overwrite: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := cfg.Contexts["home"].Endpoint; got != "https://other.example:8006/api2/json" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestConfigInitRequiredFlagFailure(t *testing.T) {
	err := RunWithDependencies([]string{
		"pvectl",
		"--config", filepath.Join(t.TempDir(), "config.yaml"),
		"config", "init",
	}, "test", Dependencies{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err == nil {
		t.Fatal("expected required flag error")
	}
}

func TestDoctorOfflineCommand(t *testing.T) {
	cfgPath := writeTestConfig(t, "table")
	var stdout bytes.Buffer

	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"doctor", "--offline",
	}, "test", Dependencies{Stdout: &stdout, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "CONFIG_PATH") || !strings.Contains(got, "API_CONNECTIVITY") || !strings.Contains(got, "skip") {
		t.Fatalf("stdout = %s", got)
	}
}

func TestDoctorOfflineJSONCommand(t *testing.T) {
	cfgPath := writeTestConfig(t, "table")
	var stdout bytes.Buffer

	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"doctor", "--offline", "-o", "json",
	}, "test", Dependencies{Stdout: &stdout, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, `"check": "CONFIG_PATH"`) || !strings.Contains(got, `"status": "skip"`) {
		t.Fatalf("stdout = %s", got)
	}
}

func TestDoctorNodeCommand(t *testing.T) {
	cfgPath := writeTestConfig(t, "table")
	backend := &commandBackend{nodes: []output.NodeRow{{Name: "pve1"}}}
	var stdout bytes.Buffer

	err := RunWithDependencies([]string{
		"pvectl",
		"--config", cfgPath,
		"doctor", "--node", "pve1",
	}, "test", testDeps(&stdout, backend))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "NODE") || !strings.Contains(got, "node pve1 exists") {
		t.Fatalf("stdout = %s", got)
	}
}

func TestDoctorFailureReturnsError(t *testing.T) {
	var stdout bytes.Buffer
	err := RunWithDependencies([]string{
		"pvectl",
		"--config", filepath.Join(t.TempDir(), "missing.yaml"),
		"doctor", "--offline",
	}, "test", Dependencies{Stdout: &stdout, Stderr: &bytes.Buffer{}})
	if err == nil {
		t.Fatal("expected doctor failure")
	}
	if err.Error() != "doctor checks failed" {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(stdout.String(), "CONFIG_FILE") {
		t.Fatalf("stdout = %s", stdout.String())
	}
}
