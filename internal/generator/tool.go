package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
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
	if err := fileutil.CreateDir(toolDir); err != nil {
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
	fileName := toSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating logger: %s\n", filePath)

	data := map[string]interface{}{
		"Name":        name,
		"Description": getDescription(description, "is a custom logger implementation"),
	}

	content, err := globalTemplateLoader.Render("tool/logger.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render logger template: %w", err)
	}

	if err := fileutil.WriteFile(filePath, content); err != nil {
		return err
	}

	return g.generateTestFile(dir, name, "logger")
}

// generateValidator generates an input validation utility
func (g *ToolGenerator) generateValidator(dir, name, description string) error {
	fileName := toSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating validator: %s\n", filePath)

	data := map[string]interface{}{
		"Name":        name,
		"Description": getDescription(description, "validates input data"),
	}

	content, err := globalTemplateLoader.Render("tool/validator.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render validator template: %w", err)
	}

	if err := fileutil.WriteFile(filePath, content); err != nil {
		return err
	}

	return g.generateTestFile(dir, name, "validator")
}

// generateMapper generates a DTO mapping utility
func (g *ToolGenerator) generateMapper(dir, name, description string) error {
	fileName := toSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating mapper: %s\n", filePath)

	data := map[string]interface{}{
		"Name":        name,
		"Description": getDescription(description, "maps between domain entities and DTOs"),
		"ModuleName":  g.config.ModuleName,
	}

	content, err := globalTemplateLoader.Render("tool/mapper.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render mapper template: %w", err)
	}

	if err := fileutil.WriteFile(filePath, content); err != nil {
		return err
	}

	return g.generateTestFile(dir, name, "mapper")
}

// generateMiddleware generates HTTP middleware
func (g *ToolGenerator) generateMiddleware(dir, name, description string) error {
	fileName := toSnakeCase(name) + ".go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating middleware: %s\n", filePath)

	data := map[string]interface{}{
		"Name":        name,
		"Description": getDescription(description, "is HTTP middleware"),
		"ModuleName":  g.config.ModuleName,
	}

	content, err := globalTemplateLoader.Render("tool/middleware.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render middleware template: %w", err)
	}

	if err := fileutil.WriteFile(filePath, content); err != nil {
		return err
	}

	return g.generateTestFile(dir, name, "middleware")
}

// generateTestFile generates a test file for the tool
func (g *ToolGenerator) generateTestFile(dir, name, toolType string) error {
	fileName := toSnakeCase(name) + "_test.go"
	filePath := filepath.Join(dir, fileName)

	fmt.Printf("📝 Creating test file: %s\n", filePath)

	data := map[string]interface{}{
		"Name":       name,
		"ToolType":   toolType,
		"ModuleName": g.config.ModuleName,
	}

	var templateName string
	switch toolType {
	case "logger":
		templateName = "tool/logger_test.go.tmpl"
	case "validator":
		templateName = "tool/validator_test.go.tmpl"
	case "mapper":
		templateName = "tool/mapper_test.go.tmpl"
	case "middleware":
		templateName = "tool/middleware_test.go.tmpl"
	default:
		templateName = "tool/generic_test.go.tmpl"
	}

	content, err := globalTemplateLoader.Render(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render tool test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

func getDescription(desc, defaultDesc string) string {
	if desc != "" {
		return desc
	}
	return defaultDesc
}
