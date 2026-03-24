package test

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

// BuildResult captures the outcome of a single test build.
type BuildResult struct {
	Project      string        `json:"project"`
	PromptSet    string        `json:"prompt_set"` // "old" or "new"
	Shipped      bool          `json:"shipped"`
	HasTests     bool          `json:"has_tests"`
	HasReadme    bool          `json:"has_readme"`
	CompileOK    bool          `json:"compile_ok"`
	Duration     time.Duration `json:"duration"`
	ErrorSnippet string        `json:"error_snippet,omitempty"`
}

// ABResult holds the full A/B test comparison.
type ABResult struct {
	OldResults []BuildResult `json:"old_results"`
	NewResults []BuildResult `json:"new_results"`
	OldShipRate float64      `json:"old_ship_rate"`
	NewShipRate float64      `json:"new_ship_rate"`
	Improved    bool         `json:"improved"`
}

// String returns a human-readable comparison.
func (r *ABResult) String() string {
	var sb strings.Builder
	sb.WriteString("=== A/B Test Results ===\n\n")

	sb.WriteString("Old Prompts:\n")
	for _, br := range r.OldResults {
		status := "SHIPPED"
		if !br.Shipped {
			status = "FAILED"
		}
		sb.WriteString(fmt.Sprintf("  %-30s %s  (compile=%v tests=%v readme=%v)\n",
			br.Project, status, br.CompileOK, br.HasTests, br.HasReadme))
	}
	sb.WriteString(fmt.Sprintf("  Ship rate: %.1f%%\n\n", r.OldShipRate*100))

	sb.WriteString("New Prompts:\n")
	for _, br := range r.NewResults {
		status := "SHIPPED"
		if !br.Shipped {
			status = "FAILED"
		}
		sb.WriteString(fmt.Sprintf("  %-30s %s  (compile=%v tests=%v readme=%v)\n",
			br.Project, status, br.CompileOK, br.HasTests, br.HasReadme))
	}
	sb.WriteString(fmt.Sprintf("  Ship rate: %.1f%%\n\n", r.NewShipRate*100))

	if r.Improved {
		sb.WriteString("VERDICT: New prompts are BETTER. Safe to deploy.\n")
	} else {
		sb.WriteString("VERDICT: New prompts are NOT better. Do not deploy.\n")
	}

	return sb.String()
}

// Tester runs A/B tests comparing old and new prompts.
type Tester struct {
	cfg *config.Config
}

// New creates a new Tester.
func New(cfg *config.Config) *Tester {
	return &Tester{cfg: cfg}
}

// SampleProjects are test project ideas used for A/B testing.
var SampleProjects = []string{
	"url-shortener",
	"log-aggregator",
	"health-checker",
	"config-merger",
	"port-scanner",
	"file-hasher",
}

// Run executes the A/B test with the given number of projects per set.
func (t *Tester) Run(count int) (*ABResult, error) {
	if count <= 0 {
		count = 3
	}
	if count > len(SampleProjects)/2 {
		count = len(SampleProjects) / 2
	}

	// Load old prompts from factory.
	promptsDir := filepath.Join(t.cfg.FactoryRepoPath, t.cfg.PromptsDir)
	oldPrompts, err := evolve.ReadPromptTemplates(promptsDir)
	if err != nil {
		return nil, fmt.Errorf("read old prompts: %w", err)
	}

	// Load evolved prompts.
	newPrompts, err := evolve.ReadPromptTemplates(t.cfg.EvolvedDir)
	if err != nil {
		return nil, fmt.Errorf("read evolved prompts: %w", err)
	}

	if len(oldPrompts) == 0 || len(newPrompts) == 0 {
		return nil, fmt.Errorf("need both old and new prompts for A/B test (old=%d, new=%d)", len(oldPrompts), len(newPrompts))
	}

	result := &ABResult{}

	// Test old prompts.
	for i := 0; i < count; i++ {
		project := SampleProjects[i]
		br := t.buildProject(project, "old", oldPrompts)
		result.OldResults = append(result.OldResults, br)
	}

	// Test new prompts.
	for i := 0; i < count; i++ {
		project := SampleProjects[count+i]
		br := t.buildProject(project, "new", newPrompts)
		result.NewResults = append(result.NewResults, br)
	}

	// Compute ship rates.
	result.OldShipRate = shipRate(result.OldResults)
	result.NewShipRate = shipRate(result.NewResults)
	result.Improved = result.NewShipRate > result.OldShipRate

	return result, nil
}

func (t *Tester) buildProject(project, promptSet string, prompts []evolve.PromptTemplate) BuildResult {
	br := BuildResult{
		Project:   project,
		PromptSet: promptSet,
	}

	start := time.Now()

	// Create a temp directory for the build.
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-test-*")
	if err != nil {
		br.ErrorSnippet = fmt.Sprintf("create temp dir: %v", err)
		return br
	}
	defer os.RemoveAll(tmpDir)

	buildDir := filepath.Join(tmpDir, project)
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		br.ErrorSnippet = fmt.Sprintf("create build dir: %v", err)
		return br
	}

	// Write prompt templates to temp location.
	promptDir := filepath.Join(tmpDir, "prompts")
	for _, p := range prompts {
		if err := evolve.WritePromptTemplate(promptDir, p); err != nil {
			br.ErrorSnippet = fmt.Sprintf("write prompt: %v", err)
			return br
		}
	}

	// Invoke Claude to build the project.
	buildPrompt := fmt.Sprintf("Build a Go CLI tool called %s in %s. Use the coding standards from the factory prompts. Include tests, README, proper go.mod.", project, buildDir)
	cmd := exec.Command(t.cfg.ClaudeBinary, "--print", "-p", buildPrompt)
	cmd.Dir = buildDir
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		br.ErrorSnippet = fmt.Sprintf("claude build: %v", err)
		br.Duration = time.Since(start)
		return br
	}

	_ = output // Claude writes files directly.
	br.Duration = time.Since(start)

	// Evaluate the build.
	br.HasReadme = fileExists(filepath.Join(buildDir, "README.md"))
	br.HasTests = hasTestFiles(buildDir)
	br.CompileOK = tryCompile(buildDir)
	br.Shipped = br.CompileOK && br.HasTests && br.HasReadme

	return br
}

func shipRate(results []BuildResult) float64 {
	if len(results) == 0 {
		return 0
	}
	shipped := 0
	for _, r := range results {
		if r.Shipped {
			shipped++
		}
	}
	return float64(shipped) / float64(len(results))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasTestFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_test.go") {
			return true
		}
	}
	// Also check subdirectories one level deep.
	for _, e := range entries {
		if e.IsDir() {
			subEntries, err := os.ReadDir(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, se := range subEntries {
				if strings.HasSuffix(se.Name(), "_test.go") {
					return true
				}
			}
		}
	}
	return false
}

func tryCompile(dir string) bool {
	// Check if there's a go.mod.
	if !fileExists(filepath.Join(dir, "go.mod")) {
		return false
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	return cmd.Run() == nil
}
