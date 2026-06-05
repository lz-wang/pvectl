# Changelog

## v0.6.0

### Added

- Add read-only `guest` aggregate command.
- Add `pvectl guest ls` to list VM/QEMU and LXC guests together.
- Add `pvectl guest get <vmid>` to inspect either VM/QEMU or LXC guests.
- Add `--type all|vm|lxc` for `guest ls`.
- Add `--type auto|vm|lxc` for `guest get`.
- Add `--status` filter for `guest ls`.
- Add table output with `KIND` column for aggregate guest views.

### Unchanged

- `vm` and `lxc` lifecycle commands remain the only mutation paths.
- `guest` is read-only.
- No server mode, Web UI, RBAC, audit, database, or REST API passthrough.
