package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timholm/prompt-evolver/internal/config"
	"github.com/timholm/prompt-evolver/internal/evolve"
)

func TestDeployerNoEvolvedPrompts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-deploy-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	emptyDir := filepath.Join(tmpDir, "evolved")
	os.MkdirAll(emptyDir, 0o755)

	cfg := &config.Config{
		FactoryRepoPath: filepath.Join(tmpDir, "factory"),
		PromptsDir:      "prompts",
		EvolvedDir:      emptyDir,
	}

	d := New(cfg)
	err = d.Run(false)
	if err == nil {
		t.Error("expected error when no evolved prompts exist")
	}
}

func TestDeployerEmptyPromptContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-deploy-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	evolvedDir := filepath.Join(tmpDir, "evolved")
	os.MkdirAll(evolvedDir, 0o755)

	// Write an empty .tmpl file.
	os.WriteFile(filepath.Join(evolvedDir, "build.md.tmpl"), []byte("   "), 0o644)

	cfg := &config.Config{
		FactoryRepoPath: filepath.Join(tmpDir, "factory"),
		PromptsDir:      "prompts",
		EvolvedDir:      evolvedDir,
	}

	d := New(cfg)
	err = d.Run(false)
	if err == nil {
		t.Error("expected error for empty prompt content without --force")
	}
}

func TestDeployerWritesPrompts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-deploy-write-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up evolved prompts.
	evolvedDir := filepath.Join(tmpDir, "evolved")
	evolve.WritePromptTemplate(evolvedDir, evolve.PromptTemplate{
		Name:    "build.md.tmpl",
		Content: "improved build prompt",
	})
	evolve.WritePromptTemplate(evolvedDir, evolve.PromptTemplate{
		Name:    "seo.md.tmpl",
		Content: "improved seo prompt",
	})

	// Set up factory repo (needs to be a git repo for commit step).
	factoryDir := filepath.Join(tmpDir, "factory")
	promptsDir := filepath.Join(factoryDir, "prompts")
	os.MkdirAll(promptsDir, 0o755)

	// Write original prompts for backup test.
	os.WriteFile(filepath.Join(promptsDir, "build.md.tmpl"), []byte("original build"), 0o644)

	cfg := &config.Config{
		FactoryRepoPath: factoryDir,
		PromptsDir:      "prompts",
		EvolvedDir:      evolvedDir,
	}

	d := New(cfg)
	// Run will fail at git commit (not a git repo), but files should be written.
	_ = d.Run(false)

	// Verify the prompt was written.
	data, err := os.ReadFile(filepath.Join(promptsDir, "build.md.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "improved build prompt" {
		t.Errorf("expected improved content, got %q", string(data))
	}

	// Verify backup was created.
	backup, err := os.ReadFile(filepath.Join(promptsDir, "build.md.tmpl.bak"))
	if err != nil {
		t.Fatal("expected backup file")
	}
	if string(backup) != "original build" {
		t.Errorf("expected original content in backup, got %q", string(backup))
	}
}
