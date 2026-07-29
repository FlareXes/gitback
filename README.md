# GitBack

Reliable, unattended GitHub backup designed for long-term operation.

GitBack is a command-line utility that creates and maintains local mirror backups of GitHub repositories and gists.

Unlike traditional backup scripts, GitBack is designed to run continuously as a scheduled job with an emphasis on reliability, operational visibility, and recoverability.

## Features

- Backup GitHub repositories
- Backup GitHub gists
- Self-heal corrupt mirrors
- Incremental concurrent synchronization
- Snapshot creation
- Automatic snapshot retention
- Repository integrity verification
- Health reporting
- Environment diagnostics
- Structured JSON logging
- Atomic state persistence

## Installation

### Pre-requisites

- Go 1.24+ (building from source)

- git

- tar

- zstd

### Install

```bash
go install github.com/flarexes/gitback/cmd/gitback@latest
```

### Build From Source (Unstable)

```bash
git clone https://github.com/flarexes/gitback.git

cd gitback

go build -o gitback ./cmd/gitback
```

## Commands

### Initialize

Creates configuration and validates GitHub authentication.

```bash
gitback init
```

### Discover

Discovers repositories and gists accessible to the configured GitHub account.

```bash
gitback discover
```

### Sync

Creates and updates local Git mirrors.

```bash
gitback sync
```

### Snapshot

Creates a compressed archive containing all mirrored repositories, gists, and backup state.

```bash
gitback snapshot
```

**Force Mode:**

By default, GitBack refuses to create a snapshot when synchronization failures are detected from the last `gitback sync`. Use `--force` to create a snapshot anyway.

```bash
gitback snapshot --force
```

### Health

The `health` command reports the current state of a backup installation, including:

- Repository statistics
- Gist statistics
- Snapshot information
- Warnings
- Recommendations

```bash
gitback health
```

### Doctor

The `doctor` command validates whether GitBack is able to perform backups.

It verifies:

- Supported operating system
- Required executables
- Configuration
- Authentication
- Required directories
- Log file accessibility

```bash
gitback doctor
```

## Authentication

GitBack authenticates using a GitHub Personal Access Token (PAT).

By default, `gitback init` stores the token locally.

Alternatively, GitBack can read the token from the `GITBACK_TOKEN` environment variable. When present, the environment variable takes precedence over the stored token.

Example:

```bash
export GITBACK_TOKEN=ghp_xxxxxxxxxxxxxxxxx

gitback run
```

Users who prefer not to store credentials on disk can remove the token file after initialization:

```bash
rm ~/.local/share/gitback/state/github.token
```

GitBack will then require `GITBACK_TOKEN` to be set before running.

### GitHub Token Permissions

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

### Using External Secret Managers

GitBack can be used with any external secret management solution without requiring special integration.

For example, using `secret-tool`:

```bash
secret-tool store  --label="GitBack GitHub Token"  github token
```

Run GitBack:

```bash
export GITBACK_TOKEN="$(secret-tool lookup github token)"
gitback run
```

This allows the token to remain outside GitBack while still requiring no changes to GitBack itself.

## Logging

GitBack writes structured JSON logs intended for machine consumption and easy to investigate manually.

Every log entry contains structured fields describing the operation, making logs suitable for:

- SIEM platforms
- Centralized logging
- Incident investigation
- Automation
- Long-term auditing

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
[snapshot]
retention = 30
```

Retains the newest 30 snapshots. Retention is disabled by default (`0 or < 1`).

GitBack also generates SHA256 checksum files alongside snapshots.

## Automation

`gitback run` performs repository discovery, mirror synchronization, and snapshot creation as a single unattended workflow.

Can be scheduled using:

- cron
    
- systemd timers
    
- CI/CD pipelines

### Cron

Run daily at 02:00:

```cron
0 2 * * * GITBACK_TOKEN="$(secret-tool lookup github token)" /usr/local/bin/gitback run
```

### Systemd

Create: ~/.config/systemd/user/gitback.service

```ini
[Unit]
Description=GitBack backup

[Service]
Type=oneshot
ExecStart=/usr/local/bin/gitback run
```
Create: ~/.config/systemd/user/gitback.timer

```ini
[Unit]
Description=Run GitBack daily

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

Enable service:

```bash
systemctl --user enable --now gitback.timer
```

## Roadmap

- [ ] Windows and macOS support
    
- [x] Multi-worker synchronization
    
- [x] Git retry and backoff support
    
- [ ] GitHub organization support
    
- [ ] Repository filtering

- [ ] Wiki backups

- [x] Improved mirror self-healing

- [ ] Additional health diagnostics


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
