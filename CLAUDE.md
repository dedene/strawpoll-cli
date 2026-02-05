# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Dev Commands

```bash
make build              # Build binary to bin/strawpoll
make strawpoll -- poll get <id>  # Build and run with args
make test               # go test ./...
make lint               # golangci-lint v2 (config: .golangci.yml, version: "2")
make fmt                # gofumpt + goimports (auto-installs to .tools/)
make ci                 # fmt-check + lint + test (matches GitHub Actions)
make tools              # Install gofumpt, goimports, golangci-lint to .tools/
```

Single test: `go test ./internal/api/... -run TestParseURL -v`

CI runs lint and `go test ./... -v -race` on ubuntu-latest.

## Architecture

CLI for [StrawPoll API v3](https://api.strawpoll.com/v3). Three poll types: multiple_choice, meeting, ranking.

**Entry:** `cmd/strawpoll/main.go` → `internal/cmd.Execute()` → Kong parser

**Packages:**
- `internal/cmd` — Kong command structs. Each command is a struct with `Run(flags *RootFlags) error`. CLI struct in `root.go` embeds all subcommands (Poll, Meeting, Ranking, Auth, Config, Completion).
- `internal/api` — HTTP client with retry transport (3 retries, exponential backoff, Retry-After), token bucket rate limiter (10 req/s), typed request/response structs in `types.go`. Auth via `X-API-Key` header.
- `internal/auth` — Keyring-based API key storage. Resolution order: `STRAWPOLL_API_KEY` env → system keyring. Linux D-Bus timeout (5s) with file backend fallback.
- `internal/config` — YAML config at `~/.config/strawpoll/config.yaml`. Holds poll creation defaults (dupcheck, results_visibility, etc).
- `internal/output` — Three output modes: table (colored), JSON, plain TSV. `NewFormatter(w, json, plain, noColor)` pattern used by all commands.
- `internal/tui` — Bubbletea/huh interactive wizards (poll, meeting creation). TUI renders on stderr; data on stdout.

**Auth flow:** `auth.GetAPIKey()` checks env var first, then opens keyring → `api.NewClient(apiKey)`. Commands call `newClientFromAuth()` helper in `helpers.go`.

**Command pattern:** Each command struct (e.g., `PollCreateCmd`) has Kong tags for args/flags, a `Run(*RootFlags) error` method, and uses `output.NewFormatter` for output. Create commands support both flag-based and interactive (wizard) paths.

## API Quirks

- `results_visibility: "hidden"` not `"never"` (API doc bug)
- `type: "ranking"` not `"ranked_choice"` (API doc bug)
- Max 30 poll options — validated client-side
- URL input accepts full URLs or bare IDs (parsed in `api/url.go`)
- No vote endpoint — voting is browser-only
- `PollResults.VoteCount` and `ParticipantCount` are camelCase in JSON (OpenAPI spec)

## Version Injection

ldflags: `-X github.com/dedene/strawpoll-cli/internal/cmd.version=...` (also commit, date)

## Release

goreleaser v2 config in `.goreleaser.yaml`. Homebrew tap at `dedene/homebrew-tap` (needs `HOMEBREW_TAP_GITHUB_TOKEN`).
