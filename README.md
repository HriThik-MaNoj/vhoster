# vhoster

[![CI](https://github.com/HriThik-MaNoj/vhoster/actions/workflows/ci.yml/badge.svg)](https://github.com/HriThik-MaNoj/vhoster/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/HriThik-MaNoj/vhoster)](go.mod)
[![License](https://img.shields.io/github/license/HriThik-MaNoj/vhoster)](LICENSE)

**TUI + CLI tool** for managing virtual-host entries in `/etc/hosts`. Safely adds, removes, and lists entries inside a managed block — never touching your hand-written lines.

Written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/HriThik-MaNoj/vhoster/main/install.sh | sh
```

The script auto-detects your OS and arch, picks the fastest method (go install or prebuilt binary), and places `vhoster` in `/usr/bin/`.

## Usage

```bash
# Interactive TUI
sudo vhoster

# Quick operations (scripting-friendly)
vhoster list                              # print managed entries (no root)
sudo vhoster add 10.0.0.5 api.local admin.local
sudo vhoster remove api.local
sudo vhoster remove --ip 10.0.0.5

# Backup management
vhoster restore --list                    # list backups (no root)
sudo vhoster restore                      # restore latest
sudo vhoster restore /var/lib/vhoster/backups/hosts.1234567890.bak
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--hosts <path>` | `/etc/hosts` | Target hosts file (use a test file to skip root) |
| `--backup-dir <dir>` | `/var/lib/vhoster/backups` | Backup directory |
| `-y` / `--yes` | `false` | Skip confirmation prompts |
| `--dry-run` | `false` | Print what would be done without writing |
| `--force` | `false` | Override shadowed-host checks |

## How it works

- **Managed block** — vhoster only touches lines between `# BEGIN vhoster` and `# END vhoster`. Everything outside is preserved, including your `127.0.0.1 localhost` and any other custom entries.
- **Atomic writes** — writes to a temp file in the same directory, `fsync`s it, then `rename()`s over the target. A crash mid-write cannot corrupt `/etc/hosts`.
- **Backups** — every mutation saves a timestamped copy to the backup directory. The last 10 backups are kept automatically.
- **Locking** — `flock`-based mutual exclusion prevents concurrent `sudo vhoster` calls from racing.
- **Validation** — IPs are validated via `net.ParseIP` (unspecified addresses rejected), hostnames per RFC 1123. Duplicates and unmanaged shadowed names are refused.

## Development

```bash
make test    # run tests
make vet     # run go vet
make build   # build the binary
make lint    # run golangci-lint (if installed)
```

## License

MIT — see [LICENSE](LICENSE).
