# AGENTS.md

## prompt-evolver

### Purpose
Autonomous prompt evolution for claude-code-factory. Closes the loop between build outcomes and prompt quality.

### Capabilities
- **analyze**: Read factory's SQLite registry, extract failure patterns, compute ship rates
- **evolve**: Send failure analysis to Claude, receive improved prompt templates
- **test**: A/B test old vs new prompts by building sample projects
- **deploy**: Write improved prompts to factory repo, git commit

### Integration
- Reads from: claude-code-factory's `data/registry.db` (SQLite)
- Writes to: claude-code-factory's `prompts/` directory
- Calls: Claude CLI (`claude --print -p`)
- Triggered by: cron, post-build hook, or manual invocation

### Data Flow
```
registry.db --> analyze --> failure_patterns --> evolve --> evolved_prompts
evolved_prompts --> test (A/B) --> results --> deploy (if improved)
```

### Error Patterns Detected
- no_test_files, wrong_module_path, compilation_error, test_failure
- missing_readme, missing_dependency, timeout, permission_denied
- empty_output, lint_failure
