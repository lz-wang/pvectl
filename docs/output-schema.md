# Output Schema

JSON and YAML output are script-facing and stable within v1.x. New fields may be
added in minor releases, but existing field names and types will not be removed,
renamed, or changed without a new major version. Table output is for humans and
should not be parsed by scripts.

Fields marked optional use `omitempty` and may be absent when the value is empty
or zero.

## NodeRow

| Field | Type |
| --- | --- |
| `name` | string |
| `status` | string |
| `cpu` | number |
| `mem` | uint64 |
| `max_mem` | uint64 |
| `disk` | uint64 |
| `max_disk` | uint64 |
| `uptime` | uint64 |

## GuestRow

Used by `guest`, `vm`, and `lxc` list/detail commands.

| Field | Type |
| --- | --- |
| `kind` | string |
| `vmid` | uint64 |
| `name` | string |
| `node` | string |
| `status` | string |
| `cpus` | int |
| `cpu` | number |
| `mem` | uint64 |
| `max_mem` | uint64 |
| `max_disk` | uint64 |
| `uptime` | uint64 |
| `tags` | string, optional |

## CloneResult

| Field | Type |
| --- | --- |
| `kind` | string |
| `source_vmid` | uint64 |
| `new_vmid` | uint64 |
| `source_node` | string |
| `target_node` | string |
| `name` | string |
| `task` | string, optional |

## SnapshotRow

| Field | Type |
| --- | --- |
| `kind` | string |
| `vmid` | uint64 |
| `node` | string |
| `name` | string |
| `description` | string |
| `parent` | string |
| `snaptime` | int64 |
| `vmstate` | int |
| `state` | string |

## BackupRow

| Field | Type |
| --- | --- |
| `node` | string |
| `storage` | string |
| `kind` | string |
| `vmid` | uint64 |
| `volid` | string |
| `format` | string |
| `size` | uint64 |
| `used` | uint64, optional |
| `ctime` | uint64 |
| `protected` | string, optional |
| `encrypted` | string, optional |
| `verify_state` | string, optional |
| `notes` | string, optional |

## BackupResult

| Field | Type |
| --- | --- |
| `kind` | string |
| `vmid` | uint64 |
| `node` | string |
| `storage` | string |
| `mode` | string |
| `task` | string, optional |

## StorageRow

| Field | Type |
| --- | --- |
| `node` | string |
| `storage` | string |
| `type` | string |
| `active` | bool |
| `enabled` | bool |
| `shared` | bool |
| `content` | string |
| `used` | uint64 |
| `avail` | uint64 |
| `total` | uint64 |
| `used_fraction` | number |

## StorageContentRow

| Field | Type |
| --- | --- |
| `node` | string |
| `storage` | string |
| `content` | string |
| `vmid` | uint64, optional |
| `volid` | string |
| `format` | string, optional |
| `size` | uint64 |
| `used` | uint64, optional |
| `ctime` | uint64, optional |
| `protected` | string, optional |
| `encrypted` | string, optional |
| `verify_state` | string, optional |
| `notes` | string, optional |

## DoctorRow

| Field | Type |
| --- | --- |
| `check` | string |
| `status` | string |
| `message` | string |

Known `status` values are `ok`, `warn`, `fail`, and `skip`.

## VersionInfo

| Field | Type |
| --- | --- |
| `version` | string |
| `commit` | string |
| `date` | string |
| `go_version` | string |
| `os` | string |
| `arch` | string |
