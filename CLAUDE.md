# CLAUDE.md

## Project

prompt-evolver is a Go CLI that optimizes claude-code-factory's prompts through automated analysis, evolution, testing, and deployment. It reads build outcomes from a SQLite registry, identifies failure patterns, uses Claude to generate improved prompts, A/B tests them, and deploys winners.

## Build

```bash
make build    # compile binary
make test     # run all tests
make lint     # go vet
```

## Architecture

- `main.go` -- CLI entry point with cobra commands: analyze, evolve, test, deploy
- `internal/config/` -- env-based configuration (FACTORY_DATA_DIR, FACTORY_REPO_PATH, CLAUDE_BINARY)
- `internal/db/` -- SQLite reader for factory's build registry
- `internal/analyze/` -- Pattern matching on error logs, failure grouping, ship rate computation
- `internal/evolve/` -- Sends analysis + current prompts to Claude, parses improved templates
- `internal/test/` -- A/B testing: builds sample projects with old vs new prompts
- `internal/deploy/` -- Copies evolved prompts to factory repo, creates git commit

## Conventions

- Pure Go, no frameworks except cobra for CLI and go-sqlite3 for DB
- All packages under `internal/` -- no exported API
- Config via environment variables, never hardcoded paths
- Error handling: wrap with context, return up, never panic
- Tests use `_test.go` suffix, test helpers in same package
