# Compatibility Policy

`pvectl` v1.x is intended to be stable for personal HomeLab scripts and
automation. This policy applies to documented behavior.

## Stable Within v1.x

- Command names and documented subcommand structure.
- Positional argument order and meaning.
- Documented flags and their accepted value shapes.
- JSON and YAML output field names and field types.
- Command results on stdout.
- Task IDs and wait progress on stderr.
- Local confirmation behavior for dangerous operations.
- Doctor diagnostics as structured rows, including failure rows on stdout.

## Allowed Non-breaking Changes

- Add new commands.
- Add new optional flags.
- Add new JSON/YAML fields.
- Improve table output formatting.
- Improve error messages without changing exit semantics.
- Add more doctor diagnostic rows.

## Breaking Changes

These require a new major version:

- Remove or rename documented commands.
- Remove or rename documented flags.
- Change positional argument meaning.
- Remove or rename JSON/YAML fields.
- Change JSON/YAML field types.
- Move command result output from stdout to stderr.
- Move task IDs or wait progress from stderr to stdout.
- Skip dangerous-operation confirmation by default.

## Table Output

Table output is for humans and may be adjusted for readability in v1.x. Scripts
should use `-o json` or `-o yaml`.

## Non-goals

See the README for project non-goals. In particular, v1.x compatibility does
not imply a future server mode, Web UI, RBAC layer, storage mutation surface,
restore/prune workflow, PBS management, or full Proxmox REST API passthrough.
