# ponto

A command-line interface for the [Ponto](https://myponto.com) banking API.

## Status

**Work in progress** - See [SPEC.md](SPEC.md) for planned features.

## Features (Planned)

- Account information (list, get, balances)
- Transaction history with CSV export
- Synchronization management
- Multiple profile support (sandbox/live)
- Command allowlist for restricted environments
- Cross-platform (macOS, Linux, Windows)

## Installation

Coming soon.

## Quick Start

```bash
# Login with Ponto credentials
ponto auth login

# List accounts
ponto accounts list

# Get transaction history
ponto transactions list --account-id=<ID> --since=-30d
```

## License

MIT
