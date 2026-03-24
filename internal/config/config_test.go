package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars that would override defaults.
	for _, key := range []string{"FACTORY_DATA_DIR", "FACTORY_REPO_PATH", "CLAUDE_BINARY", "PROMPTS_DIR", "EVOLVED_DIR"} {
		t.Setenv(key, "")
	}

	cfg := Load()

	if cfg.ClaudeBinary != "claude" {
		t.Errorf("ClaudeBinary = %q, want %q", cfg.ClaudeBinary, "claude")
	}
	if cfg.PromptsDir != "prompts" {
		t.Errorf("PromptsDir = %q, want %q", cfg.PromptsDir, "prompts")
	}
	if cfg.FactoryDataDir == "" {
		t.Error("FactoryDataDir should not be empty")
	}
	if cfg.FactoryRepoPath == "" {
		t.Error("FactoryRepoPath should not be empty")
	}
	if cfg.EvolvedDir == "" {
		t.Error("EvolvedDir should not be empty")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("FACTORY_DATA_DIR", "/custom/data")
	t.Setenv("FACTORY_REPO_PATH", "/custom/repo")
	t.Setenv("CLAUDE_BINARY", "/usr/local/bin/claude")
	t.Setenv("PROMPTS_DIR", "my-prompts")
	t.Setenv("EVOLVED_DIR", "/tmp/evolved")

	cfg := Load()

	if cfg.FactoryDataDir != "/custom/data" {
		t.Errorf("FactoryDataDir = %q, want %q", cfg.FactoryDataDir, "/custom/data")
	}
	if cfg.FactoryRepoPath != "/custom/repo" {
		t.Errorf("FactoryRepoPath = %q, want %q", cfg.FactoryRepoPath, "/custom/repo")
	}
	if cfg.ClaudeBinary != "/usr/local/bin/claude" {
		t.Errorf("ClaudeBinary = %q, want %q", cfg.ClaudeBinary, "/usr/local/bin/claude")
	}
	if cfg.PromptsDir != "my-prompts" {
		t.Errorf("PromptsDir = %q, want %q", cfg.PromptsDir, "my-prompts")
	}
	if cfg.EvolvedDir != "/tmp/evolved" {
		t.Errorf("EvolvedDir = %q, want %q", cfg.EvolvedDir, "/tmp/evolved")
	}
}

func TestDBPath(t *testing.T) {
	cfg := &Config{FactoryDataDir: "/data/factory"}
	want := filepath.Join("/data/factory", "registry.db")
	if got := cfg.DBPath(); got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

func TestPromptTemplatePath(t *testing.T) {
	cfg := &Config{
		FactoryRepoPath: "/repo",
		PromptsDir:      "prompts",
	}
	want := filepath.Join("/repo", "prompts", "build.md.tmpl")
	if got := cfg.PromptTemplatePath("build.md.tmpl"); got != want {
		t.Errorf("PromptTemplatePath() = %q, want %q", got, want)
	}
}

func TestEvolvedPromptPath(t *testing.T) {
	cfg := &Config{EvolvedDir: "/tmp/evolved"}
	want := filepath.Join("/tmp/evolved", "build.md.tmpl")
	if got := cfg.EvolvedPromptPath("build.md.tmpl"); got != want {
		t.Errorf("EvolvedPromptPath() = %q, want %q", got, want)
	}
}

func TestEnvOr(t *testing.T) {
	key := "PROMPT_EVOLVER_TEST_KEY_12345"
	os.Unsetenv(key)

	if got := envOr(key, "fallback"); got != "fallback" {
		t.Errorf("envOr unset = %q, want %q", got, "fallback")
	}

	t.Setenv(key, "override")
	if got := envOr(key, "fallback"); got != "override" {
		t.Errorf("envOr set = %q, want %q", got, "override")
	}
}
