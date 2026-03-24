# prompt-evolver

Autonomous prompt evolution for [claude-code-factory](https://github.com/timholm/claude-code-factory). Analyzes shipped vs failed builds, identifies prompt weaknesses, generates improved prompts, and A/B tests them.

## How it works

prompt-evolver is a closed-loop optimization system for the factory's build, SEO, and review prompts:

1. **ANALYZE** -- Reads all build outcomes from the factory's SQLite registry. Groups failures by pattern (no tests, wrong module path, compilation errors, etc.) and computes ship rate.

2. **EVOLVE** -- Sends the failure analysis and current prompt templates to Claude Opus. Claude generates improved prompts that specifically address the top failure patterns.

3. **TEST** -- Builds sample projects with old prompts and new prompts side by side. Compares ship rate, test coverage, README quality, and compilation success.

4. **DEPLOY** -- If new prompts outperform old ones, writes them to the factory repo and commits.

## Install

```bash
go install github.com/timholm/prompt-evolver@latest
```

Or build from source:

```bash
git clone https://github.com/timholm/prompt-evolver.git
cd prompt-evolver
make build
```

## Usage

```bash
# Step 1: Analyze build outcomes
prompt-evolver analyze

# Step 2: Generate improved prompts
prompt-evolver evolve

# Step 3: A/B test old vs new
prompt-evolver test --count 3

# Step 4: Deploy if improved
prompt-evolver deploy
```

Full pipeline:

```bash
prompt-evolver analyze && prompt-evolver evolve && prompt-evolver test && prompt-evolver deploy
```

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|---|---|---|
| `FACTORY_DATA_DIR` | `~/claude-code-factory/data` | Path to factory's SQLite registry |
| `FACTORY_REPO_PATH` | `~/claude-code-factory` | Local clone of the factory repo |
| `CLAUDE_BINARY` | `claude` | Path to the Claude CLI |
| `PROMPTS_DIR` | `prompts` | Subdirectory in factory repo containing prompt templates |
| `EVOLVED_DIR` | `~/.prompt-evolver/evolved` | Scratch directory for evolved prompt candidates |

## Architecture

```
internal/
  analyze/    -- Build outcome analysis and pattern extraction
  evolve/     -- Claude-powered prompt generation
  test/       -- A/B testing framework
  deploy/     -- Factory repo deployment
  config/     -- Configuration loading
  db/         -- SQLite registry reader
```

## Requirements

- Go 1.22+
- Claude CLI (`claude`) on PATH
- Access to factory's SQLite registry
- Local clone of claude-code-factory
