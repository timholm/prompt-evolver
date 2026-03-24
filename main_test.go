package main

import (
	"testing"
)

func TestAnalyzeCmd(t *testing.T) {
	cmd := analyzeCmd()
	if cmd.Use != "analyze" {
		t.Errorf("Use = %q, want %q", cmd.Use, "analyze")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	// Verify the --db flag exists.
	f := cmd.Flags().Lookup("db")
	if f == nil {
		t.Error("missing --db flag")
	}
}

func TestEvolveCmd(t *testing.T) {
	cmd := evolveCmd()
	if cmd.Use != "evolve" {
		t.Errorf("Use = %q, want %q", cmd.Use, "evolve")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	f := cmd.Flags().Lookup("analysis")
	if f == nil {
		t.Error("missing --analysis flag")
	}
}

func TestTestCmd(t *testing.T) {
	cmd := testCmd()
	if cmd.Use != "test" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	f := cmd.Flags().Lookup("count")
	if f == nil {
		t.Error("missing --count flag")
	}
	if f.DefValue != "3" {
		t.Errorf("--count default = %q, want %q", f.DefValue, "3")
	}
}

func TestDeployCmd(t *testing.T) {
	cmd := deployCmd()
	if cmd.Use != "deploy" {
		t.Errorf("Use = %q, want %q", cmd.Use, "deploy")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	f := cmd.Flags().Lookup("force")
	if f == nil {
		t.Error("missing --force flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", f.DefValue, "false")
	}
}
