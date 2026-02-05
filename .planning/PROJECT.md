# StrawPoll CLI

## What This Is

A command-line interface for the StrawPoll API v3, enabling poll creation, management, editing, and results viewing from the terminal. Modeled after [gogcli](https://github.com/steipete/gogcli) and Peter's existing Go CLI toolkit (delijn-cli, harvest-cli, frontapp-cli). Built for power users who live in the terminal and want full StrawPoll lifecycle without leaving it.

## Core Value

Create, manage, and view results of StrawPoll polls entirely from the command line — with the same quality and patterns as the existing Go CLI suite.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Auth: store/retrieve/remove API key via system keyring (macOS Keychain, Linux Secret Service, etc.)
- [ ] Auth: STRAWPOLL_API_KEY env var fallback for CI/headless
- [ ] Auth: `auth set-key`, `auth status`, `auth remove` commands
- [ ] Poll: create multiple-choice polls with full config flags (20+ flags for all poll_config fields)
- [ ] Poll: grouped flag help output (Creation, Voting Rules, Privacy & Access, Display & Scheduling)
- [ ] Poll: get poll details by ID or full URL
- [ ] Poll: view results with ASCII table + summary (vote counts, percentages)
- [ ] Poll: view per-participant breakdown with `--participants` flag
- [ ] Poll: delete with confirmation prompt (skippable with `--force`)
- [ ] Poll: update poll title, options, config via `poll update`
- [ ] Poll: reset poll results/deadline via `poll reset`
- [ ] Poll: `poll list` showing user's polls from API (remote, paginated)
- [ ] Meeting: create scheduling polls with combined all-day dates + time ranges in one flow
- [ ] Meeting: bubbletea TUI for interactive date/time selection when no flags given
- [ ] Meeting: inline flags for scripted creation (`--date`, `--range`, `--tz`)
- [ ] Meeting: view results as participant-by-timeslot grid
- [ ] Meeting: update meeting poll via `meeting update`
- [ ] Ranking: create ranking polls
- [ ] Ranking: update ranking poll via `ranking update`
- [ ] Ranking: results showing weighted summary (Borda count) + position breakdown with `--verbose`
- [ ] Config: `config show`, `config set`, `config path` commands
- [ ] Config: default poll creation values in config.yaml (dupcheck, results_visibility, etc.), overridable by flags
- [ ] Output: `--json` (structured), `--plain` (TSV with headers), `--no-color` modes
- [ ] Output: `--copy` flag to copy poll URL to clipboard after creation
- [ ] Output: `--open` flag to open poll in default browser after creation
- [ ] URL parsing: accept both poll IDs (`NPgxkzPqrn2`) and full URLs (`https://strawpoll.com/NPgxkzPqrn2`)
- [ ] Shell completion: bash, zsh, fish
- [ ] Version command with build metadata

### Out of Scope

- Webhooks — complexity not justified for CLI use case; revisit if API adds list/management endpoints
- Voting — no vote endpoint in OpenAPI spec; voting done via browser
- OAuth flow — StrawPoll uses simple API keys, not OAuth
- Result caching — always fetch fresh to avoid stale/privacy issues
- Live API integration tests in CI — unit tests with httptest mocks only, no API key in CI secrets
- QR code generation — nice-to-have but not v1
- Local poll index (polls.yaml) — API has `GET /users/@me/polls` for server-side listing
- Custom design/branding — premium-only, poor CLI UX
- Image/media upload — web UI better for this

## Context

### StrawPoll API v3

- **Base URL:** `https://api.strawpoll.com/v3`
- **Auth:** `X-API-KEY` header
- **Endpoints:**
  - `POST /polls` — create poll (multiple_choice, meeting, ranking types)
  - `GET /polls/{id}` — get poll details
  - `PUT /polls/{id}` — update poll (title, options, config)
  - `GET /polls/{id}/results` — get results with per-participant breakdown
  - `DELETE /polls/{id}/results` — reset poll results/deadline
  - `DELETE /polls/{id}` — permanently delete poll
  - `GET /users/@me/polls` — list user's polls (paginated)
- **Poll types:** `multiple_choice`, `meeting`, `ranking`
- **Poll config fields:** is_private, vote_type, allow_comments, allow_indeterminate, allow_other_option, custom_design_colors, deadline_at, duplication_checking (ip/session/none), allow_vpn_users, edit_vote_permissions, force_appearance, hide_participants, is_multiple_choice, multiple_choice_min/max, number_of_winners, randomize_options, require_voter_names, results_visibility, use_custom_design, send_webhooks
- **Known doc bugs:** `results_visibility: "hidden"` not `"never"`, `type: "ranking"` not `"ranked_choice"` — hardcode corrections
- **OpenAPI spec:** `github.com/strawpoll-com/strawpoll-api-v3/blob/main/openapi/strawpoll_v3.yml`
- **Pricing:** Free / Basic (€8/mo) / Pro (€28/mo) / Business (€52/mo) — API available on all plans
- **Rate limits:** Not documented; implement defensive backoff

### Existing CLI Patterns (Peter's Go CLI Suite)

All of Peter's Go CLIs share a consistent architecture:

- **Framework:** `alecthomas/kong` v1.13.0 with struct-based command definitions
- **Auth:** `99designs/keyring` v1.2.2 for secure credential storage
- **Config:** YAML at `~/.config/<app>/config.yaml` via `go.yaml.in/yaml/v3`
- **Output:** Three modes via RootFlags: `--json`, `--plain`, `--no-color`
- **Error handling:** `ExitError` struct with exit codes (0=OK, 1=Error, 2=Usage)
- **Help:** Custom help printer with better formatting than Kong defaults
- **Completion:** bash/zsh/fish generation
- **Version:** Build-time injection via ldflags
- **Project structure:**
  ```
  cmd/strawpoll/main.go
  internal/cmd/          (CLI commands + root parser)
  internal/config/       (config file handling)
  internal/auth/         (keyring store)
  internal/api/          (HTTP client)
  internal/output/       (formatting: table, JSON, TSV)
  internal/ui/           (bubbletea TUI components)
  ```

### Reference Implementations

- **Auth pattern:** `/Users/peter/Development/Go/delijn-cli/internal/auth/keyring.go` (simple API key keyring)
- **Config pattern:** `/Users/peter/Development/Go/delijn-cli/internal/config/config.go` (YAML config)
- **Help printer:** `/Users/peter/Development/Go/delijn-cli/internal/cmd/help_printer.go`
- **TUI pattern:** `/Users/peter/Development/Go/harvest-cli/internal/ui/` (bubbletea components)
- **Root structure:** `/Users/peter/Development/Go/delijn-cli/internal/cmd/root.go`
- **Exit handling:** `/Users/peter/Development/Go/harvest-cli/internal/cmd/exit.go`

## Constraints

- **Tech stack**: Go 1.25, Kong CLI, keyring, bubbletea, goreleaser — matching existing CLI suite
- **Module path**: `github.com/dedene/strawpoll-cli`
- **Binary name**: `strawpoll`
- **Repo**: `git@github.com:dedene/strawpoll-cli.git`
- **API limitations**: No vote endpoint in OpenAPI spec, undocumented rate limits
- **Enum corrections**: Must use actual API values (`hidden`, `ranking`) not documented ones (`never`, `ranked_choice`)
- **Platform targets**: macOS arm64+amd64, Linux amd64+arm64, Windows amd64, Homebrew tap (`dedene/tap/strawpoll`)
- **CI**: GitHub Actions (golangci-lint + tests + goreleaser)
- **Testing**: Unit tests with httptest mocks only, no live API tests in CI
- **Files**: <500 LOC per file, split/refactor as needed

## Command Tree

```
strawpoll
├── auth
│   ├── set-key          # Store API key in keyring
│   ├── status           # Show whether key is stored
│   └── remove           # Delete stored key
├── poll
│   ├── create           # Create multiple-choice poll (flags or defaults)
│   ├── get <id|url>     # Get poll details
│   ├── update <id|url>  # Update poll title, options, config
│   ├── results <id|url> # View results (table + optional --participants)
│   ├── reset <id|url>   # Reset poll results/deadline (with confirmation)
│   ├── delete <id|url>  # Delete poll (with confirmation)
│   └── list             # List user's polls from API (paginated)
├── meeting
│   ├── create           # Create meeting/scheduling poll (TUI or flags)
│   ├── get <id|url>     # Get meeting details
│   ├── update <id|url>  # Update meeting poll
│   ├── results <id|url> # View participant grid
│   ├── delete <id|url>  # Delete meeting (with confirmation)
│   └── list             # List user's meeting polls
├── ranking
│   ├── create           # Create ranking poll
│   ├── get <id|url>     # Get ranking details
│   ├── update <id|url>  # Update ranking poll
│   ├── results <id|url> # View weighted + position breakdown
│   ├── delete <id|url>  # Delete ranking (with confirmation)
│   └── list             # List user's ranking polls
├── config
│   ├── show             # Display current config
│   ├── set <key> <val>  # Set config value
│   └── path             # Show config file location
├── completion           # Generate shell completions
└── version              # Show version info
```

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Kong CLI framework | Matches all existing Go CLIs, struct-based, good help | — Pending |
| Keyring for API key storage | Secure, cross-platform, proven in delijn-cli | — Pending |
| Split commands (poll/meeting/ranking) | More discoverable than unified with --type flag | — Pending |
| Full flags for all poll config | Power users want full control; sensible defaults from config | — Pending |
| Bubbletea TUI for creation wizards | Proven in harvest-cli/irail-cli, better UX than stdin prompts | — Pending |
| Remote-only poll listing | API has GET /users/@me/polls; no need for local polls.yaml | — Pending |
| Skip voting commands | No vote endpoint in OpenAPI spec; voting via browser | — Pending |
| Poll update via PUT endpoint | Research found PUT /polls/{id} exists; enables editing | — Pending |
| Results reset command | DELETE /polls/{id}/results available; confirm prompt required | — Pending |
| go.yaml.in/yaml/v3 over gopkg.in | gopkg.in/yaml.v3 archived April 2025; fork is API-compatible | — Pending |
| Hardcode enum corrections | Known doc bugs; hardcoding avoids user confusion | — Pending |
| Human-readable flag names | --dupcheck not --duplication-checking; CLI maps internally | — Pending |
| TSV for --plain output | Tab-separated with headers; works for both 1D and 2D result data | — Pending |
| Confirm prompt on delete | Permanent operation; --force for scripting | — Pending |
| URL + ID parsing | Users copy URLs from browser; auto-extract ID for convenience | — Pending |
| Config defaults for poll creation | Saves typing for repeat users; flags override | — Pending |
| Unit tests only | No live API key in CI; httptest mocks sufficient | — Pending |
| All platforms + Homebrew | Maximum reach; goreleaser handles multi-platform | — Pending |

---
*Last updated: 2026-02-05 after research phase — corrected API capabilities, removed voting, added update/reset/remote listing*
