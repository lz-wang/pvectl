# pvectl v0.1 Usage

## Global Flags

```bash
pvectl \
  --config ~/.config/pvectl/config.yaml \
  --context home \
  -o table \
  --timeout 30s \
  --insecure \
  --verbose \
  <resource> <action>
```

Supported output formats are `table`, `json`, and `yaml`.

For async guest control commands, use:

```bash
pvectl vm start 100 --wait --wait-timeout 5m
```

Task IDs and wait progress are written to stderr. Command results are written
to stdout.

## Config

```bash
pvectl config set-context home \
  --endpoint https://pve.lan:8006/api2/json \
  --token-id automation@pve!pvectl \
  --token-secret-env PVECTL_HOME_TOKEN_SECRET \
  --insecure \
  --timeout 30s \
  --default-output table

pvectl config use-context home
pvectl config current-context
pvectl config view
```

Config schema:

```yaml
current_context: home
contexts:
  home:
    endpoint: https://pve.lan:8006/api2/json
    token_id: automation@pve!pvectl
    token_secret_env: PVECTL_HOME_TOKEN_SECRET
    insecure_skip_verify: true
    timeout: 30s
    default_output: table
```

The token secret must be supplied through the named environment variable:

```bash
export PVECTL_HOME_TOKEN_SECRET="your-token-secret"
```

## Nodes

```bash
pvectl node ls
pvectl node ls -o json
```

## VM/QEMU

```bash
pvectl vm ls
pvectl vm ls --node pve1
pvectl vm get 100
pvectl vm get 100 --node pve1 -o yaml
pvectl vm start 100 --wait
pvectl vm shutdown 100 --wait
pvectl vm stop 100
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the VMID automatically.

## LXC

```bash
pvectl lxc ls
pvectl lxc ls --node pve1
pvectl lxc get 200
pvectl lxc get 200 --node pve1 -o json
pvectl lxc start 200 --wait
pvectl lxc shutdown 200 --wait
pvectl lxc stop 200
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the CTID automatically.
