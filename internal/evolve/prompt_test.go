package evolve

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWritePromptTemplates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-tmpl-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write some templates.
	templates := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "Build the project with tests."},
		{Name: "seo.md.tmpl", Content: "Add README, CLAUDE.md, llms.txt."},
		{Name: "review.md.tmpl", Content: "Review code quality and test coverage."},
	}

	for _, tmpl := range templates {
		if err := WritePromptTemplate(tmpDir, tmpl); err != nil {
			t.Fatalf("write %s: %v", tmpl.Name, err)
		}
	}

	// Verify files exist.
	for _, tmpl := range templates {
		path := filepath.Join(tmpDir, tmpl.Name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist", path)
		}
	}

	// Read them back.
	read, err := ReadPromptTemplates(tmpDir)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if len(read) != len(templates) {
		t.Fatalf("expected %d templates, got %d", len(templates), len(read))
	}

	// Check content matches.
	nameToContent := make(map[string]string)
	for _, tmpl := range read {
		nameToContent[tmpl.Name] = tmpl.Content
	}

	for _, tmpl := range templates {
		got, ok := nameToContent[tmpl.Name]
		if !ok {
			t.Errorf("missing template %s", tmpl.Name)
			continue
		}
		if got != tmpl.Content {
			t.Errorf("template %s content mismatch: got %q, want %q", tmpl.Name, got, tmpl.Content)
		}
	}
}

func TestReadPromptTemplatesEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	templates, err := ReadPromptTemplates(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}
}

func TestReadPromptTemplatesNonexistent(t *testing.T) {
	_, err := ReadPromptTemplates("/nonexistent/dir")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestWritePromptTemplateCreatesDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-mkdir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	deepDir := filepath.Join(tmpDir, "a", "b", "c")
	tmpl := PromptTemplate{Name: "test.md.tmpl", Content: "test content"}

	if err := WritePromptTemplate(deepDir, tmpl); err != nil {
		t.Fatalf("write to deep dir: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(deepDir, "test.md.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test content" {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestReadSkipsNonTmplFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prompt-evolver-skip-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a .tmpl and a non-.tmpl file.
	os.WriteFile(filepath.Join(tmpDir, "build.md.tmpl"), []byte("tmpl content"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("not a template"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("readme"), 0o644)

	templates, err := ReadPromptTemplates(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Errorf("expected 1 template (only .tmpl), got %d", len(templates))
	}
	if templates[0].Name != "build.md.tmpl" {
		t.Errorf("expected build.md.tmpl, got %s", templates[0].Name)
	}
}
