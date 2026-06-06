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

func TestInitContextCreatesAndSetsCurrent(t *testing.T) {
	cfg := Empty()

	if err := cfg.InitContext(InitOptions{
		Name:    "home",
		Context: testContext(),
		Use:     true,
	}); err != nil {
		t.Fatalf("init context: %v", err)
	}

	if cfg.CurrentContext != "home" {
		t.Fatalf("current context = %q, want home", cfg.CurrentContext)
	}
	if got := cfg.Contexts["home"].Endpoint; got != "https://pve.lan:8006/api2/json" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestInitContextNoUseLeavesCurrentEmpty(t *testing.T) {
	cfg := Empty()

	if err := cfg.InitContext(InitOptions{
		Name:    "home",
		Context: testContext(),
		Use:     false,
	}); err != nil {
		t.Fatalf("init context: %v", err)
	}

	if cfg.CurrentContext != "" {
		t.Fatalf("current context = %q, want empty", cfg.CurrentContext)
	}
	if _, ok := cfg.Contexts["home"]; !ok {
		t.Fatal("expected context to be created")
	}
}

func TestInitContextRejectsExistingWithoutOverwrite(t *testing.T) {
	cfg := Empty()
	if err := cfg.InitContext(InitOptions{Name: "home", Context: testContext(), Use: true}); err != nil {
		t.Fatalf("init context: %v", err)
	}

	replacement := testContext()
	replacement.Endpoint = "https://other.example:8006/api2/json"
	if err := cfg.InitContext(InitOptions{Name: "home", Context: replacement, Use: true}); err == nil {
		t.Fatal("expected existing context error")
	}
	if got := cfg.Contexts["home"].Endpoint; got != "https://pve.lan:8006/api2/json" {
		t.Fatalf("endpoint changed without overwrite: %q", got)
	}
}

func TestInitContextOverwritesExisting(t *testing.T) {
	cfg := Empty()
	if err := cfg.InitContext(InitOptions{Name: "home", Context: testContext(), Use: true}); err != nil {
		t.Fatalf("init context: %v", err)
	}

	replacement := testContext()
	replacement.Endpoint = "https://other.example:8006/api2/json"
	if err := cfg.InitContext(InitOptions{
		Name:      "home",
		Context:   replacement,
		Overwrite: true,
		Use:       true,
	}); err != nil {
		t.Fatalf("overwrite context: %v", err)
	}
	if got := cfg.Contexts["home"].Endpoint; got != "https://other.example:8006/api2/json" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestInitContextValidatesRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Context)
		wantErr string
	}{
		{
			name:    "endpoint",
			mutate:  func(ctx *Context) { ctx.Endpoint = "" },
			wantErr: "endpoint is required",
		},
		{
			name:    "token id",
			mutate:  func(ctx *Context) { ctx.TokenID = "" },
			wantErr: "token-id is required",
		},
		{
			name:    "token secret env",
			mutate:  func(ctx *Context) { ctx.TokenSecretEnv = "" },
			wantErr: "token-secret-env is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testContext()
			tt.mutate(&ctx)
			err := Empty().InitContext(InitOptions{Name: "home", Context: ctx, Use: true})
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
		})
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

func testContext() Context {
	return Context{
		Endpoint:       "https://pve.lan:8006/api2/json",
		TokenID:        "automation@pve!pvectl",
		TokenSecretEnv: "PVECTL_HOME_TOKEN_SECRET",
		Timeout:        "30s",
		DefaultOutput:  "table",
	}
}
