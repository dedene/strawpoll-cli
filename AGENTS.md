# Repository Guidelines

## Project Structure

- `cmd/strawpoll/`: CLI entrypoint
- `internal/`: implementation
  - `cmd/`: command routing (kong CLI framework)
  - `api/`: StrawPoll v3 API client (rate limiter, retry logic)
  - `auth/`: API key management (keyring)
  - `config/`: YAML config (poll defaults, keyring settings)
  - `output/`: formatters (table, JSON, TSV)
  - `tui/`: Bubbletea/huh interactive wizards
- `bin/`: build outputs

**Note**: Embeds timezone data for consistent time handling.

## Build, Test, and Development Commands

- `make build`: compile to `bin/strawpoll`
- `make fmt` / `make lint` / `make test` / `make ci`: format, lint, test, full local gate
- `make tools`: install pinned dev tools into `.tools/`
- `make clean`: remove bin/ and .tools/

## Coding Style & Naming Conventions

- Formatting: `make fmt` (goimports local prefix `github.com/dedene/strawpoll-cli` + gofumpt)
- Output: keep stdout parseable (`--json`, `--plain` for TSV); send human hints/progress to stderr
- Linting: golangci-lint v2.8.0 with project config
- TUI: use Charmbracelet ecosystem (bubbletea + huh for forms)

## Testing Guidelines

- Unit tests: stdlib `testing`
- 14 test files; comprehensive coverage:
  - API (clients, polls, URL parsing, rate limiting, transport)
  - Config (paths, config)
  - Auth (keyring)
  - Output (formatting, tables)
  - TUI (terminal detection, wizards)
  - Commands (version, exit)
- CI gate: fmt-check, lint, test

## Config & Secrets

- **Keyring**: 99designs/keyring for API key storage
- **Env var**: `STRAWPOLL_API_KEY` as alternative
- **Config file**: `~/.config/strawpoll/config.yaml`
  - Poll defaults: `dupcheck`, `results_visibility`, `privacy`, `comments`, `vpn`, `participants`, `edit_perms`
  - `keyring_backend`: backend preference

## Key Commands

- `poll`: create/get/list/update/delete polls; results; reset votes
- `meeting`: meeting polls (availability scheduling)
- `ranking`: ranked-choice polls
- Interactive wizards for poll/meeting creation
- Global flags: `--json`, `--plain` (TSV), `--no-color`, `--copy`, `--open`

## API Features

- Rate limiter: 10 req/sec token bucket
- Retry logic: exponential backoff (3 retries)
- Poll types: `multiple_choice`, `meeting`, `ranking`
- Max 30 options per poll

## Commit & Pull Request Guidelines

- Conventional Commits: `feat|fix|refactor|build|ci|chore|docs|style|perf|test`
- Group related changes; avoid bundling unrelated refactors
- PR review: use `gh pr view` / `gh pr diff`; don't switch branches

## Security Tips

- Never commit API keys
- Prefer OS keychain; env var for CI/headless
- Rate limiting protects against API abuse
