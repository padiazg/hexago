package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// Field represents a struct field
type Field struct {
	Name string
	Type string
}

// DomainGenerator generates domain entities and value objects
type DomainGenerator struct {
	config *ProjectConfig
}

// NewDomainGenerator creates a new domain generator
func NewDomainGenerator(config *ProjectConfig) *DomainGenerator {
	return &DomainGenerator{
		config: config,
	}
}

// GenerateEntity creates a new domain entity
func (g *DomainGenerator) GenerateEntity(entityName string, fields []Field) error {
	domainDir := filepath.Join("internal", "core", "domain")

	if !fileutil.FileExists(domainDir) {
		return fmt.Errorf("directory %s does not exist", domainDir)
	}

	fileName := toSnakeCase(entityName) + ".go"
	testFileName := toSnakeCase(entityName) + "_test.go"

	filePath := filepath.Join(domainDir, fileName)
	testFilePath := filepath.Join(domainDir, testFileName)

	if fileutil.FileExists(filePath) {
		return fmt.Errorf("entity file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating entity file: %s\n", filePath)

	if err := g.generateEntityFile(filePath, entityName, fields); err != nil {
		return err
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	if err := g.generateEntityTestFile(testFilePath, entityName); err != nil {
		return err
	}

	return nil
}

// GenerateValueObject creates a new value object
func (g *DomainGenerator) GenerateValueObject(voName string, fields []Field) error {
	domainDir := filepath.Join("internal", "core", "domain")

	if !fileutil.FileExists(domainDir) {
		return fmt.Errorf("directory %s does not exist", domainDir)
	}

	fileName := toSnakeCase(voName) + ".go"
	testFileName := toSnakeCase(voName) + "_test.go"

	filePath := filepath.Join(domainDir, fileName)
	testFilePath := filepath.Join(domainDir, testFileName)

	if fileutil.FileExists(filePath) {
		return fmt.Errorf("value object file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating value object file: %s\n", filePath)

	if err := g.generateValueObjectFile(filePath, voName, fields); err != nil {
		return err
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	if err := g.generateValueObjectTestFile(testFilePath, voName); err != nil {
		return err
	}

	return nil
}

// generateEntityFile generates the entity implementation
func (g *DomainGenerator) generateEntityFile(filePath, entityName string, fields []Field) error {
	hasTimeField := false
	for _, f := range fields {
		if strings.Contains(f.Type, "time.Time") {
			hasTimeField = true
			break
		}
	}

	imports := `import (
	"errors"
`
	if hasTimeField {
		imports += `	"time"
`
	}
	imports += ")"

	// Generate field definitions
	fieldDefs := ""
	if len(fields) > 0 {
		for _, field := range fields {
			fieldDefs += fmt.Sprintf("\t%s %s\n", field.Name, field.Type)
		}
	} else {
		// Default fields if none provided
		fieldDefs = `	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
`
		hasTimeField = true
		imports = `import (
	"errors"
	"time"
)`
	}

	data := map[string]interface{}{
		"EntityName": entityName,
		"FieldDefs":  fieldDefs,
		"Imports":    imports,
	}

	content, err := globalTemplateLoader.Render("domain/entity.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render entity template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateEntityTestFile generates entity test file
func (g *DomainGenerator) generateEntityTestFile(filePath, entityName string) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"EntityName": entityName,
	}

	content, err := globalTemplateLoader.Render("domain/entity_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render entity test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateValueObjectFile generates the value object implementation
func (g *DomainGenerator) generateValueObjectFile(filePath, voName string, fields []Field) error {
	hasTimeField := false
	for _, f := range fields {
		if strings.Contains(f.Type, "time.Time") {
			hasTimeField = true
			break
		}
	}

	imports := `import (
	"errors"
	"fmt"
`
	if hasTimeField {
		imports += `	"time"
`
	}
	imports += ")"

	// Generate field definitions
	fieldDefs := ""
	if len(fields) > 0 {
		for _, field := range fields {
			fieldDefs += fmt.Sprintf("\t%s %s\n", field.Name, field.Type)
		}
	} else {
		// Default field if none provided
		fieldDefs = "\tvalue string\n"
	}

	data := map[string]interface{}{
		"VOName":    voName,
		"FieldDefs": fieldDefs,
		"Imports":   imports,
	}

	content, err := globalTemplateLoader.Render("domain/value_object.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render value object template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateValueObjectTestFile generates value object test file
func (g *DomainGenerator) generateValueObjectTestFile(filePath, voName string) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"VOName":     voName,
	}

	content, err := globalTemplateLoader.Render("domain/value_object_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render value object test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}
