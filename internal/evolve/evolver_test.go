package evolve

import (
	"strings"
	"testing"

	"github.com/timholm/prompt-evolver/internal/analyze"
)

func TestParseEvolvedTemplates(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old build content"},
		{Name: "review.md.tmpl", Content: "old review content"},
	}

	response := "Here are the improved templates:\n\n### build.md.tmpl\n\n```\nNew build instructions with guard rails.\nAlways include tests.\n```\n\n### review.md.tmpl\n\n```\nNew review checklist with verification steps.\nCheck compilation before shipping.\n```\n\n## Summary\nImproved both templates to address compilation errors and missing tests."

	evolved := parseEvolvedTemplates(response, current)

	if len(evolved) != 2 {
		t.Fatalf("expected 2 evolved templates, got %d", len(evolved))
	}

	buildFound := false
	reviewFound := false
	for _, tmpl := range evolved {
		switch tmpl.Name {
		case "build.md.tmpl":
			buildFound = true
			if !strings.Contains(tmpl.Content, "guard rails") {
				t.Error("build template missing expected content")
			}
		case "review.md.tmpl":
			reviewFound = true
			if !strings.Contains(tmpl.Content, "verification steps") {
				t.Error("review template missing expected content")
			}
		}
	}
	if !buildFound {
		t.Error("build.md.tmpl not found in evolved templates")
	}
	if !reviewFound {
		t.Error("review.md.tmpl not found in evolved templates")
	}
}

func TestParseEvolvedTemplatesNoMatch(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old"},
	}
	response := "Here is some text without any code blocks or matching template names."

	evolved := parseEvolvedTemplates(response, current)
	if len(evolved) != 0 {
		t.Errorf("expected 0 evolved templates, got %d", len(evolved))
	}
}

func TestParseEvolvedTemplatesEmptyCodeBlock(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old"},
	}
	response := "### build.md.tmpl\n```\n\n```\n"

	evolved := parseEvolvedTemplates(response, current)
	if len(evolved) != 0 {
		t.Errorf("expected 0 evolved templates for empty block, got %d", len(evolved))
	}
}

func TestParseEvolvedTemplatesWithLanguageTag(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old"},
	}
	response := "### build.md.tmpl\n```markdown\nNew content with language tag.\n```\n"

	evolved := parseEvolvedTemplates(response, current)
	if len(evolved) != 1 {
		t.Fatalf("expected 1 evolved template, got %d", len(evolved))
	}
	if !strings.Contains(evolved[0].Content, "language tag") {
		t.Error("content missing expected text")
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name:     "with_h2_summary",
			response: "Some text\n## Summary\nThis is the summary.\n## Next\nmore",
			want:     "## Summary\nThis is the summary.",
		},
		{
			name:     "with_h1_summary",
			response: "# Summary\nJust the summary text.",
			want:     "# Summary\nJust the summary text.",
		},
		{
			name:     "with_bold_summary",
			response: "**Summary**: Changed everything.\nDone.",
			want:     "**Summary**: Changed everything.\nDone.",
		},
		{
			name:     "no_summary_short",
			response: "Here is the response without a summary section.",
			want:     "Here is the response without a summary section.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSummary(tt.response)
			if got != tt.want {
				t.Errorf("extractSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractSummaryLongFallback(t *testing.T) {
	long := strings.Repeat("x", 600)
	got := extractSummary(long)
	if len(got) > 504 {
		t.Errorf("expected truncated summary, got %d chars", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("expected truncated summary to end with ...")
	}
}

func TestBuildEvolutionPrompt(t *testing.T) {
	report := &analyze.Report{
		TotalBuilds:  100,
		ShippedCount: 70,
		FailedCount:  30,
		ShipRate:     0.7,
		FailureGroups: []analyze.FailureGroup{
			{Pattern: "no_test_files", Desc: "No tests", Count: 15, Percentage: 0.5},
		},
		ShippedTraits: analyze.ShippedTraits{
			TestRate:   0.9,
			ReadmeRate: 0.8,
		},
	}

	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "Build a Go project."},
	}

	prompt := buildEvolutionPrompt(report, current)

	if !strings.Contains(prompt, "100") {
		t.Error("prompt missing total builds")
	}
	if !strings.Contains(prompt, "70.0%") {
		t.Error("prompt missing ship rate")
	}
	if !strings.Contains(prompt, "no_test_files") {
		t.Error("prompt missing failure pattern")
	}
	if !strings.Contains(prompt, "build.md.tmpl") {
		t.Error("prompt missing template name")
	}
	if !strings.Contains(prompt, "Build a Go project.") {
		t.Error("prompt missing template content")
	}
	if !strings.Contains(prompt, "expert prompt engineer") {
		t.Error("prompt missing system instruction")
	}
}

func TestResultString(t *testing.T) {
	result := &Result{
		Summary: "Improved test coverage instructions.",
		EvolvedPrompts: []PromptTemplate{
			{Name: "build.md.tmpl", Content: "new content here"},
		},
	}

	s := result.String()
	if !strings.Contains(s, "Prompt Evolution Result") {
		t.Error("result string missing header")
	}
	if !strings.Contains(s, "build.md.tmpl") {
		t.Error("result string missing template name")
	}
	if !strings.Contains(s, "Improved test coverage") {
		t.Error("result string missing summary")
	}
}
