# pvectl Usage

`pvectl` is a personal HomeLab Proxmox VE CLI for daily VM/QEMU and LXC
operations.

## Configuration

For a typical HomeLab setup, create one context and make it current:

```bash
export PVECTL_HOME_TOKEN_SECRET="your-token-secret"

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

The token secret is read from the named environment variable at runtime.
`pvectl` does not write token secrets to disk.

## Daily Commands

### Nodes

```bash
pvectl node ls
```

### Guest Aggregate View

`guest` is a read-only aggregate view across VM/QEMU and LXC guests.

```bash
pvectl guest ls
pvectl guest ls --node pve1
pvectl guest ls --type vm
pvectl guest ls --type lxc
pvectl guest ls --status running
pvectl guest get 100
pvectl guest get 100 --type vm
pvectl guest get 200 --type lxc
```

Use `guest` for inventory and inspection. Use `vm` and `lxc` commands for
lifecycle and maintenance operations.

`guest ls` defaults to `--type all`. `guest get` defaults to `--type auto` and
searches both VM/QEMU and LXC guests. If both a VM and an LXC exist with the
same ID, specify `--type vm` or `--type lxc`.

### VM/QEMU

```bash
pvectl vm ls
pvectl vm ls --node pve1
pvectl vm get 100
pvectl vm get 100 --node pve1 -o json
pvectl vm start 100 --wait
pvectl vm shutdown 100 --wait
pvectl vm reboot 100 --wait
pvectl vm stop 100
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the VMID automatically.

### LXC

```bash
pvectl lxc ls
pvectl lxc ls --node pve1
pvectl lxc get 200
pvectl lxc get 200 --node pve1 -o json
pvectl lxc start 200 --wait
pvectl lxc shutdown 200 --wait
pvectl lxc reboot 200 --wait
pvectl lxc stop 200
```

When `--node` is omitted, `pvectl` traverses all nodes returned by the cluster
and resolves the CTID automatically.

## Maintenance Commands

### Clone

```bash
pvectl vm clone 9000 --newid 101 --name app-vm --target pve1 --wait
pvectl vm clone 9000 --name app-vm --target pve1 --storage local-lvm --full --wait

pvectl lxc clone 900 --newid 201 --hostname app-lxc --target pve1 --wait
pvectl lxc clone 900 --hostname app-lxc --target pve1 --storage local-lvm --full --wait
```

Omit `--newid` to let Proxmox allocate the next available VMID/CTID. Clone
results are written to stdout and include `new_vmid`, so scripts can capture
the allocated ID:

```bash
pvectl vm clone 9000 --name app-vm --target pve1 -o json
```

### Config

```bash
pvectl vm config 101 --set memory=4096 --set cores=4 --wait
pvectl lxc config 201 --set memory=2048 --set cores=2 --wait
```

`config` passes generic `key=value` options to the Proxmox guest config API.

### Resize

```bash
pvectl vm resize 101 --disk scsi0 --size +20G --wait
pvectl lxc resize 201 --disk rootfs --size +10G --wait
```

### Migrate

```bash
pvectl vm migrate 101 --target pve2 --online --wait
pvectl lxc migrate 201 --target pve2 --online --wait
```

## Snapshot Commands

```bash
pvectl vm snapshot ls 101
pvectl vm snapshot create 101 before-upgrade --wait

pvectl lxc snapshot ls 201
pvectl lxc snapshot create 201 before-upgrade --wait
```

Snapshot rollback is a dangerous operation and is documented separately below.

## Dangerous Operations

### Delete

Delete commands require a local confirmation prompt unless `--force` is passed:

```bash
pvectl vm delete 101
pvectl lxc delete 201
```

The prompt requires typing the exact VMID/CTID. The `--force` flag only skips
this local prompt; it is not passed to the Proxmox LXC delete API.

Use `--wait` when scripts need completion status:

```bash
pvectl vm delete 101 --force --wait
pvectl lxc delete 201 --force --wait
```

### Snapshot Rollback

Snapshot rollback commands require typing the exact snapshot name unless
`--force` is passed:

```bash
pvectl vm snapshot rollback 101 before-upgrade
pvectl lxc snapshot rollback 201 before-upgrade
```

The `--force` flag only skips this local prompt. Rollback is an asynchronous
PVE task, so use `--wait` when scripts need completion status:

```bash
pvectl vm snapshot rollback 101 before-upgrade --force --wait
pvectl lxc snapshot rollback 201 before-upgrade --force --wait
```

## Output Formats

Supported output formats are `table`, `json`, and `yaml`.

```bash
pvectl node ls -o table
pvectl guest ls -o json
pvectl vm get 100 -o json
pvectl lxc get 200 -o yaml
```

Use `table` for interactive use, `json` for scripts and agents, and `yaml` as
an optional human-readable structured format.

## Scripting Notes

Global flags:

```bash
pvectl \
  --config ~/.config/pvectl/config.yaml \
  --context home \
  -o json \
  --timeout 30s \
  --insecure \
  --verbose \
  <resource> <action>
```

Async guest operations support:

```bash
pvectl vm reboot 100 --wait --wait-timeout 5m
```

Task IDs and wait progress are written to stderr. Command results are written
to stdout.
