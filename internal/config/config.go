package config

import (
	"os"
	"path/filepath"
)

// Config holds all configuration for prompt-evolver.
type Config struct {
	// FactoryDataDir is the directory containing the factory's SQLite registry.
	FactoryDataDir string

	// FactoryRepoPath is the local path to the claude-code-factory repository.
	FactoryRepoPath string

	// ClaudeBinary is the path to the Claude CLI binary.
	ClaudeBinary string

	// PromptsDir is the subdirectory within FactoryRepoPath that holds prompt templates.
	PromptsDir string

	// EvolvedDir is a scratch directory for storing evolved prompt candidates.
	EvolvedDir string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	home, _ := os.UserHomeDir()

	cfg := &Config{
		FactoryDataDir:  envOr("FACTORY_DATA_DIR", filepath.Join(home, "claude-code-factory", "data")),
		FactoryRepoPath: envOr("FACTORY_REPO_PATH", filepath.Join(home, "claude-code-factory")),
		ClaudeBinary:    envOr("CLAUDE_BINARY", "claude"),
		PromptsDir:      envOr("PROMPTS_DIR", "prompts"),
		EvolvedDir:      envOr("EVOLVED_DIR", filepath.Join(home, ".prompt-evolver", "evolved")),
	}

	return cfg
}

// DBPath returns the full path to the factory's SQLite database.
func (c *Config) DBPath() string {
	return filepath.Join(c.FactoryDataDir, "registry.db")
}

// PromptTemplatePath returns the full path to a prompt template in the factory repo.
func (c *Config) PromptTemplatePath(name string) string {
	return filepath.Join(c.FactoryRepoPath, c.PromptsDir, name)
}

// EvolvedPromptPath returns the full path to an evolved prompt candidate.
func (c *Config) EvolvedPromptPath(name string) string {
	return filepath.Join(c.EvolvedDir, name)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
