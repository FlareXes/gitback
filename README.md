# GitBack

Reliable, unattended GitHub backup designed for long-term operation.

GitBack is a command-line utility that creates and maintains local mirror backups of GitHub repositories and gists.

Unlike traditional backup scripts, GitBack is designed to run continuously as a scheduled job with an emphasis on reliability, operational visibility, and recoverability.

## Features

- Backup GitHub repositories
- Backup GitHub gists (optional)
- Incremental concurrent synchronization
- Snapshot creation
- Repository integrity verification
- Health reporting
- Environment diagnostics
- Structured JSON logging
- Atomic state persistence

# Installation

### Pre-requisites

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

# Commands

## Initialize

Creates configuration and validates GitHub authentication.

```bash
gitback init
```

## Discover

Discovers repositories and gists accessible to the configured GitHub account.

```bash
gitback discover
```

## Sync

Creates and updates local Git mirrors.

```bash
gitback sync
```

## Snapshot

Creates a compressed archive containing all mirrored repositories, gists, and backup state.

```bash
gitback snapshot
```

**Force Mode:**

By default, GitBack refuses to create a snapshot when synchronization failures are detected from the last `gitback sync`. Use `--force` to create a snapshot anyway.

```bash
gitback snapshot --force
```

## Health

The `health` command reports the current state of a backup installation, including:

- Repository statistics
- Gist statistics
- Snapshot information
- Warnings
- Recommendations

```bash
gitback health
```

## Doctor

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

# GitHub Token Permissions

GitBack supports either a **Classic Personal Access Token** or a **Fine-Grained Personal Access Token**.

## Classic PAT

Scope:

```text
repo
```

## Fine-Grained PAT

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

# Logging

GitBack writes structured JSON logs intended for machine consumption and easy to investigate manually.

Every log entry contains structured fields describing the operation, making logs suitable for:

- SIEM platforms
- Centralized logging
- Incident investigation
- Automation
- Long-term auditing

# Snapshots

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

# Automation

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

# Roadmap

- [ ] Windows and macOS support
    
- [ ] Multi-worker synchronization
    
- [x] Git retry and backoff support
    
- [ ] GitHub organization support
    
- [ ] Repository filtering

- [ ] Wiki backups

- [ ] Improved mirror self-healing

- [ ] Additional health diagnostics


# Contributing

Bug reports, feature requests, and pull requests are welcome.

Please keep contributions aligned with the project's core principles:

- Simplicity
    
- Transparency
    
- Reliability
    

# License

BSD 3-Clause License.

See [LICENSE](LICENSE) for details.

# Why GitBack?

GitBack is built to solve a straightforward problem: reliably backing up Git repositories without unnecessary complexity.

Many existing solutions provide hosted services, dashboards, integrations, and management platforms. Those tools provide real value and are often the right choice for teams that need them.

However, many individuals, open source maintainers, and small teams simply need dependable repository backups they can run themselves.

GitBack focuses on that use case using standard Git mirrors, standard archive formats, and straightforward recovery procedures.

If GitBack disappears tomorrow, your backups remain usable with standard tools.

The project is still young and will continue to evolve, but simplicity, reliability, and operational transparency will remain the primary design goals.
