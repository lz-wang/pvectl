package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndUseContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Empty()

	if err := cfg.SetContext("home", Context{
		Endpoint:       "https://pve.lan:8006/api2/json",
		TokenID:        "automation@pve!pvectl",
		TokenSecretEnv: "PVECTL_HOME_TOKEN_SECRET",
		Timeout:        "30s",
		DefaultOutput:  "json",
	}); err != nil {
		t.Fatalf("set context: %v", err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CurrentContext != "home" {
		t.Fatalf("current context = %q, want home", loaded.CurrentContext)
	}

	name, ctx, err := loaded.SelectContext("")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if name != "home" || ctx.TokenID != "automation@pve!pvectl" {
		t.Fatalf("selected %q/%q", name, ctx.TokenID)
	}

	if err := loaded.UseContext("missing"); err == nil {
		t.Fatal("expected missing context error")
	}
}

func TestLoadOrEmptyMissingFile(t *testing.T) {
	cfg, err := LoadOrEmpty(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("load or empty: %v", err)
	}
	if cfg == nil || len(cfg.Contexts) != 0 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestResolveTokenSecret(t *testing.T) {
	t.Setenv("PVECTL_TEST_TOKEN", "secret")

	secret, err := ResolveTokenSecret(Context{TokenSecretEnv: "PVECTL_TEST_TOKEN"})
	if err != nil {
		t.Fatalf("resolve secret: %v", err)
	}
	if secret != "secret" {
		t.Fatalf("secret = %q", secret)
	}

	os.Unsetenv("PVECTL_TEST_TOKEN")
	if _, err := ResolveTokenSecret(Context{TokenSecretEnv: "PVECTL_TEST_TOKEN"}); err == nil {
		t.Fatal("expected missing env error")
	}
}
