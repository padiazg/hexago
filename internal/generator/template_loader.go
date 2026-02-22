package generator

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/padiazg/hexago/pkg/fileutil"
	"github.com/padiazg/hexago/pkg/utils"
)

//go:embed templates/**/*.tmpl
var embeddedTemplates embed.FS

// TemplateLoader handles loading and rendering templates
type TemplateLoader struct {
	funcMap template.FuncMap
	cache   map[string]*template.Template
	sources []TemplateSource
}

// TemplateSource represents a source of templates
type TemplateSource struct {
	Name     string
	Path     string
	Priority int
	exists   func(string) bool
	read     func(string) ([]byte, error)
}

// NewTemplateLoader creates a new template loader with multi-source support
func NewTemplateLoader() *TemplateLoader {
	loader := &TemplateLoader{
		funcMap: createTemplateFuncMap(),
		cache:   make(map[string]*template.Template),
	}

	// Setup template sources in priority order
	loader.sources = []TemplateSource{
		// 1. Binary-local templates (./templates/ relative to binary)
		{
			Name:     "binary-local",
			Path:     filepath.Join(fileutil.BinaryDir(), "templates"),
			Priority: 1,
			exists:   fileutil.FileExists,
			read:     os.ReadFile,
		},
		// 2. Project-local overrides (./.hexago/templates/)
		{
			Name:     "project-local",
			Path:     ".hexago/templates",
			Priority: 2,
			exists:   fileutil.FileExists,
			read:     os.ReadFile,
		},
		// 3. User-global overrides (~/.hexago/templates/)
		{
			Name:     "user-global",
			Path:     filepath.Join(fileutil.HomeDir(), ".hexago", "templates"),
			Priority: 3,
			exists:   fileutil.FileExists,
			read:     os.ReadFile,
		},
		// 4. Embedded templates (fallback)
		{
			Name:     "embedded",
			Priority: 4,
			exists: func(name string) bool {
				path := filepath.Join("templates", name)
				_, err := embeddedTemplates.ReadFile(path)
				return err == nil
			},
			read: func(name string) ([]byte, error) {
				path := filepath.Join("templates", name)
				return embeddedTemplates.ReadFile(path)
			},
		},
	}

	return loader
}

// Load loads and parses a template by name
func (l *TemplateLoader) Load(name string) (*template.Template, error) {
	// Check cache first
	if tmpl, ok := l.cache[name]; ok {
		return tmpl, nil
	}

	// Try each source in priority order
	for _, source := range l.sources {
		var content []byte
		var err error

		if source.Name == "embedded" {
			// Read from embedded FS
			if source.exists(name) {
				content, err = source.read(name)
				if err == nil {
					return l.parseTemplate(name, content, source.Name)
				}
			}
		} else {
			// Read from filesystem
			path := filepath.Join(source.Path, name)
			if source.exists(path) {
				content, err = source.read(path)
				if err == nil {
					return l.parseTemplate(name, content, source.Name)
				}
			}
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// parseTemplate parses template content with custom functions
func (l *TemplateLoader) parseTemplate(name string, content []byte, source string) (*template.Template, error) {
	tmpl, err := template.New(name).Funcs(l.funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s from %s: %w", name, source, err)
	}

	// Cache the parsed template
	l.cache[name] = tmpl
	return tmpl, nil
}

// Render renders a template with the given data
func (l *TemplateLoader) Render(name string, data interface{}) ([]byte, error) {
	tmpl, err := l.Load(name)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

// Exists checks if a template exists in any source
func (l *TemplateLoader) Exists(name string) bool {
	for _, source := range l.sources {
		if source.Name == "embedded" {
			if source.exists(name) {
				return true
			}
		} else {
			path := filepath.Join(source.Path, name)
			if source.exists(path) {
				return true
			}
		}
	}
	return false
}

// Which returns the source that will be used for a template
func (l *TemplateLoader) Which(name string) (string, error) {
	for _, source := range l.sources {
		if source.Name == "embedded" {
			if source.exists(name) {
				return fmt.Sprintf("%s (embedded)", source.Name), nil
			}
		} else {
			path := filepath.Join(source.Path, name)
			if source.exists(path) {
				return fmt.Sprintf("%s (%s)", source.Name, path), nil
			}
		}
	}
	return "", fmt.Errorf("template not found: %s", name)
}

// List returns all available templates
func (l *TemplateLoader) List() ([]string, error) {
	templates := make(map[string]bool)

	// Walk embedded templates
	err := fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			// Remove "templates/" prefix
			name := strings.TrimPrefix(path, "templates/")
			templates[name] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	var result []string
	for name := range templates {
		result = append(result, name)
	}

	return result, nil
}

// Export copies a template to a local or global override location
func (l *TemplateLoader) Export(name string, global bool) error {
	// Load the template to ensure it exists
	content, err := l.loadRawTemplate(name)
	if err != nil {
		return err
	}

	// Determine destination
	var destPath string
	if global {
		destPath = filepath.Join(fileutil.HomeDir(), ".hexago", "templates", name)
	} else {
		destPath = filepath.Join(".hexago", "templates", name)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write template
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// Validate parses the template at path to check for syntax errors
func (l *TemplateLoader) Validate(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	_, err = template.New("validate").Funcs(l.funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}
	return nil
}

// Reset removes a custom template override (project-local or user-global)
func (l *TemplateLoader) Reset(name string, global bool) error {
	var path string
	if global {
		path = filepath.Join(fileutil.HomeDir(), ".hexago", "templates", name)
	} else {
		path = filepath.Join(".hexago", "templates", name)
	}
	if !fileutil.FileExists(path) {
		return fmt.Errorf("no custom override found at %s", path)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove template override: %w", err)
	}
	return nil
}

// loadRawTemplate loads template content without parsing
func (l *TemplateLoader) loadRawTemplate(name string) ([]byte, error) {
	for _, source := range l.sources {
		if source.Name == "embedded" {
			if source.exists(name) {
				return source.read(name)
			}
		} else {
			path := filepath.Join(source.Path, name)
			if source.exists(path) {
				return source.read(path)
			}
		}
	}
	return nil, fmt.Errorf("template not found: %s", name)
}

// createTemplateFuncMap creates custom template functions
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": utils.ToTitleCase,
		"snake": utils.ToSnakeCase,
	}
}
