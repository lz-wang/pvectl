package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultPath = "~/.config/pvectl/config.yaml"

type Config struct {
	CurrentContext string             `yaml:"current_context,omitempty"`
	Contexts       map[string]Context `yaml:"contexts,omitempty"`
}

type Context struct {
	Endpoint           string `yaml:"endpoint,omitempty"`
	TokenID            string `yaml:"token_id,omitempty"`
	TokenSecretEnv     string `yaml:"token_secret_env,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
	Timeout            string `yaml:"timeout,omitempty"`
	DefaultOutput      string `yaml:"default_output,omitempty"`
}

type InitOptions struct {
	Name      string
	Context   Context
	Overwrite bool
	Use       bool
}

func Empty() *Config {
	return &Config{Contexts: map[string]Context{}}
}

func Load(path string) (*Config, error) {
	resolved, err := ExpandPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", resolved, err)
	}

	cfg := Empty()
	if len(data) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", resolved, err)
	}
	cfg.ensure()
	return cfg, nil
}

func LoadOrEmpty(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return Empty(), nil
	}
	return nil, err
}

func Save(path string, cfg *Config) error {
	resolved, err := ExpandPath(path)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = Empty()
	}
	cfg.ensure()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(resolved, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", resolved, err)
	}
	return nil
}

func ToYAML(cfg *Config) ([]byte, error) {
	if cfg == nil {
		cfg = Empty()
	}
	cfg.ensure()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode config: %w", err)
	}
	return data, nil
}

func ExpandPath(path string) (string, error) {
	if path == "" {
		path = DefaultPath
	}
	path = os.ExpandEnv(path)

	if path == "~" || len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}

	return filepath.Clean(path), nil
}

func (c *Config) SetContext(name string, ctx Context) error {
	if name == "" {
		return errors.New("context name is required")
	}
	if err := validateContext(ctx); err != nil {
		return err
	}
	c.ensure()
	c.Contexts[name] = ctx
	if c.CurrentContext == "" {
		c.CurrentContext = name
	}
	return nil
}

func (c *Config) InitContext(options InitOptions) error {
	name := strings.TrimSpace(options.Name)
	if name == "" {
		return errors.New("context name is required")
	}
	if err := validateContext(options.Context); err != nil {
		return err
	}

	c.ensure()
	if _, exists := c.Contexts[name]; exists && !options.Overwrite {
		return fmt.Errorf("context %q already exists; use --overwrite to replace it", name)
	}
	c.Contexts[name] = options.Context
	if options.Use {
		c.CurrentContext = name
	}
	return nil
}

func (c *Config) UseContext(name string) error {
	if name == "" {
		return errors.New("context name is required")
	}
	c.ensure()
	if _, ok := c.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	c.CurrentContext = name
	return nil
}

func (c *Config) SelectContext(name string) (string, Context, error) {
	c.ensure()
	if name == "" {
		name = c.CurrentContext
	}
	if name == "" {
		return "", Context{}, errors.New("context is required; set current_context or pass --context")
	}
	ctx, ok := c.Contexts[name]
	if !ok {
		return "", Context{}, fmt.Errorf("context %q not found", name)
	}
	return name, ctx, nil
}

func ResolveTokenSecret(ctx Context) (string, error) {
	if ctx.TokenSecretEnv == "" {
		return "", errors.New("token_secret_env is required")
	}
	secret := os.Getenv(ctx.TokenSecretEnv)
	if secret == "" {
		return "", fmt.Errorf("environment variable %s is empty", ctx.TokenSecretEnv)
	}
	return secret, nil
}

func validateContext(ctx Context) error {
	if ctx.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if ctx.TokenID == "" {
		return errors.New("token-id is required")
	}
	if ctx.TokenSecretEnv == "" {
		return errors.New("token-secret-env is required")
	}
	return nil
}

func (c *Config) ensure() {
	if c.Contexts == nil {
		c.Contexts = map[string]Context{}
	}
}
