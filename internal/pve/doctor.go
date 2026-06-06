package pve

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lz-wang/pvectl/internal/config"
	"github.com/lz-wang/pvectl/internal/output"
)

type DoctorOptions struct {
	ConfigPath  string
	ContextName string
	Offline     bool
	Node        string
	Timeout     time.Duration
	Insecure    bool
	Output      string
	OutputSet   bool
}

type DoctorResult struct {
	Rows   []output.DoctorRow
	Format string
	Failed bool
}

type DoctorService struct {
	backendFactory func(config.Context, ClientOptions) (Backend, error)
}

func NewDoctorService(backendFactory func(config.Context, ClientOptions) (Backend, error)) *DoctorService {
	if backendFactory == nil {
		backendFactory = NewProxmoxBackend
	}
	return &DoctorService{backendFactory: backendFactory}
}

func (s *DoctorService) Run(ctx context.Context, options DoctorOptions) DoctorResult {
	result := DoctorResult{Format: output.FormatTable}
	if options.OutputSet {
		result.Format = output.NormalizeFormat(options.Output)
	}

	resolved, err := config.ExpandPath(options.ConfigPath)
	if err != nil {
		result.add("CONFIG_PATH", output.DoctorStatusFail, err.Error())
		result.skipLocalAfter("CONFIG_PATH")
		result.skipOnline(options)
		return result
	}
	result.add("CONFIG_PATH", output.DoctorStatusOK, resolved)

	info, err := os.Stat(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.add("CONFIG_FILE", output.DoctorStatusFail, "does not exist")
		} else {
			result.add("CONFIG_FILE", output.DoctorStatusFail, err.Error())
		}
		result.skipLocalAfter("CONFIG_FILE")
		result.skipOnline(options)
		return result
	}
	if info.IsDir() {
		result.add("CONFIG_FILE", output.DoctorStatusFail, "is a directory")
		result.skipLocalAfter("CONFIG_FILE")
		result.skipOnline(options)
		return result
	}
	result.add("CONFIG_FILE", output.DoctorStatusOK, "exists")

	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		result.add("CONFIG_PARSE", output.DoctorStatusFail, err.Error())
		result.skipLocalAfter("CONFIG_PARSE")
		result.skipOnline(options)
		return result
	}
	result.add("CONFIG_PARSE", output.DoctorStatusOK, "yaml parsed")

	contextName, ctxCfg, err := cfg.SelectContext(options.ContextName)
	if err != nil {
		result.add("CURRENT_CONTEXT", output.DoctorStatusFail, err.Error())
		result.skipContextChecks("CURRENT_CONTEXT")
		result.skipOnline(options)
		return result
	}
	result.add("CURRENT_CONTEXT", output.DoctorStatusOK, contextName)

	fieldsOK := result.checkContextFields(ctxCfg)
	tokenSecret, tokenOK := result.checkTokenSecret(ctxCfg)
	timeout, timeoutOK := result.checkTimeout(ctxCfg, options.Timeout)
	defaultOutputOK := result.checkDefaultOutput(ctxCfg)
	endpointOK := result.checkEndpoint(ctxCfg, options.Insecure)

	if !options.OutputSet && defaultOutputOK && ctxCfg.DefaultOutput != "" {
		result.Format = output.NormalizeFormat(ctxCfg.DefaultOutput)
	}

	if options.Offline {
		result.skipOnline(options)
		return result
	}

	if !fieldsOK || !tokenOK || !timeoutOK || !endpointOK {
		result.skipOnline(options)
		return result
	}

	backend, err := s.backendFactory(ctxCfg, ClientOptions{
		TokenSecret: tokenSecret,
		Timeout:     timeout,
		Insecure:    options.Insecure,
	})
	if err != nil {
		result.add("API_CONNECTIVITY", output.DoctorStatusFail, err.Error())
		result.add("NODES", output.DoctorStatusSkip, "skipped due to API_CONNECTIVITY failure")
		result.skipNode(options.Node, "skipped due to API_CONNECTIVITY failure")
		return result
	}

	nodes, err := backend.Nodes(ctx)
	if err != nil {
		result.add("API_CONNECTIVITY", output.DoctorStatusFail, fmt.Sprintf("list nodes: %v", err))
		result.add("NODES", output.DoctorStatusSkip, "skipped due to API_CONNECTIVITY failure")
		result.skipNode(options.Node, "skipped due to API_CONNECTIVITY failure")
		return result
	}

	result.add("API_CONNECTIVITY", output.DoctorStatusOK, "connected")
	result.add("NODES", output.DoctorStatusOK, fmt.Sprintf("%d node(s)", len(nodes)))
	result.checkNode(options.Node, nodes)
	return result
}

func (r *DoctorResult) add(check string, status output.DoctorStatus, message string) {
	r.Rows = append(r.Rows, output.DoctorRow{
		Check:   check,
		Status:  status,
		Message: message,
	})
	if status == output.DoctorStatusFail {
		r.Failed = true
	}
}

func (r *DoctorResult) skipLocalAfter(check string) {
	checks := []string{
		"CONFIG_FILE",
		"CONFIG_PARSE",
		"CURRENT_CONTEXT",
		"CONTEXT_FIELDS",
		"TOKEN_SECRET_ENV",
		"TIMEOUT",
		"DEFAULT_OUTPUT",
		"ENDPOINT",
	}
	reason := fmt.Sprintf("skipped due to %s failure", check)
	for _, name := range checks {
		if name == check {
			continue
		}
		r.add(name, output.DoctorStatusSkip, reason)
	}
}

func (r *DoctorResult) skipContextChecks(check string) {
	reason := fmt.Sprintf("skipped due to %s failure", check)
	for _, name := range []string{"CONTEXT_FIELDS", "TOKEN_SECRET_ENV", "TIMEOUT", "DEFAULT_OUTPUT", "ENDPOINT"} {
		r.add(name, output.DoctorStatusSkip, reason)
	}
}

func (r *DoctorResult) skipOnline(options DoctorOptions) {
	reason := "skipped due to previous failure"
	if options.Offline {
		reason = "skipped in offline mode"
	}
	r.add("API_CONNECTIVITY", output.DoctorStatusSkip, reason)
	r.add("NODES", output.DoctorStatusSkip, reason)
	r.skipNode(options.Node, reason)
}

func (r *DoctorResult) skipNode(node, reason string) {
	if node == "" {
		return
	}
	r.add("NODE", output.DoctorStatusSkip, reason)
}

func (r *DoctorResult) checkContextFields(ctxCfg config.Context) bool {
	var missing []string
	if ctxCfg.Endpoint == "" {
		missing = append(missing, "endpoint")
	}
	if ctxCfg.TokenID == "" {
		missing = append(missing, "token_id")
	}
	if ctxCfg.TokenSecretEnv == "" {
		missing = append(missing, "token_secret_env")
	}
	if len(missing) > 0 {
		r.add("CONTEXT_FIELDS", output.DoctorStatusFail, "missing "+strings.Join(missing, ", "))
		return false
	}
	r.add("CONTEXT_FIELDS", output.DoctorStatusOK, "endpoint, token_id, token_secret_env")
	return true
}

func (r *DoctorResult) checkTokenSecret(ctxCfg config.Context) (string, bool) {
	secret, err := config.ResolveTokenSecret(ctxCfg)
	if err != nil {
		r.add("TOKEN_SECRET_ENV", output.DoctorStatusFail, err.Error())
		return "", false
	}
	r.add("TOKEN_SECRET_ENV", output.DoctorStatusOK, fmt.Sprintf("%s is set", ctxCfg.TokenSecretEnv))
	return secret, true
}

func (r *DoctorResult) checkTimeout(ctxCfg config.Context, override time.Duration) (time.Duration, bool) {
	if override > 0 {
		r.add("TIMEOUT", output.DoctorStatusOK, fmt.Sprintf("%s (from flag)", override))
		return override, true
	}
	if ctxCfg.Timeout == "" {
		r.add("TIMEOUT", output.DoctorStatusWarn, "timeout is empty, runtime default 30s will be used")
		return 0, true
	}
	timeout, err := time.ParseDuration(ctxCfg.Timeout)
	if err != nil {
		r.add("TIMEOUT", output.DoctorStatusFail, fmt.Sprintf("invalid timeout %q: %v", ctxCfg.Timeout, err))
		return 0, false
	}
	r.add("TIMEOUT", output.DoctorStatusOK, ctxCfg.Timeout)
	return timeout, true
}

func (r *DoctorResult) checkDefaultOutput(ctxCfg config.Context) bool {
	if ctxCfg.DefaultOutput == "" {
		r.add("DEFAULT_OUTPUT", output.DoctorStatusWarn, "default_output is empty, table will be used")
		return false
	}
	format := output.NormalizeFormat(ctxCfg.DefaultOutput)
	if err := output.ValidateFormat(format); err != nil {
		r.add("DEFAULT_OUTPUT", output.DoctorStatusFail, err.Error())
		return false
	}
	r.add("DEFAULT_OUTPUT", output.DoctorStatusOK, format)
	return true
}

func (r *DoctorResult) checkEndpoint(ctxCfg config.Context, insecure bool) bool {
	if ctxCfg.Endpoint == "" {
		r.add("ENDPOINT", output.DoctorStatusFail, "endpoint is required")
		return false
	}
	parsed, err := url.Parse(ctxCfg.Endpoint)
	if err != nil {
		r.add("ENDPOINT", output.DoctorStatusFail, err.Error())
		return false
	}
	if parsed.Scheme == "" {
		r.add("ENDPOINT", output.DoctorStatusFail, "endpoint scheme is required")
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		r.add("ENDPOINT", output.DoctorStatusFail, fmt.Sprintf("unsupported endpoint scheme %q", parsed.Scheme))
		return false
	}
	if parsed.Host == "" {
		r.add("ENDPOINT", output.DoctorStatusFail, "endpoint host is required")
		return false
	}

	var warnings []string
	if !strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), "/api2/json") {
		warnings = append(warnings, "path does not end with /api2/json")
	}
	if insecure || ctxCfg.InsecureSkipVerify {
		warnings = append(warnings, "TLS verification disabled")
	}
	if len(warnings) > 0 {
		r.add("ENDPOINT", output.DoctorStatusWarn, fmt.Sprintf("%s (%s)", ctxCfg.Endpoint, strings.Join(warnings, "; ")))
		return true
	}
	r.add("ENDPOINT", output.DoctorStatusOK, ctxCfg.Endpoint)
	return true
}

func (r *DoctorResult) checkNode(node string, nodes []output.NodeRow) {
	if node == "" {
		return
	}
	for _, row := range nodes {
		if row.Name == node {
			r.add("NODE", output.DoctorStatusOK, fmt.Sprintf("node %s exists", node))
			return
		}
	}
	r.add("NODE", output.DoctorStatusFail, fmt.Sprintf("node %s not found", node))
}
