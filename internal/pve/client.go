package pve

import (
	"fmt"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"

	"github.com/lz-wang/pvectl/internal/config"
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
