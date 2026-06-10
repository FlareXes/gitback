# GitBack

Simple, transparent GitHub repository backups.

GitBack discovers repositories from GitHub, maintains local mirrors, and creates compressed snapshots for long-term storage. It is designed to run unattended while remaining easy to inspect, troubleshoot, and recover from.

## Features

- Backup Public and private GitHub repository
    
- Local Git mirror synchronization
    
- Compressed snapshots (`tar + zstd`)
    
- SHA256 checksum for integrity check
    
- Structured JSON logging
    
- Health reporting
    
- Environment diagnostics
    
- No database
    
- No daemon
    
- No proprietary formats
    

## Installation

### Go Install

```bash
go install github.com/flarexes/gitback/cmd/gitback@latest
```

### Build From Source

```bash
git clone https://github.com/flarexes/gitback.git

cd gitback

go build -o gitback ./cmd/gitback
```

## Requirements

Required tools:

- git
    
- tar
    
- zstd
    

Verify your environment:

```bash
gitback doctor
```

## Commands

### Initialize

```bash
gitback init
```

Creates configuration and validates GitHub authentication.

### Discover

```bash
gitback discover
```

Discovers repositories accessible to the configured GitHub account.

### Sync

```bash
gitback sync
```

Creates and updates local Git mirrors.

### Snapshot

```bash
gitback snapshot
```

Creates a compressed archive containing all mirrored repositories.

Force mode:

```bash
gitback snapshot --force
```

Continues snapshot creation even if repository sync was failed.

### Health

```bash
gitback health
```

Displays repository status, snapshot information, storage usage, and recommendations.

### Doctor

```bash
gitback doctor
```

Performs environment and configuration checks. Recommendation after initialization, `gitback init`.

## GitHub Token Permissions

GitBack supports either a **Classic Personal Access Token** or a **Fine-Grained Personal Access Token**.

### Classic PAT

Scope:

```text
repo
```

### Fine-Grained PAT

Repository Access:

```text
All repositories
```

Permissions:

```text
Contents: Read-only
Metadata: Read-only
```

Any one token type is required.

## Directory Layout

Configuration:

```text
~/.config/gitback/
└── config.yaml
```

Data:

```text
~/.local/share/gitback/
├── mirrors/
├── snapshots/
└── state/
    ├── github.token
    ├── repositories.txt
    └── mirrors.json
```

Runtime state:

```text
~/.local/state/gitback/
├── gitback.log
└── gitback.lock
```

## Logging

GitBack produces structured JSON logs.

Example:

```json
{
  "timestamp": "2026-06-07T10:00:00Z",
  "run_id": "7b3f4a1c",
  "level": "info",
  "event": "sync_completed"
}
```

## Snapshots

Snapshots are stored as:

```text
YYYY-MM-DDTHH-MM-SSZ.tar.zst
```

Each snapshot includes:

```text
mirrors/
state/mirrors.json
```

Snapshot retention can be configured to automatically remove older snapshots.

Example:

```yaml
snapshot_retention: 30
```

Retains the newest 30 snapshots. Retention is disabled by default (`0 or < 1`).

GitBack also generates SHA256 checksum files alongside snapshots.

## Automation

Typical unattended workflow:

```bash
gitback discover
gitback sync
gitback snapshot --force
```

Can be scheduled using:

- cron
    
- systemd timers
    
- CI/CD pipelines
    

## Roadmap

- [ ] Windows and macOS support
    
- [ ] Multi-worker synchronization
    
- [ ] Git retry and backoff support
    
- [ ] GitHub organization support
    
- [ ] Repository filtering


## Contributing

Bug reports, feature requests, and pull requests are welcome.

Please keep contributions aligned with the project's core principles:

- Simplicity
    
- Transparency
    
- Reliability
    

## License

BSD 3-Clause License.

See [LICENSE](LICENSE) for details.

## Why GitBack?

GitBack is built to solve a straightforward problem: reliably backing up Git repositories without unnecessary complexity.

Many existing solutions provide hosted services, dashboards, integrations, and management platforms. Those tools provide real value and are often the right choice for teams that need them.

However, many individuals, open source maintainers, and small teams simply need dependable repository backups they can run themselves.

GitBack focuses on that use case using standard Git mirrors, standard archive formats, and straightforward recovery procedures.

If GitBack disappears tomorrow, your backups remain usable with standard tools.

The project is still young and will continue to evolve, but simplicity, reliability, and operational transparency will remain the primary design goals.
