package analyze

import (
	"testing"
)

func TestDefaultPatternsNotEmpty(t *testing.T) {
	patterns := DefaultPatterns()
	if len(patterns) == 0 {
		t.Fatal("DefaultPatterns() returned empty slice")
	}
	for _, p := range patterns {
		if p.Name == "" {
			t.Error("pattern has empty name")
		}
		if p.Regex == nil {
			t.Errorf("pattern %q has nil regex", p.Name)
		}
		if p.Desc == "" {
			t.Errorf("pattern %q has empty description", p.Name)
		}
		if p.Matches != 0 {
			t.Errorf("pattern %q has non-zero initial matches: %d", p.Name, p.Matches)
		}
	}
}

func TestMatchPatternsNoTestFiles(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "no test files found in project directory")
	assertPatternMatched(t, matched, "no_test_files")
}

func TestMatchPatternsCompilationError(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "build failed: cannot compile main.go: syntax error")
	assertPatternMatched(t, matched, "compilation_error")
}

func TestMatchPatternsTestFailure(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "--- FAIL: TestSomething (0.01s)\n    expected foo got bar")
	assertPatternMatched(t, matched, "test_failure")
}

func TestMatchPatternsWrongModule(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "go.mod module path does not match expected")
	assertPatternMatched(t, matched, "wrong_module_path")
}

func TestMatchPatternsMissingDependency(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "cannot find package \"github.com/foo/bar\" in any of")
	assertPatternMatched(t, matched, "missing_dependency")
}

func TestMatchPatternsTimeout(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "context deadline exceeded after 300s")
	assertPatternMatched(t, matched, "timeout")
}

func TestMatchPatternsPermissionDenied(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "permission denied: /var/run/factory.sock")
	assertPatternMatched(t, matched, "permission_denied")
}

func TestMatchPatternsEmptyOutput(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "build produced no files in output directory")
	assertPatternMatched(t, matched, "empty_output")
}

func TestMatchPatternsLintFailure(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "golangci-lint found 3 issues")
	assertPatternMatched(t, matched, "lint_failure")
}

func TestMatchPatternsMissingReadme(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "README not found in build output")
	assertPatternMatched(t, matched, "missing_readme")
}

func TestMatchPatternsNoMatch(t *testing.T) {
	patterns := DefaultPatterns()
	matched := MatchPatterns(patterns, "everything is fine, nothing to see here")
	if len(matched) != 0 {
		t.Errorf("expected no matches, got %d", len(matched))
	}
}

func TestMatchPatternsMultipleMatches(t *testing.T) {
	patterns := DefaultPatterns()
	errorLog := "build failed: syntax error at line 5; also no test files found"
	matched := MatchPatterns(patterns, errorLog)
	if len(matched) < 2 {
		t.Errorf("expected >= 2 matches, got %d", len(matched))
	}
	assertPatternMatched(t, matched, "compilation_error")
	assertPatternMatched(t, matched, "no_test_files")
}

func assertPatternMatched(t *testing.T, matched []*Pattern, name string) {
	t.Helper()
	for _, p := range matched {
		if p.Name == name {
			return
		}
	}
	names := make([]string, len(matched))
	for i, p := range matched {
		names[i] = p.Name
	}
	t.Errorf("expected pattern %q to match, got: %v", name, names)
}
