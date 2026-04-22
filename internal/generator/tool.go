package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/utils"
)

// ToolGenerator generates infrastructure tools
type ToolGenerator struct {
	config *ProjectConfig
}

// NewToolGenerator creates a new tool generator
func NewToolGenerator(config *ProjectConfig) *ToolGenerator {
	return &ToolGenerator{
		config: config,
	}
}

// Generate creates a new infrastructure tool
func (g *ToolGenerator) Generate(toolType, toolName, description string) error {
	// Create directory
	toolDir := filepath.Join("internal", "infrastructure", toolType)
	if err := utils.CreateDir(toolDir); err != nil {
		return err
	}

	// Generate based on type
	switch toolType {
	case "logger":
		return g.generateLogger(toolDir, toolName, description)
	case "validator":
		return g.generateValidator(toolDir, toolName, description)
	case "mapper":
		return g.generateMapper(toolDir, toolName, description)
	case "middleware":
		return g.generateMiddleware(toolDir, toolName, description)
	default:
		return fmt.Errorf("unsupported tool type: %s", toolType)
	}
}

// generateLogger generates a custom logger implementation
func (g *ToolGenerator) generateLogger(dir, name, description string) error {
	fileName := utils.ToSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating logger: %s\n", filePath)

	data := map[string]any{
		"Name":        name,
		"Description": getDescription(description, "is a custom logger implementation"),
	}

	content, err := g.config.templateLoader.Render("tool/logger.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render logger template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}

	return nil
}

// generateValidator generates an input validation utility
func (g *ToolGenerator) generateValidator(dir, name, description string) error {
	fileName := utils.ToSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating validator: %s\n", filePath)

	data := map[string]any{
		"Name":        name,
		"Description": getDescription(description, "validates input data"),
	}

	content, err := g.config.templateLoader.Render("tool/validator.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render validator template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}

	return nil
}

// generateMapper generates a DTO mapping utility
func (g *ToolGenerator) generateMapper(dir, name, description string) error {
	fileName := utils.ToSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating mapper: %s\n", filePath)

	data := map[string]any{
		"Name":        name,
		"Description": getDescription(description, "maps between domain entities and DTOs"),
		"ModuleName":  g.config.ModuleName,
	}

	content, err := g.config.templateLoader.Render("tool/mapper.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render mapper template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}

	return nil
}

// generateMiddleware generates HTTP middleware
func (g *ToolGenerator) generateMiddleware(dir, name, description string) error {
	fileName := utils.ToSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating middleware: %s\n", filePath)

	data := map[string]any{
		"Name":        name,
		"Description": getDescription(description, "is HTTP middleware"),
		"ModuleName":  g.config.ModuleName,
	}

	content, err := g.config.templateLoader.Render("tool/middleware.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render middleware template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}

	return nil
}

func getDescription(desc, defaultDesc string) string {
	if desc != "" {
		return desc
	}
	return defaultDesc
}
