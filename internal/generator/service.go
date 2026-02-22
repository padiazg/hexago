package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
	"github.com/padiazg/hexago/pkg/utils"
)

// ServiceGenerator generates service/usecase files
type ServiceGenerator struct {
	config *ProjectConfig
}

// NewServiceGenerator creates a new service generator
func NewServiceGenerator(config *ProjectConfig) *ServiceGenerator {
	return &ServiceGenerator{
		config: config,
	}
}

// Generate creates a new service file
func (g *ServiceGenerator) Generate(serviceName, description string) error {
	// Determine service directory
	serviceDir := filepath.Join("internal", "core", g.config.CoreLogicDir())

	// Check if directory exists
	if !fileutil.FileExists(serviceDir) {
		return fmt.Errorf("directory %s does not exist. Are you in a hexagonal project?", serviceDir)
	}

	// Convert service name to file name (snake_case)
	fileName := utils.ToSnakeCase(serviceName) + ".go"
	testFileName := utils.ToSnakeCase(serviceName) + "_test.go"

	filePath := filepath.Join(serviceDir, fileName)
	testFilePath := filepath.Join(serviceDir, testFileName)

	// Check if file already exists
	if fileutil.FileExists(filePath) {
		return fmt.Errorf("service file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating service file: %s\n", filePath)

	// Generate service file
	if err := g.generateServiceFile(filePath, serviceName, description); err != nil {
		return err
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	// Generate test file
	if err := g.generateTestFile(testFilePath, serviceName); err != nil {
		return err
	}

	return nil
}

// generateServiceFile generates the service implementation file
func (g *ServiceGenerator) generateServiceFile(filePath, serviceName, description string) error {
	desc := description
	if desc == "" {
		desc = fmt.Sprintf("handles %s operations", serviceName)
	}

	data := map[string]interface{}{
		"CoreLogic":   g.config.CoreLogicDir(),
		"ModuleName":  g.config.ModuleName,
		"ServiceName": serviceName,
		"Description": desc,
	}

	content, err := g.config.templateLoader.Render("service/service.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render service template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateTestFile generates the test file
func (g *ServiceGenerator) generateTestFile(filePath, serviceName string) error {
	data := map[string]interface{}{
		"CoreLogic":   g.config.CoreLogicDir(),
		"ModuleName":  g.config.ModuleName,
		"ServiceName": serviceName,
	}

	content, err := g.config.templateLoader.Render("service/service_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render service test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// renderTemplateString is a helper to render templates
func renderTemplateString(tmpl string, data interface{}) ([]byte, error) {
	// Use the existing renderTemplate from project generator
	gen := &ProjectGenerator{}
	return gen.renderTemplate(tmpl, data)
}
