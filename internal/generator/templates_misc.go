package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// generateMakefile generates the Makefile
func (g *ProjectGenerator) generateMakefile(projectPath string) error {
	content, err := g.config.templateLoader.Render("misc/makefile.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render makefile template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "Makefile"), content)
}

// generateGitignore generates the .gitignore file
func (g *ProjectGenerator) generateGitignore(projectPath string) error {
	content, err := g.config.templateLoader.Render("misc/gitignore.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render gitignore template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, ".gitignore"), content)
}

// generateReadme generates the README.md file
func (g *ProjectGenerator) generateReadme(projectPath string) error {
	content, err := g.config.templateLoader.Render("misc/readme.md.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render readme template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "README.md"), content)
}

// generateDockerFiles generates Dockerfile and compose.yaml
func (g *ProjectGenerator) generateDockerFiles(projectPath string) error {
	// Generate Dockerfile
	dockerContent, err := g.config.templateLoader.Render("misc/dockerfile.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render dockerfile template: %w", err)
	}

	if err := fileutil.WriteFile(filepath.Join(projectPath, "Dockerfile"), dockerContent); err != nil {
		return err
	}

	// Generate compose.yaml
	composeContent, err := g.config.templateLoader.Render("misc/compose.yaml.tmpl", g.config)
	if err != nil {
		return fmt.Errorf("failed to render compose template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "compose.yaml"), composeContent)
}

// generateObservability generates observability files
func (g *ProjectGenerator) generateObservability(projectPath string) error {
	// Generate health.go
	healthContent, err := g.config.templateLoader.Render("observability/health.go.tmpl", nil)
	if err != nil {
		return fmt.Errorf("failed to render health template: %w", err)
	}

	if err := fileutil.WriteFile(filepath.Join(projectPath, "internal", "observability", "health.go"), healthContent); err != nil {
		return err
	}

	// Generate metrics.go
	metricsContent, err := g.config.templateLoader.Render("observability/metrics.go.tmpl", nil)
	if err != nil {
		return fmt.Errorf("failed to render metrics template: %w", err)
	}

	if err := fileutil.WriteFile(filepath.Join(projectPath, "internal", "observability", "metrics.go"), metricsContent); err != nil {
		return err
	}

	// Generate server.go
	serverContent, err := g.config.templateLoader.Render("observability/server.go.tmpl", nil)
	if err != nil {
		return fmt.Errorf("failed to render server template: %w", err)
	}

	return fileutil.WriteFile(filepath.Join(projectPath, "internal", "observability", "server.go"), serverContent)
}
