package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// Package-level template loader (initialized once)
var globalTemplateLoader *TemplateLoader

func init() {
	globalTemplateLoader = NewTemplateLoader()
}

// generateMainFile generates the main.go file
func (g *ProjectGenerator) generateMainFile(projectPath string) error {
	content, err := globalTemplateLoader.Render("project/main.go.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render main.go template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "main.go"), content)
}

// generateRootCommand generates cmd/root.go
func (g *ProjectGenerator) generateRootCommand(projectPath string) error {
	content, err := globalTemplateLoader.Render("project/root_cmd.go.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render root_cmd.go template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "cmd", "root.go"), content)
}

// generateRunCommand generates cmd/run.go
func (g *ProjectGenerator) generateRunCommand(projectPath string) error {
	content, err := globalTemplateLoader.Render("project/run_cmd.go.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render run_cmd.go template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "cmd", "run.go"), content)
}

// generateConfig generates internal/config/config.go
func (g *ProjectGenerator) generateConfig(projectPath string) error {
	content, err := globalTemplateLoader.Render("project/config.go.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render config.go template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "internal", "config", "config.go"), content)
}

// generateLogger generates pkg/logger/logger.go
func (g *ProjectGenerator) generateLogger(projectPath string) error {
	content, err := globalTemplateLoader.Render("project/logger.go.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render logger.go template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "pkg", "logger", "logger.go"), content)
}

// renderTemplate renders a template with the given data
func (g *ProjectGenerator) renderTemplate(tmplStr string, data interface{}) ([]byte, error) {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"upper": func(s string) string {
			// Simple uppercase - can be enhanced
			return s
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]-32) + s[1:]
		},
	}).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
