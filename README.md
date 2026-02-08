# 🏦 ponto-cli - Access Ponto from the terminal

[![CI](https://github.com/dedene/ponto-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dedene/ponto-cli/actions/workflows/ci.yml)
[![Go 1.23+](https://img.shields.io/badge/go-1.23+-00ADD8.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/dedene/ponto-cli)](https://goreportcard.com/report/github.com/dedene/ponto-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

A command-line interface for the [Ponto](https://myponto.com) banking API.

## Features

- Account information (list, get, balances)
- Transaction history with CSV/JSON export
- Synchronization management
- Pending transactions
- Financial institutions listing
- Multiple profile support (sandbox/live)
- Command allowlist for restricted environments
- Cross-platform (macOS, Linux, Windows)
- OS keyring credential storage

## Installation

### From Source

```bash
go install github.com/dedene/ponto-cli/cmd/ponto@latest
```

### From Releases

Download the binary for your platform from
[GitHub Releases](https://github.com/dedene/ponto-cli/releases).

## Quick Start

```bash
# Login with Ponto credentials (from dashboard)
ponto auth login

# Check authentication status
ponto auth status

# List accounts
ponto accounts list

# Set default account (optional, skips --account-id on future commands)
ponto config set account-id <ACCOUNT_ID>

# List transactions (last 30 days)
ponto transactions list --since=-30d

# Filter by type (income/expense)
ponto transactions list --type=income
ponto transactions list --type=expense

# Export transactions as CSV
ponto transactions export --format=csv > transactions.csv

# Trigger account sync
ponto sync create --subtype=accountTransactions
```

## Commands

```
ponto auth login           Store credentials in keyring
ponto auth logout          Remove credentials from keyring
ponto auth status          Show authentication status

ponto accounts list        List all accounts
ponto accounts get <ID>    Get account details
ponto accounts sync <ID>   Trigger synchronization

ponto transactions list    List transactions (--type=income|expense|all)
ponto transactions get     Get transaction details
ponto transactions export  Export transactions (--type=income|expense|all)

ponto sync create          Create synchronization
ponto sync get             Get sync status
ponto sync list            List synchronizations

ponto pending-transactions list    List pending transactions
ponto financial-institutions list  List financial institutions
ponto organization show            Show organization info

ponto config set <key> <value>     Set configuration value
ponto config get <key>             Get configuration value
```

## Output Formats

```bash
# Table (default) - human readable
ponto accounts list

# JSON - for scripting
ponto accounts list --json

# CSV - for spreadsheets
ponto accounts list --csv

# Plain TSV - for cut/awk
ponto accounts list --plain
```

## Profiles

Use profiles to manage multiple environments:

```bash
# Login to sandbox
ponto auth login --profile=sandbox

# Use sandbox profile
ponto --profile=sandbox accounts list

# Or use shorthand
ponto --sandbox accounts list
```

## Command Allowlist

Restrict available commands in sensitive environments:

```bash
# Only allow read commands
export PONTO_ENABLE_COMMANDS="auth.status,accounts.list,transactions.list"

# Or via flag
ponto --enable-commands=accounts.list accounts list
```

## Environment Variables

| Variable                 | Description                          |
| ------------------------ | ------------------------------------ |
| `PONTO_PROFILE`          | Default profile name                 |
| `PONTO_ENABLE_COMMANDS`  | Comma-separated allowed commands     |
| `PONTO_KEYRING_BACKEND`  | Keyring backend (auto/keychain/file) |
| `PONTO_KEYRING_PASSWORD` | Password for file backend            |

## Configuration

Config file: `~/.config/ponto/config.yaml`

```yaml
default_profile: live
keyring_backend: auto
profiles:
  default:
    account_id: abc-123-def # Default account for commands
  sandbox:
    account_id: sandbox-456
```

### Default Account

Set a default account to avoid specifying `--account-id` on every command:

```bash
# Set default account for current profile
ponto config set account-id <ACCOUNT_ID>

# View current setting
ponto config get account-id

# Now these work without --account-id:
ponto transactions list --since=-30d
ponto pending-transactions list
```

**Resolution order:** flag → config → auto-detect (if single account)

## Agent Skill

This CLI is available as an [open agent skill](https://skills.sh/) for AI assistants including [Claude Code](https://claude.ai/code), [OpenClaw](https://openclaw.ai/), Cursor, and GitHub Copilot:

```bash
npx skills add dedene/ponto-cli
```

## License

MIT
