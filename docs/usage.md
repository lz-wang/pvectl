# pvectl Usage

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
pvectl vm clone 9000 --newid 101 --name app-vm --target pve1 --wait
pvectl vm clone 9000 --name app-vm --target pve1 --storage local-lvm --full --wait
pvectl vm config 101 --set memory=4096 --set cores=4 --wait
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the VMID automatically.

For clone, omit `--newid` to let Proxmox allocate the next available VMID.
Clone results are written to stdout and include `new_vmid`, so scripts can
capture the actual allocated ID:

```bash
pvectl vm clone 9000 --name app-vm --target pve1 -o json
```

## LXC

```bash
pvectl lxc ls
pvectl lxc ls --node pve1
pvectl lxc get 200
pvectl lxc get 200 --node pve1 -o json
pvectl lxc start 200 --wait
pvectl lxc shutdown 200 --wait
pvectl lxc stop 200
pvectl lxc clone 900 --newid 201 --hostname app-lxc --target pve1 --wait
pvectl lxc clone 900 --hostname app-lxc --target pve1 --storage local-lvm --full --wait
pvectl lxc config 201 --set memory=2048 --set cores=2 --wait
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the CTID automatically.

For clone, omit `--newid` to let Proxmox allocate the next available CTID.
The JSON/YAML field is still named `new_vmid` to match the existing guest DTOs.
