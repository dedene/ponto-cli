---
name: ponto-cli
description: >
  Access Ponto banking data via the ponto CLI. Use when the user wants to list bank accounts,
  view transactions, export transaction history to CSV/JSON, check account balances, trigger
  bank syncs, view pending transactions, or automate banking data workflows. Triggered by
  mentions of Ponto, bank accounts, transactions, banking API, or financial data export.
---

# ponto-cli

Command-line interface for the [Ponto](https://myponto.com) banking API. Access account data, transactions, and sync status.

## Quick Start

```bash
# Verify auth
ponto auth status

# List accounts
ponto accounts list --json

# List recent transactions
ponto transactions list --since=-30d --json
```

## Authentication

Ponto uses OAuth2 with client credentials from the Ponto dashboard. The user must run `ponto auth login` interactively to store credentials in the system keyring. Do not attempt auth setup on behalf of the user.

```bash
# Check if authenticated
ponto auth status

# Credentials stored in OS keyring
# Config at ~/.config/ponto/config.yaml
```

## Core Rules

1. **Always use `--json`** when parsing output. Table format is for display only.
2. **Read before write** -- fetch current state before triggering syncs.
3. **Set default account** -- use `ponto config set account-id <ID>` to avoid `--account-id` on every command.
4. **Pipe with jq** -- extract IDs/fields: `ponto accounts list --json | jq -r '.[].id'`
5. **Multi-profile** -- use `--profile=sandbox` or `--sandbox` for sandbox environment.

## Output Formats

| Flag | Format | Use case |
|------|--------|----------|
| (default) | Table | User-facing display |
| `--json` | JSON | Agent parsing, scripting |
| `--csv` | CSV | Spreadsheet export |
| `--plain` | TSV | Pipe to awk/cut |

## Workflows

### List Accounts and Set Default

```bash
# Get all accounts
ponto accounts list --json

# Set default to avoid --account-id on future commands
ponto config set account-id <ACCOUNT_ID>

# Verify setting
ponto config get account-id
```

### View Transactions

```bash
# Last 30 days
ponto transactions list --since=-30d --json

# Filter by type
ponto transactions list --type=income --json
ponto transactions list --type=expense --json

# Get single transaction
ponto transactions get <TRANSACTION_ID> --json
```

### Export Transactions

```bash
# Export to CSV
ponto transactions export --format=csv > transactions.csv

# Export with date filter
ponto transactions export --since=-90d --format=csv > q1.csv

# Export as JSON
ponto transactions export --format=json > transactions.json
```

### Trigger Bank Sync

```bash
# Create synchronization for account transactions
ponto sync create --subtype=accountTransactions

# Check sync status
ponto sync get <SYNC_ID> --json

# List recent syncs
ponto sync list --json
```

### Pending Transactions

```bash
# List pending (not yet cleared) transactions
ponto pending-transactions list --json
```

### Financial Institutions

```bash
# List supported banks
ponto financial-institutions list --json
```

### Organization Info

```bash
# Show Ponto organization details
ponto organization show --json
```

## Scripting Examples

```bash
# Get account ID
ACCOUNT=$(ponto accounts list --json | jq -r '.[0].id')

# Export income only
ponto transactions list --type=income --json | jq -r '.[] | [.date, .amount, .counterpartName] | @tsv'

# Sum expenses last 30 days
ponto transactions list --type=expense --since=-30d --json | jq '[.[].amount] | add'
```

## Profiles

```bash
# Login to sandbox environment
ponto auth login --profile=sandbox

# Use sandbox profile
ponto --profile=sandbox accounts list
ponto --sandbox accounts list  # shorthand

# Set profile via env
export PONTO_PROFILE=sandbox
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PONTO_PROFILE` | Default profile name |
| `PONTO_ENABLE_COMMANDS` | Comma-separated allowed commands |
| `PONTO_KEYRING_BACKEND` | Keyring backend (auto/keychain/file) |

## Command Reference

| Command | Description |
|---------|-------------|
| `auth login` | Store credentials in keyring |
| `auth logout` | Remove credentials |
| `auth status` | Show auth status |
| `accounts list` | List all accounts |
| `accounts get <ID>` | Get account details |
| `accounts sync <ID>` | Trigger account sync |
| `transactions list` | List transactions |
| `transactions get <ID>` | Get transaction details |
| `transactions export` | Export transactions |
| `sync create` | Create synchronization |
| `sync get <ID>` | Get sync status |
| `sync list` | List synchronizations |
| `pending-transactions list` | List pending transactions |
| `financial-institutions list` | List banks |
| `organization show` | Show organization info |
| `config set <key> <value>` | Set config value |
| `config get <key>` | Get config value |

## Guidelines

- Never expose or log OAuth credentials or keyring contents.
- Banking data is sensitive -- avoid logging transaction details unnecessarily.
- Syncs may take time to complete -- poll status if needed.
- Rate limits apply -- the CLI handles retries automatically.


## Installation

```bash
brew install dedene/tap/ponto-cli
```
