# 🍓 strawpoll-cli - StrawPoll in your terminal

A command-line interface for the [StrawPoll](https://strawpoll.com/) API v3 - create and manage polls from the command line.

[![CI](https://github.com/dedene/strawpoll-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dedene/strawpoll-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dedene/strawpoll-cli)](https://goreportcard.com/report/github.com/dedene/strawpoll-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

## Features

- **Create polls** - Multiple-choice polls with 20+ configuration options
- **View results** - Colored ASCII tables with vote counts and percentages
- **Participant breakdown** - See who voted for what with `--participants`
- **Delete polls** - With confirmation prompt (or `--force` for scripting)
- **Multiple output formats** - Human-readable tables, JSON, or plain TSV
- **Clipboard & browser** - Copy poll URL or open it directly after creation
- **Config defaults** - Save your preferred poll settings to avoid repetitive flags

## Installation

### Homebrew (macOS/Linux)

```bash
brew install dedene/tap/strawpoll
```

### Go install

```bash
go install github.com/dedene/strawpoll-cli/cmd/strawpoll@latest
```

### From source

```bash
git clone https://github.com/dedene/strawpoll-cli.git
cd strawpoll-cli
make build
```

## Setup

You need a StrawPoll API key. Get one at [strawpoll.com/account/settings](https://strawpoll.com/account/settings).

```bash
# Store your API key securely in the system keyring
strawpoll auth set-key

# Verify it's configured
strawpoll auth status
```

## Usage

### Create a poll

```bash
# Simple poll
strawpoll poll create "Favorite color?" Red Blue Green Yellow

# With options
strawpoll poll create "Best framework?" React Vue Svelte Angular \
  --dupcheck session \
  --results-vis after_vote \
  --deadline 24h

# Copy URL to clipboard after creation
strawpoll poll create "Lunch?" Pizza Sushi Tacos --copy

# Open in browser
strawpoll poll create "Meeting day?" Monday Wednesday Friday --open
```

### View poll details

```bash
# By poll ID
strawpoll poll get NPgxkzPqrn2

# By full URL (auto-extracts ID)
strawpoll poll get https://strawpoll.com/NPgxkzPqrn2
```

### View results

```bash
# Results table with vote counts and percentages
strawpoll poll results NPgxkzPqrn2

# With per-participant breakdown
strawpoll poll results NPgxkzPqrn2 --participants
```

### Delete a poll

```bash
# With confirmation prompt
strawpoll poll delete NPgxkzPqrn2

# Skip confirmation (for scripting)
strawpoll poll delete NPgxkzPqrn2 --force
```

### Output formats

```bash
# JSON output (for scripting)
strawpoll poll get NPgxkzPqrn2 --json

# Plain TSV output
strawpoll poll results NPgxkzPqrn2 --plain

# Disable colors
strawpoll poll results NPgxkzPqrn2 --no-color
```

### Configuration

```bash
# Set default poll options
strawpoll config set dupcheck session
strawpoll config set results_visibility after_vote

# View current config
strawpoll config show

# Show config file path
strawpoll config path
```

## Shell completions

```bash
# Bash
strawpoll completion bash > /etc/bash_completion.d/strawpoll

# Zsh
strawpoll completion zsh > "${fpath[1]}/_strawpoll"

# Fish
strawpoll completion fish > ~/.config/fish/completions/strawpoll.fish
```

## AI Agent Skill

Install the [Agent Skill](https://agentskills.io/) to use strawpoll-cli with Claude Code, Cursor, Windsurf, or other AI coding agents:

```bash
npx skills add dedene/strawpoll-cli -g
```

## Environment variables

| Variable | Description |
|---|---|
| `STRAWPOLL_API_KEY` | API key (overrides keyring) |
| `STRAWPOLL_KEYRING_BACKEND` | Keyring backend: `keychain`, `file`, `pass` |
| `STRAWPOLL_KEYRING_PASSWORD` | Password for file-based keyring |
| `NO_COLOR` | Disable colored output |

## License

MIT - see [LICENSE](LICENSE)

## Credits

Powered by the [StrawPoll API v3](https://strawpoll.com/).
