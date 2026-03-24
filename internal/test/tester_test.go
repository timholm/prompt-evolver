package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShipRate(t *testing.T) {
	tests := []struct {
		name    string
		results []BuildResult
		want    float64
	}{
		{"empty", nil, 0},
		{"all_shipped", []BuildResult{{Shipped: true}, {Shipped: true}}, 1.0},
		{"none_shipped", []BuildResult{{Shipped: false}, {Shipped: false}}, 0},
		{"half_shipped", []BuildResult{{Shipped: true}, {Shipped: false}}, 0.5},
		{"two_thirds", []BuildResult{{Shipped: true}, {Shipped: true}, {Shipped: false}}, 2.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shipRate(tt.results)
			if got != tt.want {
				t.Errorf("shipRate() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.txt")
	os.WriteFile(existing, []byte("hi"), 0o644)

	if !fileExists(existing) {
		t.Error("fileExists() = false for existing file")
	}
	if fileExists(filepath.Join(dir, "nope.txt")) {
		t.Error("fileExists() = true for nonexistent file")
	}
}

func TestHasTestFiles(t *testing.T) {
	t.Run("no_files", func(t *testing.T) {
		dir := t.TempDir()
		if hasTestFiles(dir) {
			t.Error("hasTestFiles() = true for empty dir")
		}
	})

	t.Run("no_test_files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
		if hasTestFiles(dir) {
			t.Error("hasTestFiles() = true with no test files")
		}
	})

	t.Run("top_level_test", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main"), 0o644)
		if !hasTestFiles(dir) {
			t.Error("hasTestFiles() = false with top-level test file")
		}
	})

	t.Run("subdirectory_test", func(t *testing.T) {
		dir := t.TempDir()
		subdir := filepath.Join(dir, "pkg")
		os.MkdirAll(subdir, 0o755)
		os.WriteFile(filepath.Join(subdir, "handler_test.go"), []byte("package pkg"), 0o644)
		if !hasTestFiles(dir) {
			t.Error("hasTestFiles() = false with subdirectory test file")
		}
	})

	t.Run("nonexistent_dir", func(t *testing.T) {
		if hasTestFiles("/nonexistent/dir/12345") {
			t.Error("hasTestFiles() = true for nonexistent dir")
		}
	})
}

func TestABResultString(t *testing.T) {
	result := &ABResult{
		OldResults: []BuildResult{
			{Project: "url-shortener", PromptSet: "old", Shipped: true, CompileOK: true, HasTests: true, HasReadme: true},
			{Project: "log-parser", PromptSet: "old", Shipped: false, CompileOK: true, HasTests: false, HasReadme: true},
		},
		NewResults: []BuildResult{
			{Project: "config-merger", PromptSet: "new", Shipped: true, CompileOK: true, HasTests: true, HasReadme: true},
			{Project: "port-scanner", PromptSet: "new", Shipped: true, CompileOK: true, HasTests: true, HasReadme: true},
		},
		OldShipRate: 0.5,
		NewShipRate: 1.0,
		Improved:    true,
	}

	s := result.String()
	if !strings.Contains(s, "A/B Test Results") {
		t.Error("result string missing header")
	}
	if !strings.Contains(s, "url-shortener") {
		t.Error("result string missing old project")
	}
	if !strings.Contains(s, "config-merger") {
		t.Error("result string missing new project")
	}
	if !strings.Contains(s, "BETTER") {
		t.Error("result string missing BETTER verdict")
	}
	if !strings.Contains(s, "50.0%") {
		t.Error("result string missing old ship rate")
	}
	if !strings.Contains(s, "100.0%") {
		t.Error("result string missing new ship rate")
	}
}

func TestABResultStringNotImproved(t *testing.T) {
	result := &ABResult{
		OldShipRate: 0.8,
		NewShipRate: 0.5,
		Improved:    false,
	}

	s := result.String()
	if !strings.Contains(s, "NOT better") {
		t.Error("result string missing NOT better verdict")
	}
}

func TestSampleProjectsNotEmpty(t *testing.T) {
	if len(SampleProjects) == 0 {
		t.Error("SampleProjects is empty")
	}
	if len(SampleProjects) < 6 {
		t.Errorf("SampleProjects has %d entries, need at least 6", len(SampleProjects))
	}
}

func TestBuildResultFields(t *testing.T) {
	br := BuildResult{
		Project:      "test-project",
		PromptSet:    "new",
		Shipped:      true,
		HasTests:     true,
		HasReadme:    true,
		CompileOK:    true,
		ErrorSnippet: "",
	}

	if br.Project != "test-project" {
		t.Errorf("Project = %q", br.Project)
	}
	if !br.Shipped {
		t.Error("expected Shipped = true")
	}
}

func TestTryCompileNoGoMod(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}"), 0o644)
	if tryCompile(dir) {
		t.Error("tryCompile() = true without go.mod")
	}
}
