package evolve

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPromptTemplatesEmpty(t *testing.T) {
	dir := t.TempDir()
	templates, err := ReadPromptTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPromptTemplates: %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}
}

func TestReadPromptTemplatesNonexistentDir(t *testing.T) {
	_, err := ReadPromptTemplates("/nonexistent/dir/12345")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestReadPromptTemplatesFiltersNonTmpl(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "build.md.tmpl"), []byte("build instructions"), 0o644)
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("some notes"), 0o644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("readme"), 0o644)
	os.WriteFile(filepath.Join(dir, "review.md.tmpl"), []byte("review instructions"), 0o644)

	templates, err := ReadPromptTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPromptTemplates: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	names := map[string]bool{}
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	if !names["build.md.tmpl"] {
		t.Error("missing build.md.tmpl")
	}
	if !names["review.md.tmpl"] {
		t.Error("missing review.md.tmpl")
	}
}

func TestReadPromptTemplatesContent(t *testing.T) {
	dir := t.TempDir()
	content := "You are a code generation agent.\nBuild the project as specified."
	os.WriteFile(filepath.Join(dir, "system.md.tmpl"), []byte(content), 0o644)

	templates, err := ReadPromptTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPromptTemplates: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].Content != content {
		t.Errorf("content = %q, want %q", templates[0].Content, content)
	}
}

func TestReadPromptTemplatesSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir.tmpl"), 0o755)
	os.WriteFile(filepath.Join(dir, "real.tmpl"), []byte("real template"), 0o644)

	templates, err := ReadPromptTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPromptTemplates: %v", err)
	}
	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}

func TestWritePromptTemplate(t *testing.T) {
	dir := t.TempDir()
	tmpl := PromptTemplate{
		Name:    "build.md.tmpl",
		Content: "Build the project with tests and README.",
	}

	if err := WritePromptTemplate(dir, tmpl); err != nil {
		t.Fatalf("WritePromptTemplate: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "build.md.tmpl"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != tmpl.Content {
		t.Errorf("written content = %q, want %q", string(data), tmpl.Content)
	}
}

func TestWritePromptTemplateCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	tmpl := PromptTemplate{Name: "test.tmpl", Content: "content"}

	if err := WritePromptTemplate(dir, tmpl); err != nil {
		t.Fatalf("WritePromptTemplate: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "test.tmpl")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	templates := []PromptTemplate{
		{Name: "build.md.tmpl", Content: "Build instructions here."},
		{Name: "review.md.tmpl", Content: "Review checklist here."},
	}

	for _, tmpl := range templates {
		if err := WritePromptTemplate(dir, tmpl); err != nil {
			t.Fatalf("WritePromptTemplate(%s): %v", tmpl.Name, err)
		}
	}

	read, err := ReadPromptTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPromptTemplates: %v", err)
	}
	if len(read) != len(templates) {
		t.Fatalf("expected %d templates, got %d", len(templates), len(read))
	}

	readMap := map[string]string{}
	for _, tmpl := range read {
		readMap[tmpl.Name] = tmpl.Content
	}

	for _, tmpl := range templates {
		if readMap[tmpl.Name] != tmpl.Content {
			t.Errorf("round-trip mismatch for %s: got %q, want %q", tmpl.Name, readMap[tmpl.Name], tmpl.Content)
		}
	}
}
