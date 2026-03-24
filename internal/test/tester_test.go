package test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-fe-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	existing := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(existing, []byte("hi"), 0o644)

	if !fileExists(existing) {
		t.Error("expected file to exist")
	}
	if fileExists(filepath.Join(tmpDir, "nope.txt")) {
		t.Error("expected file to not exist")
	}
}

func TestHasTestFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-htf-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// No test files.
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0o644)
	if hasTestFiles(tmpDir) {
		t.Error("expected no test files")
	}

	// Add a test file.
	os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main"), 0o644)
	if !hasTestFiles(tmpDir) {
		t.Error("expected test file to be found")
	}
}

func TestHasTestFilesSubdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-htfs-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "internal")
	os.MkdirAll(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "foo_test.go"), []byte("package internal"), 0o644)

	if !hasTestFiles(tmpDir) {
		t.Error("expected test file in subdirectory to be found")
	}
}

func TestShipRate(t *testing.T) {
	results := []BuildResult{
		{Shipped: true},
		{Shipped: true},
		{Shipped: false},
	}

	rate := shipRate(results)
	if rate < 0.66 || rate > 0.67 {
		t.Errorf("expected ~0.667 ship rate, got %f", rate)
	}
}

func TestShipRateEmpty(t *testing.T) {
	rate := shipRate(nil)
	if rate != 0 {
		t.Errorf("expected 0 for empty results, got %f", rate)
	}
}

func TestShipRateAllShipped(t *testing.T) {
	results := []BuildResult{
		{Shipped: true},
		{Shipped: true},
	}
	rate := shipRate(results)
	if rate != 1.0 {
		t.Errorf("expected 1.0, got %f", rate)
	}
}

func TestShipRateNoneShipped(t *testing.T) {
	results := []BuildResult{
		{Shipped: false},
		{Shipped: false},
	}
	rate := shipRate(results)
	if rate != 0 {
		t.Errorf("expected 0, got %f", rate)
	}
}

func TestABResultString(t *testing.T) {
	r := &ABResult{
		OldResults: []BuildResult{
			{Project: "test-a", Shipped: true, CompileOK: true, HasTests: true, HasReadme: true},
		},
		NewResults: []BuildResult{
			{Project: "test-b", Shipped: true, CompileOK: true, HasTests: true, HasReadme: true},
		},
		OldShipRate: 1.0,
		NewShipRate: 1.0,
		Improved:    false,
	}

	s := r.String()
	if s == "" {
		t.Error("string should not be empty")
	}
}

func TestSampleProjectsNotEmpty(t *testing.T) {
	if len(SampleProjects) < 6 {
		t.Errorf("expected at least 6 sample projects, got %d", len(SampleProjects))
	}
}

func TestTryCompileNoGoMod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-tc-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// No go.mod = should fail.
	if tryCompile(tmpDir) {
		t.Error("expected compile to fail without go.mod")
	}
}
