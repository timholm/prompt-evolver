package analyze

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/timholm/prompt-evolver/internal/config"
	"github.com/timholm/prompt-evolver/internal/db"
)

// Report is the output of the analysis phase.
type Report struct {
	TotalBuilds   int              `json:"total_builds"`
	ShippedCount  int              `json:"shipped_count"`
	FailedCount   int              `json:"failed_count"`
	ShipRate      float64          `json:"ship_rate"`
	FailureGroups []FailureGroup   `json:"failure_groups"`
	ShippedTraits ShippedTraits    `json:"shipped_traits"`
	RawErrors     []string         `json:"-"`
}

// FailureGroup aggregates a single failure pattern across all failed builds.
type FailureGroup struct {
	Pattern    string  `json:"pattern"`
	Desc       string  `json:"description"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Examples   []string `json:"examples,omitempty"`
}

// ShippedTraits summarizes what went right in shipped builds.
type ShippedTraits struct {
	WithTests      int     `json:"with_tests"`
	WithReadme     int     `json:"with_readme"`
	CorrectModPath int     `json:"correct_mod_path"`
	TestRate       float64 `json:"test_rate"`
	ReadmeRate     float64 `json:"readme_rate"`
}

// String returns a human-readable report.
func (r *Report) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Build Analysis Report ===\n"))
	sb.WriteString(fmt.Sprintf("Total builds:  %d\n", r.TotalBuilds))
	sb.WriteString(fmt.Sprintf("Shipped:       %d (%.1f%%)\n", r.ShippedCount, r.ShipRate*100))
	sb.WriteString(fmt.Sprintf("Failed:        %d (%.1f%%)\n", r.FailedCount, (1-r.ShipRate)*100))
	sb.WriteString(fmt.Sprintf("\n--- Shipped Build Traits ---\n"))
	sb.WriteString(fmt.Sprintf("Has tests:     %d/%d (%.1f%%)\n", r.ShippedTraits.WithTests, r.ShippedCount, r.ShippedTraits.TestRate*100))
	sb.WriteString(fmt.Sprintf("Has README:    %d/%d (%.1f%%)\n", r.ShippedTraits.WithReadme, r.ShippedCount, r.ShippedTraits.ReadmeRate*100))
	sb.WriteString(fmt.Sprintf("Correct mod:   %d/%d\n", r.ShippedTraits.CorrectModPath, r.ShippedCount))
	sb.WriteString(fmt.Sprintf("\n--- Failure Patterns ---\n"))

	for _, g := range r.FailureGroups {
		sb.WriteString(fmt.Sprintf("  %-25s %3d  (%.1f%%)  %s\n", g.Pattern, g.Count, g.Percentage*100, g.Desc))
	}

	return sb.String()
}

// JSON returns the report as JSON bytes.
func (r *Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// Analyzer runs the analysis phase.
type Analyzer struct {
	cfg *config.Config
}

// New creates a new Analyzer.
func New(cfg *config.Config) *Analyzer {
	return &Analyzer{cfg: cfg}
}

// Run executes the full analysis against the factory registry.
func (a *Analyzer) Run() (*Report, error) {
	reg, err := db.Open(a.cfg.DBPath())
	if err != nil {
		return nil, fmt.Errorf("open registry: %w", err)
	}
	defer reg.Close()

	return a.AnalyzeBuilds(reg)
}

// AnalyzeBuilds processes builds from the given registry.
func (a *Analyzer) AnalyzeBuilds(reg *db.Registry) (*Report, error) {
	all, err := reg.AllBuilds()
	if err != nil {
		return nil, fmt.Errorf("fetch builds: %w", err)
	}

	return AnalyzeBuildList(all), nil
}

// AnalyzeBuildList processes a pre-fetched list of builds into a Report.
func AnalyzeBuildList(all []db.Build) *Report {
	report := &Report{
		TotalBuilds: len(all),
	}

	var shipped, failed []db.Build
	for _, b := range all {
		switch b.Status {
		case "shipped":
			shipped = append(shipped, b)
		case "failed":
			failed = append(failed, b)
		}
	}

	report.ShippedCount = len(shipped)
	report.FailedCount = len(failed)
	if report.TotalBuilds > 0 {
		report.ShipRate = float64(report.ShippedCount) / float64(report.TotalBuilds)
	}

	// Analyze shipped traits.
	for _, b := range shipped {
		if b.HasTests {
			report.ShippedTraits.WithTests++
		}
		if b.HasReadme {
			report.ShippedTraits.WithReadme++
		}
		if b.ModPath != "" {
			report.ShippedTraits.CorrectModPath++
		}
	}
	if report.ShippedCount > 0 {
		report.ShippedTraits.TestRate = float64(report.ShippedTraits.WithTests) / float64(report.ShippedCount)
		report.ShippedTraits.ReadmeRate = float64(report.ShippedTraits.WithReadme) / float64(report.ShippedCount)
	}

	// Match failure patterns.
	patterns := DefaultPatterns()
	for _, b := range failed {
		report.RawErrors = append(report.RawErrors, b.ErrorLog)
		matched := MatchPatterns(patterns, b.ErrorLog)
		for _, p := range matched {
			p.Matches++
		}
	}

	// Build failure groups sorted by count descending.
	for _, p := range patterns {
		if p.Matches > 0 {
			pct := 0.0
			if report.FailedCount > 0 {
				pct = float64(p.Matches) / float64(report.FailedCount)
			}
			fg := FailureGroup{
				Pattern:    p.Name,
				Desc:       p.Desc,
				Count:      p.Matches,
				Percentage: pct,
			}
			// Collect up to 3 example error snippets for this pattern.
			for _, errLog := range report.RawErrors {
				if p.Regex.MatchString(errLog) && len(fg.Examples) < 3 {
					snippet := errLog
					if len(snippet) > 200 {
						snippet = snippet[:200] + "..."
					}
					fg.Examples = append(fg.Examples, snippet)
				}
			}
			report.FailureGroups = append(report.FailureGroups, fg)
		}
	}

	sort.Slice(report.FailureGroups, func(i, j int) bool {
		return report.FailureGroups[i].Count > report.FailureGroups[j].Count
	})

	return report
}
