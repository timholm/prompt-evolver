package evolve

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/timholm/prompt-evolver/internal/analyze"
	"github.com/timholm/prompt-evolver/internal/config"
)

// Result holds the outcome of the evolution phase.
type Result struct {
	OldPrompts     []PromptTemplate `json:"-"`
	EvolvedPrompts []PromptTemplate `json:"-"`
	Summary        string           `json:"summary"`
}

// String returns a human-readable result.
func (r *Result) String() string {
	var sb strings.Builder
	sb.WriteString("=== Prompt Evolution Result ===\n")
	sb.WriteString(r.Summary)
	sb.WriteString(fmt.Sprintf("\nEvolved %d prompt templates.\n", len(r.EvolvedPrompts)))
	for _, p := range r.EvolvedPrompts {
		sb.WriteString(fmt.Sprintf("  - %s (%d bytes)\n", p.Name, len(p.Content)))
	}
	return sb.String()
}

// Evolver generates improved prompts by sending analysis + current prompts to Claude.
type Evolver struct {
	cfg *config.Config
}

// New creates a new Evolver.
func New(cfg *config.Config) *Evolver {
	return &Evolver{cfg: cfg}
}

// Run executes the evolution phase.
// If analysisFile is empty, it reads the default analysis output.
func (e *Evolver) Run(analysisFile string) (*Result, error) {
	// Load the analysis report.
	report, err := e.loadAnalysis(analysisFile)
	if err != nil {
		return nil, fmt.Errorf("load analysis: %w", err)
	}

	// Read current prompt templates.
	promptsDir := filepath.Join(e.cfg.FactoryRepoPath, e.cfg.PromptsDir)
	current, err := ReadPromptTemplates(promptsDir)
	if err != nil {
		return nil, fmt.Errorf("read current prompts: %w", err)
	}

	if len(current) == 0 {
		return nil, fmt.Errorf("no prompt templates found in %s", promptsDir)
	}

	// Build the evolution prompt for Claude.
	evolutionPrompt := buildEvolutionPrompt(report, current)

	// Call Claude to generate improved prompts.
	evolved, summary, err := e.callClaude(evolutionPrompt, current)
	if err != nil {
		return nil, fmt.Errorf("call claude: %w", err)
	}

	// Write evolved prompts to scratch directory.
	for _, tmpl := range evolved {
		if err := WritePromptTemplate(e.cfg.EvolvedDir, tmpl); err != nil {
			return nil, fmt.Errorf("write evolved prompt: %w", err)
		}
	}

	return &Result{
		OldPrompts:     current,
		EvolvedPrompts: evolved,
		Summary:        summary,
	}, nil
}

func (e *Evolver) loadAnalysis(path string) (*analyze.Report, error) {
	if path == "" {
		// Run a fresh analysis.
		a := analyze.New(e.cfg)
		return a.Run()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report analyze.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse analysis: %w", err)
	}
	return &report, nil
}

func buildEvolutionPrompt(report *analyze.Report, current []PromptTemplate) string {
	var sb strings.Builder

	sb.WriteString("You are an expert prompt engineer for an autonomous code factory that generates Go repositories.\n\n")
	sb.WriteString("## Current Build Performance\n\n")
	sb.WriteString(fmt.Sprintf("- Total builds: %d\n", report.TotalBuilds))
	sb.WriteString(fmt.Sprintf("- Ship rate: %.1f%%\n", report.ShipRate*100))
	sb.WriteString(fmt.Sprintf("- Failed: %d builds\n\n", report.FailedCount))

	sb.WriteString("## Top Failure Patterns\n\n")
	for _, fg := range report.FailureGroups {
		sb.WriteString(fmt.Sprintf("- **%s** (%d occurrences, %.1f%%): %s\n", fg.Pattern, fg.Count, fg.Percentage*100, fg.Desc))
		for _, ex := range fg.Examples {
			sb.WriteString(fmt.Sprintf("  Example: `%s`\n", ex))
		}
	}

	sb.WriteString("\n## Shipped Build Traits\n\n")
	sb.WriteString(fmt.Sprintf("- Test rate: %.1f%%\n", report.ShippedTraits.TestRate*100))
	sb.WriteString(fmt.Sprintf("- README rate: %.1f%%\n", report.ShippedTraits.ReadmeRate*100))

	sb.WriteString("\n## Current Prompt Templates\n\n")
	for _, tmpl := range current {
		sb.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", tmpl.Name, tmpl.Content))
	}

	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Generate improved versions of each prompt template that address the failure patterns above.\n")
	sb.WriteString("Focus on:\n")
	sb.WriteString("1. Explicit instructions that prevent the top failure patterns\n")
	sb.WriteString("2. Reinforcing patterns that lead to shipped builds (tests, README, correct module path)\n")
	sb.WriteString("3. Adding guard rails and verification steps\n\n")
	sb.WriteString("Output each improved template in a fenced code block with the filename as a header.\n")
	sb.WriteString("Also provide a brief summary of what changed and why.\n")

	return sb.String()
}

// callClaude sends the evolution prompt to Claude and parses the response.
func (e *Evolver) callClaude(prompt string, current []PromptTemplate) ([]PromptTemplate, string, error) {
	cmd := exec.Command(e.cfg.ClaudeBinary, "--print", "-p", prompt)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("claude command: %w", err)
	}

	responseText := string(output)

	// Parse evolved templates from the response.
	evolved := parseEvolvedTemplates(responseText, current)
	if len(evolved) == 0 {
		return nil, "", fmt.Errorf("claude returned no parseable templates")
	}

	// Extract summary (text before the first code block or after the last).
	summary := extractSummary(responseText)

	return evolved, summary, nil
}

// parseEvolvedTemplates extracts template content from Claude's markdown response.
func parseEvolvedTemplates(response string, current []PromptTemplate) []PromptTemplate {
	var evolved []PromptTemplate

	for _, tmpl := range current {
		// Look for the template name followed by a code block.
		nameIdx := strings.Index(response, tmpl.Name)
		if nameIdx == -1 {
			continue
		}

		// Find the next code block after the template name.
		remaining := response[nameIdx:]
		blockStart := strings.Index(remaining, "```")
		if blockStart == -1 {
			continue
		}

		// Skip the opening ``` and any language tag.
		afterFence := remaining[blockStart+3:]
		newlineIdx := strings.Index(afterFence, "\n")
		if newlineIdx == -1 {
			continue
		}
		contentStart := afterFence[newlineIdx+1:]

		// Find the closing ```.
		blockEnd := strings.Index(contentStart, "```")
		if blockEnd == -1 {
			continue
		}

		content := strings.TrimSpace(contentStart[:blockEnd])
		if content != "" {
			evolved = append(evolved, PromptTemplate{
				Name:    tmpl.Name,
				Content: content,
			})
		}
	}

	return evolved
}

// extractSummary pulls out the summary text from Claude's response.
func extractSummary(response string) string {
	// Look for a "Summary" section.
	lower := strings.ToLower(response)
	idx := strings.Index(lower, "## summary")
	if idx == -1 {
		idx = strings.Index(lower, "# summary")
	}
	if idx == -1 {
		idx = strings.Index(lower, "**summary")
	}

	if idx != -1 {
		section := response[idx:]
		// Take until the next heading or end.
		nextHeading := strings.Index(section[1:], "\n#")
		if nextHeading != -1 {
			section = section[:nextHeading+1]
		}
		return strings.TrimSpace(section)
	}

	// Fallback: first 500 chars.
	if len(response) > 500 {
		return response[:500] + "..."
	}
	return response
}
