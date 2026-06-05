# Changelog

## v0.7 - 2026-06-05

### Added

- Add `pvectl backup ls` for listing backup files on a specified node and
  storage.
- Add `pvectl vm backup <vmid>` for one-off VM/QEMU backups.
- Add `pvectl lxc backup <vmid>` for one-off LXC backups.
- Add backup output schemas for backup rows and backup task results.
- Add backup filters for `--vmid`, `--kind`, and `--latest`.

### Notes

- Backup support is intentionally limited to listing backup files and creating
  one-off guest backups.
- No backup deletion, restore, prune, scheduled backup job management, or PBS
  management is included.

## v0.6 - 2026-06-05

### Added

- Add read-only `guest` aggregate commands for inventory across VM/QEMU and LXC
  guests.
- Add `pvectl guest ls` with `--type all|vm|lxc`, `--node`, and `--status`
  filters.
- Add `pvectl guest get <vmid>` with `--type auto|vm|lxc`.
- Add aggregate table output with a `KIND` column so VM and LXC rows can be
  distinguished.
- Add GitHub Actions release automation for pushed tags, including multi-platform
  builds, checksums, Release asset upload, and Pushover notifications.
- Add `make dist` for local multi-platform release builds.

### Changed

- Keep ordinary branch pushes to test and build only; tag pushes publish release
  assets.
- Skip duplicate branch build/notification work when `git push origin main --tags`
  also triggers a tag workflow for the same commit.
- Update README and usage docs with the `guest` aggregate workflow.

### Notes

- `guest` is intentionally read-only. Mutating lifecycle and maintenance
  operations remain under `vm` and `lxc`.

## v0.5.1 - 2026-06-05

### Changed

- Refine project positioning docs around the personal HomeLab CLI scope.
- Add agent guidance for repository structure, documentation boundaries, security
  expectations, and common validation commands.
- Update roadmap documentation.

## v0.5 - 2026-06-05

### Added

- Add `pvectl vm reboot <vmid>` and `pvectl lxc reboot <vmid>`.
- Document reboot as part of the daily VM/QEMU and LXC lifecycle workflow.

### Changed

- Reorganize command code by guest operation area:
  - core lifecycle commands
  - clone and config commands
  - maintenance commands
  - dangerous commands
  - snapshot commands
- Restructure usage documentation so daily commands stay separate from
  maintenance, snapshots, dangerous operations, output formats, and scripting
  notes.
- Keep README focused on install, configuration, daily usage, and non-goals.

## v0.4 - 2026-06-05

### Added

- Add VM/QEMU snapshot listing with `pvectl vm snapshot ls <vmid>`.
- Add VM/QEMU snapshot creation with `pvectl vm snapshot create <vmid> <name>`.
- Add VM/QEMU snapshot rollback with
  `pvectl vm snapshot rollback <vmid> <name>`.
- Add LXC snapshot listing with `pvectl lxc snapshot ls <vmid>`.
- Add LXC snapshot creation with `pvectl lxc snapshot create <vmid> <name>`.
- Add LXC snapshot rollback with `pvectl lxc snapshot rollback <vmid> <name>`.
- Add local confirmation for snapshot rollback unless `--force` is passed.

### Notes

- Snapshot create and rollback are asynchronous Proxmox tasks and support
  `--wait`.

## v0.3 - 2026-06-05

### Added

- Add VM/QEMU migration with `pvectl vm migrate <vmid> --target <node>`.
- Add LXC migration with `pvectl lxc migrate <vmid> --target <node>`.
- Add VM/QEMU disk resize with
  `pvectl vm resize <vmid> --disk <disk> --size <size>`.
- Add LXC disk resize with
  `pvectl lxc resize <vmid> --disk <disk> --size <size>`.
- Add VM/QEMU delete with `pvectl vm delete <vmid>`.
- Add LXC delete with `pvectl lxc delete <vmid>`.
- Add local delete confirmation prompts, with `--force` available to skip the
  local prompt.

### Notes

- Delete, migrate, and resize commands support the existing async task waiting
  flow where applicable.

## v0.2 - 2026-06-05

### Added

- Add VM/QEMU clone with `pvectl vm clone <source-vmid>`.
- Add LXC clone with `pvectl lxc clone <source-vmid>`.
- Add clone options for explicit IDs, generated IDs, target node, storage, full
  clone mode, and guest name/hostname.
- Add VM/QEMU config updates with `pvectl vm config <vmid> --set key=value`.
- Add LXC config updates with `pvectl lxc config <vmid> --set key=value`.
- Return clone results with `new_vmid` for script-friendly JSON/YAML output.

## v0.1 - 2026-06-05

### Added

- Add initial `pvectl` CLI entrypoint and version wiring.
- Add YAML config management with contexts:
  - `pvectl config set-context`
  - `pvectl config use-context`
  - `pvectl config current-context`
  - `pvectl config view`
- Store Proxmox API token secret references through `token_secret_env` instead
  of writing token secret values to disk.
- Add global flags for config path, context, output format, timeout, TLS
  verification, and verbose mode.
- Add table, JSON, and YAML output rendering.
- Add node listing with `pvectl node ls`.
- Add VM/QEMU list, get, start, shutdown, and stop commands.
- Add LXC list, get, start, shutdown, and stop commands.
- Add automatic guest lookup across nodes when `--node` is omitted.
- Add async task handling with `--wait` and `--wait-timeout`; task IDs and wait
  progress go to stderr while command results go to stdout.
