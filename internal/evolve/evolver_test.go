package evolve

import (
	"strings"
	"testing"
)

func TestParseEvolvedTemplates(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old build content"},
		{Name: "seo.md.tmpl", Content: "old seo content"},
	}

	response := `Here are the improved templates:

### build.md.tmpl
` + "```" + `
You are building a Go project. Always include:
1. Tests in every package
2. Correct go.mod module path
3. Error handling
` + "```" + `

### seo.md.tmpl
` + "```" + `markdown
Add these files to every project:
- README.md with install, usage, architecture
- CLAUDE.md with build commands and conventions
- llms.txt with project summary
` + "```" + `

## Summary
Improved build prompt to explicitly require tests and correct module paths.
Added specific file requirements to SEO prompt.
`

	evolved := parseEvolvedTemplates(response, current)
	if len(evolved) != 2 {
		t.Fatalf("expected 2 evolved templates, got %d", len(evolved))
	}

	for _, e := range evolved {
		if e.Content == "" {
			t.Errorf("template %s has empty content", e.Name)
		}
		if e.Name != "build.md.tmpl" && e.Name != "seo.md.tmpl" {
			t.Errorf("unexpected template name: %s", e.Name)
		}
	}
}

func TestParseEvolvedTemplatesNoMatch(t *testing.T) {
	current := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "old"},
	}

	response := "Here is some text with no code blocks for any template."
	evolved := parseEvolvedTemplates(response, current)
	if len(evolved) != 0 {
		t.Errorf("expected 0 evolved templates, got %d", len(evolved))
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "h2 summary",
			input:    "some text\n## Summary\nThis is the summary.\n## Next",
			contains: "This is the summary",
		},
		{
			name:     "h1 summary",
			input:    "text\n# Summary\nSummary here.\n# Other",
			contains: "Summary here",
		},
		{
			name:     "bold summary",
			input:    "intro\n**Summary**\nBold summary text.",
			contains: "Bold summary text",
		},
		{
			name:     "no summary fallback",
			input:    "Just some response without a summary heading.",
			contains: "Just some response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSummary(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected summary to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestResultString(t *testing.T) {
	r := &Result{
		Summary: "Improved test coverage instructions.",
		EvolvedPrompts: []PromptTemplate{
			{Name: "build.md.tmpl", Content: "new content here"},
		},
	}

	s := r.String()
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
