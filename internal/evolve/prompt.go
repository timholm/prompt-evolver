package evolve

import (
	"fmt"
	"os"
	"path/filepath"
)

// PromptTemplate represents a prompt template file on disk.
type PromptTemplate struct {
	Name    string
	Content string
}

// ReadPromptTemplates reads all .md.tmpl files from the given directory.
func ReadPromptTemplates(dir string) ([]PromptTemplate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read prompt dir %s: %w", dir, err)
	}

	var templates []PromptTemplate
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".tmpl" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", name, err)
		}
		templates = append(templates, PromptTemplate{
			Name:    name,
			Content: string(content),
		})
	}

	return templates, nil
}

// WritePromptTemplate writes a prompt template to the given directory.
func WritePromptTemplate(dir string, tmpl PromptTemplate) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}
	path := filepath.Join(dir, tmpl.Name)
	if err := os.WriteFile(path, []byte(tmpl.Content), 0o644); err != nil {
		return fmt.Errorf("write template %s: %w", path, err)
	}
	return nil
}
