# pvectl

`pvectl` is a small Proxmox VE CLI focused on daily VM/QEMU and LXC
operations. It wraps the Proxmox VE API through
[`go-proxmox`](https://github.com/luthermonson/go-proxmox) and keeps user
commands resource-oriented instead of exposing raw API paths.

## v0.1 scope

- Manage config contexts backed by YAML.
- List PVE nodes.
- List and inspect VM/QEMU guests.
- List and inspect LXC containers.
- Start, gracefully shut down, and stop VM/LXC guests.
- Optionally wait for async PVE tasks.
- Render `table`, `json`, or `yaml` output.

## Install from source

```bash
make build
```

The binary is written to `bin/pvectl`.

## Configure

Create an API token in Proxmox VE, export the token secret, then create a
context:

```bash
export PVECTL_HOME_TOKEN_SECRET="your-token-secret"

pvectl config set-context home \
  --endpoint https://pve.lan:8006/api2/json \
  --token-id automation@pve!pvectl \
  --token-secret-env PVECTL_HOME_TOKEN_SECRET \
  --insecure \
  --timeout 30s \
  --default-output table
```

`pvectl` stores only the environment variable name in
`~/.config/pvectl/config.yaml`; it does not write token secrets to disk.

## Examples

```bash
pvectl node ls
pvectl vm ls
pvectl vm get 100 -o json
pvectl vm start 100 --wait
pvectl vm shutdown 100 --wait
pvectl vm stop 100

pvectl lxc ls --node pve1
pvectl lxc get 200
pvectl lxc start 200 --wait
```

See [docs/usage.md](docs/usage.md) for the complete v0.1 command list.
