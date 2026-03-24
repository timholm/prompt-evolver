package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timholm/prompt-evolver/internal/config"
	"github.com/timholm/prompt-evolver/internal/evolve"
)

func TestDeployerRunNoEvolvedPrompts(t *testing.T) {
	dir := t.TempDir()
	evolvedDir := filepath.Join(dir, "evolved")
	os.MkdirAll(evolvedDir, 0o755)

	cfg := &config.Config{
		EvolvedDir: evolvedDir,
	}

	d := New(cfg)
	err := d.Run(false)
	if err == nil {
		t.Error("expected error when no evolved prompts exist")
	}
}

func TestDeployerRunEmptyPromptNoForce(t *testing.T) {
	dir := t.TempDir()
	evolvedDir := filepath.Join(dir, "evolved")
	repoDir := filepath.Join(dir, "repo")
	os.MkdirAll(evolvedDir, 0o755)
	os.MkdirAll(repoDir, 0o755)

	evolve.WritePromptTemplate(evolvedDir, evolve.PromptTemplate{
		Name:    "build.md.tmpl",
		Content: "   ",
	})

	cfg := &config.Config{
		EvolvedDir:      evolvedDir,
		FactoryRepoPath: repoDir,
		PromptsDir:      "prompts",
	}

	d := New(cfg)
	err := d.Run(false)
	if err == nil {
		t.Error("expected error for empty evolved prompt without force")
	}
}

func TestDeployerRunEmptyPromptWithForce(t *testing.T) {
	dir := t.TempDir()
	evolvedDir := filepath.Join(dir, "evolved")
	repoDir := filepath.Join(dir, "repo")
	os.MkdirAll(evolvedDir, 0o755)
	os.MkdirAll(repoDir, 0o755)

	evolve.WritePromptTemplate(evolvedDir, evolve.PromptTemplate{
		Name:    "build.md.tmpl",
		Content: "   ",
	})

	cfg := &config.Config{
		EvolvedDir:      evolvedDir,
		FactoryRepoPath: repoDir,
		PromptsDir:      "prompts",
	}

	d := New(cfg)
	err := d.Run(true)
	if err == nil {
		t.Error("expected error (git commit should fail on non-git dir)")
	}
}

func TestDeployerCopiesFiles(t *testing.T) {
	dir := t.TempDir()
	evolvedDir := filepath.Join(dir, "evolved")
	repoDir := filepath.Join(dir, "repo")
	promptsDir := filepath.Join(repoDir, "prompts")
	os.MkdirAll(evolvedDir, 0o755)
	os.MkdirAll(promptsDir, 0o755)

	os.WriteFile(filepath.Join(promptsDir, "build.md.tmpl"), []byte("original"), 0o644)

	evolve.WritePromptTemplate(evolvedDir, evolve.PromptTemplate{
		Name:    "build.md.tmpl",
		Content: "evolved content with improvements",
	})

	cfg := &config.Config{
		EvolvedDir:      evolvedDir,
		FactoryRepoPath: repoDir,
		PromptsDir:      "prompts",
	}

	d := New(cfg)
	_ = d.Run(false)

	data, err := os.ReadFile(filepath.Join(promptsDir, "build.md.tmpl"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "evolved content with improvements" {
		t.Errorf("deployed content = %q, want evolved content", string(data))
	}

	backup, err := os.ReadFile(filepath.Join(promptsDir, "build.md.tmpl.bak"))
	if err != nil {
		t.Fatalf("ReadFile backup: %v", err)
	}
	if string(backup) != "original" {
		t.Errorf("backup content = %q, want original", string(backup))
	}
}

func TestNewDeployer(t *testing.T) {
	cfg := &config.Config{
		EvolvedDir:      "/tmp/evolved",
		FactoryRepoPath: "/tmp/repo",
		PromptsDir:      "prompts",
	}
	d := New(cfg)
	if d == nil {
		t.Fatal("New returned nil")
	}
	if d.cfg != cfg {
		t.Error("deployer cfg mismatch")
	}
}
