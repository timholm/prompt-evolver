package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/timholm/prompt-evolver/internal/config"
	"github.com/timholm/prompt-evolver/internal/evolve"
)

// Deployer copies evolved prompts to the factory repo and commits them.
type Deployer struct {
	cfg *config.Config
}

// New creates a new Deployer.
func New(cfg *config.Config) *Deployer {
	return &Deployer{cfg: cfg}
}

// Run deploys evolved prompts to the factory repository.
// If force is false, it checks that evolved prompts exist and contain content.
func (d *Deployer) Run(force bool) error {
	// Read evolved prompts.
	evolved, err := evolve.ReadPromptTemplates(d.cfg.EvolvedDir)
	if err != nil {
		return fmt.Errorf("read evolved prompts: %w", err)
	}

	if len(evolved) == 0 {
		return fmt.Errorf("no evolved prompts found in %s (run 'evolve' first)", d.cfg.EvolvedDir)
	}

	// Validate evolved prompts have content.
	for _, tmpl := range evolved {
		if strings.TrimSpace(tmpl.Content) == "" {
			if !force {
				return fmt.Errorf("evolved prompt %s is empty (use --force to override)", tmpl.Name)
			}
		}
	}

	// Copy to factory prompts directory.
	destDir := filepath.Join(d.cfg.FactoryRepoPath, d.cfg.PromptsDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create prompts dir: %w", err)
	}

	for _, tmpl := range evolved {
		destPath := filepath.Join(destDir, tmpl.Name)

		// Back up the original.
		if data, err := os.ReadFile(destPath); err == nil {
			backupPath := destPath + ".bak"
			if err := os.WriteFile(backupPath, data, 0o644); err != nil {
				return fmt.Errorf("backup %s: %w", tmpl.Name, err)
			}
		}

		if err := os.WriteFile(destPath, []byte(tmpl.Content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", tmpl.Name, err)
		}
		fmt.Printf("  Deployed: %s\n", tmpl.Name)
	}

	// Git commit in the factory repo.
	if err := d.gitCommit(evolved); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

func (d *Deployer) gitCommit(evolved []evolve.PromptTemplate) error {
	repoPath := d.cfg.FactoryRepoPath

	// Check if it's a git repo.
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("factory repo is not a git repo: %w", err)
	}

	// Stage the changed prompt files.
	promptsDir := filepath.Join(d.cfg.PromptsDir)
	for _, tmpl := range evolved {
		path := filepath.Join(promptsDir, tmpl.Name)
		cmd := exec.Command("git", "add", path)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git add %s: %w", tmpl.Name, err)
		}
	}

	// Build a list of changed prompt names.
	var names []string
	for _, tmpl := range evolved {
		names = append(names, tmpl.Name)
	}

	commitMsg := fmt.Sprintf("prompt-evolver: update prompts (%s)\n\nEvolved at %s based on build outcome analysis.\nUpdated: %s",
		time.Now().Format("2006-01-02"),
		time.Now().Format(time.RFC3339),
		strings.Join(names, ", "),
	)

	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
