# GitBack

A command-line backup utility for GitHub repositories and gists. GitBack keeps your GitHub data mirrored locally and creates compressed snapshots, designed for scheduled, unattended operation e.g. cron, systemd or CI.

**Your backups don't depend on GitBack.** Mirrors are standard bare git repositories, and snapshots are plain `tar.zst` files. If GitBack disappears, `git`, `tar` and `zstd` are all you need to get your data back. See [Restoring without GitBack](#restoring-without-gitback).

## Getting Started

### Installation

```bash
go install github.com/flarexes/gitback/cmd/gitback@latest
```

Or download a [prebuilt binary](https://github.com/FlareXes/gitback/releases).

**Required:**
- Linux (other operating systems are not supported yet)
- `git`
- `tar`
- `zstd`

### First Run

```bash
gitback init
gitback doctor
gitback run
gitback health
```

The `init` command sets up authentication and creates the config file, `doctor` validates that the environment is ready to back up, `run` performs the full backup workflow (discover → sync → snapshot), and `health` reports the current state of the backups.

## Authentication

GitBack needs a GitHub Personal Access Token. You have two options:

### Token from File (Default)

Best for: local machines, interactive setup

```bash
gitback init
```

GitBack will prompt for your token and save it to `~/.local/share/gitback/state/github.token` (mode 0600).

**Token requirements:**
- **Classic PAT:** `repo` scope
- **Fine-Grained PAT:** All repositories, `Contents: Read-only`, `Metadata: Read-only`

### Token from Environment (CI/CD, automation)

Best for: non-interactive environments, shared servers

```bash
export GITBACK_TOKEN="ghp_your_token_here"
gitback init --use-env-token
gitback run
```

The token is read from the environment every run and never written to disk. This is useful in CI/CD pipelines or when you want to manage the token yourself.

**For systems with credential managers:**

```bash
export GITBACK_TOKEN=$(secret-tool lookup service github)
gitback init --use-env-token
```

## Commands

| Command            | What it does                                                                                                          |
| ------------------ | --------------------------------------------------------------------------------------------------------------------- |
| `gitback init`     | Set up authentication and create the config file. `--force` re-initializes an existing install.                       |
| `gitback doctor`   | Check the OS, required binaries, config, token, GitHub authentication, directories and log access.                    |
| `gitback discover` | Query GitHub and save the repository and gist lists to inventory files.                                               |
| `gitback sync`     | Clone or update mirrors, check their integrity, retry on transient failure.                                           |
| `gitback snapshot` | Create a compressed archive of the current mirrors, with a checksum.                                                  |
| `gitback run`      | The full workflow: discover → sync → snapshot.                                                                        |
| `gitback health`   | Report repository and gist status, quarantine, orphaned mirrors, snapshots, disk space, warnings and recommendations. |

`gitback snapshot` refuses to run if any mirror's last sync failed. Run `gitback sync` again, or use `gitback snapshot --force` to archive anyway.

Only one GitBack run can happen at a time. A second run is refused while the first holds the lock. `Ctrl+C` or `SIGTERM` stops a run cleanly: finished work is saved, and unfinished work is retried on the next run.

## Configuration

`gitback init` writes `~/.config/gitback/config.toml`. A missing config file is an error. GitBack never runs on silent defaults. 

These are the defaults:

```toml
[github]
backup_gists = true                           # Include gists (true/false)

[storage]
mirror_root = "/home/user/.local/share/gitback/mirrors"

[snapshot]
output_directory = "/home/user/.local/share/gitback/snapshots"
retention = 0                                 # 0 = keep all, N = keep N newest

[sync]
workers = 3                                   # Parallel clone/update jobs
retry_attempts = 3                            # Retries on transient failure

[health]
minimum_free_disk_percent = 20

[logging]
retention_days = 30                           # Prune logs older than N days (0 = keep forever)
min_keep = 10                                 # Always keep N newest log files
```

## Scheduling

### Cron

```bash
# Daily at 2 AM
0 2 * * * /usr/local/bin/gitback run
```

Cron has no access to your shell environment. If you use `--use-env-token`, set `GITBACK_TOKEN` in the crontab itself.

### Systemd

User units use your own `~`, so the paths in this README apply unchanged.

Save as `/etc/systemd/system/gitback.service`:

```ini
[Unit]
Description=GitBack Backup
After=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/gitback run

# Only needed with --use-env-token (file contains: GITBACK_TOKEN=ghp_...):
# EnvironmentFile=%h/.config/gitback/token.env
```

Save as `/etc/systemd/system/gitback.timer`:

```ini
[Unit]
Description=GitBack Daily Backup

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start:

```bash
sudo systemctl enable --now gitback.timer
sudo systemctl status gitback.timer

# Run even when you're not logged in:
loginctl enable-linger "$USER"
```

If you prefer system-wide units, add `User=` to the service, or `~` will resolve to root's home.

## Where Your Data Lives

```text
~/.config/gitback/
└── config.toml

~/.local/share/gitback/
├── mirrors/
│   ├── repositories/        # bare mirrors, grouped by owner
│   └── gists/               # bare mirrors of gists
├── quarantine/              # corrupt mirrors, kept until replaced (never auto-deleted)
├── snapshots/               # compressed archives + checksums
└── state/
    ├── github.token         # saved token (absent with --use-env-token)
    ├── mirrors.json         # per-asset result of the last sync
    ├── repositories.txt     # inventory from the last discovery
    ├── gists.txt            # inventory from the last discovery
    └── tmp/

~/.local/state/gitback/
└── gitback-YYYY-MM-DD-weekday.log   # one JSON-lines file per day

/tmp/gitback.lock            # run lock
```

`mirrors/`, `snapshots/` and `quarantine/` follow your `[storage]` and `[snapshot]` settings. `quarantine/` sits next to `mirror_root`.

## Snapshots

Each snapshot is named after its UTC creation time:

```text
2026-09-05T14-32-10Z.tar.zst
2026-09-05T14-32-10Z.tar.zst.sha256
```

The archive contains the whole mirror directory `(repos + gists)` plus `mirrors.json`. Quarantine is not included. The `.sha256` file is in standard `sha256sum` format. When `retention` is greater than 0, older snapshots are deleted after a new one succeeds.

## Restoring without GitBack

```bash
# 1. Verify the archive
sha256sum -c 2026-09-05T14-32-10Z.tar.zst.sha256

# 2. Extract everything (or name a path to extract part of it)
zstd -dc 2026-09-05T14-32-10Z.tar.zst | tar -xf -

# 3. Use a mirror like any git repository
git clone mirrors/repositories/<owner>/<repo> my-working-copy
```

You can also extract a single repository by passing its path to `tar -x`. A live mirror (not a snapshot) can be cloned directly with `git clone`.

## Logs

Every operation writes one JSON object per line to `~/.local/state/gitback/gitback-YYYY-MM-DD-weekday.log`. Files rotate daily in local time and are pruned according to `[logging]`.

| Field                           | Meaning                                                        |
| ------------------------------- | -------------------------------------------------------------- |
| `schema_version`                | Log format version                                             |
| `ts`                            | Local time, RFC 3339 with offset                               |
| `level`                         | `INFO`, `WARN`, `ERROR`, `CRITICAL`                            |
| `run_id`                        | Short ID shared by every line of one run                       |
| `host`, `user`                  | Machine and account that ran it                                |
| `component`, `event`, `message` | What happened (`event` is `component.code`)                    |
| `repo`                          | The repository or gist concerned (if any)                      |
| `duration_ms`                   | How long it took (if relevant)                                 |
| `cause`                         | Stable failure category: `auth_failure`, `network_error`, etc. | 
| `error`                         | Raw underlying error text                                      |
| `remediation`                   | How to fix it                                                  |
| `details`                       | Event-specific structured data                                 |

Filter on `cause`, not on `error`. The `error` text can change between versions.

```bash
# Follow today's log
tail -f ~/.local/state/gitback/gitback-*.log | jq

# All problems
jq 'select(.level=="ERROR" or .level=="CRITICAL")' ~/.local/state/gitback/gitback-*.log

# Everything from one run
jq 'select(.run_id=="a1b2c3d4")' ~/.local/state/gitback/gitback-*.log

# Only authentication failures
jq 'select(.cause=="auth_failure")' ~/.local/state/gitback/gitback-*.log
```

Logs contain repository names, hostnames and usernames. Remove them before sharing logs with anyone, including AI tools.

### Analyzing logs with AI

For large or confusing failures, you can give the logs to an AI assistant. For privacy-conscious users, we suggest [Lumo by Proton](https://lumo.proton.me/), Proton's privacy-focused AI assistant.

Send whole file or only the lines that matter, from one run:

```bash
# 1. Find the run you care about (the latest problems)
jq -r 'select(.level=="ERROR" or .level=="CRITICAL") | [.ts,.run_id,.event] | @tsv' \
  ~/.local/state/gitback/gitback-*.log | tail

# 2. Export that run's warnings and errors, without host and user
jq -c 'select(.run_id=="a1b2c3d4" and .level!="INFO") | del(.host,.user)' \
  ~/.local/state/gitback/gitback-*.log > gitback-issue.jsonl
```

Paste or upload logs and ask something like: 
- *"Explain these GitBack log entries and what I should fix first."*
- *"Why fshare repo failed to recover from quarantine?"*.

## Troubleshooting

Start with `gitback doctor`, then `gitback health`, then the logs. Every warning or error line has a `remediation` field.

| Symptom                                                      | Meaning and fix                                                                                                                                                                                                                               |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cause: lock_held`                                           | Another GitBack run is in progress. Wait for it to finish. A crashed run never leaves a stale lock.                                                                                                                                           |
| `cause: cancelled`, or `health` reports *interrupted* assets | A run was stopped (Ctrl+C, SIGTERM). This is not a failure. Run `gitback sync` again.                                                                                                                                                         |
| `cause: auth_failure`                                        | The token is invalid, expired or missing a scope. See [Authentication](#authentication).                                                                                                                                                      |
| Snapshot refused: mirror verification failed                 | Some mirrors failed their last sync. Run `gitback sync`, or use `gitback snapshot --force`.                                                                                                                                                   |
| Mirrors in `quarantine/`                                     | GitBack found a corrupt mirror, moved it aside and re-cloned it. The quarantined copy is deleted only after the replacement is verified. GitBack never deletes quarantine otherwise.                                                          |
| `health` lists *orphaned* mirrors                            | These mirrors exist on disk but the latest discovery no longer lists them. They may be deleted, renamed, or no longer visible to your token. They are kept as backups but no longer updated. Review them yourself. GitBack won't remove them. |

## Principles

- **Simplicity**: the smallest design that fully solves the problem.
- **Transparency**: every log line says what happened, when, why, and how to fix it.
- **Reliability**: it's a backup tool, so it fails loudly instead of quietly.

## Contributing

Issues and pull requests are welcome. Keep to the three principles above. When in doubt, choose the simpler design.

## License

BSD 3-Clause "New" or "Revised" License. See [LICENSE](LICENSE).
