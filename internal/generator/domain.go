package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/pkg/utils"
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
	baseDomainDir := filepath.Join("internal", "core", "domain")
	if !utils.FileExists(baseDomainDir) {
		return fmt.Errorf("directory %s does not exist", baseDomainDir)
	}

	pkgName := utils.ToPlural(strings.ToLower(entityName))
	domainDir := filepath.Join(baseDomainDir, pkgName)

	if err := utils.CreateDir(domainDir); err != nil {
		return fmt.Errorf("creating directory %s: %w", domainDir, err)
	}

	fileName := pkgName + ".go"

	filePath := filepath.Join(domainDir, fileName)

	if utils.FileExists(filePath) {
		return fmt.Errorf("entity file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating entity file: %s\n", filePath)

	if err := g.generateEntityFile(filePath, entityName, pkgName, fields); err != nil {
		return err
	}

	fmt.Printf("📝 Creating port file: %s\n", filepath.Join(domainDir, "port.go"))

	if err := g.generatePortFile(filepath.Join(domainDir, "port.go"), entityName, pkgName); err != nil {
		return err
	}

	return nil
}

// GenerateValueObject creates a new value object.
// If entityName is non-empty, the VO is co-located inside that entity's sub-package.
// If entityName is empty, the VO gets its own standalone sub-package.
func (g *DomainGenerator) GenerateValueObject(voName, entityName string, fields []Field) error {
	baseDomainDir := filepath.Join("internal", "core", "domain")
	if !utils.FileExists(baseDomainDir) {
		return fmt.Errorf("directory %s does not exist", baseDomainDir)
	}

	var pkgName, voDir string
	if entityName != "" {
		// Entity-bound: co-locate inside the entity's sub-package (must already exist)
		pkgName = utils.ToPlural(strings.ToLower(entityName))
		voDir = filepath.Join(baseDomainDir, pkgName)
		if !utils.FileExists(voDir) {
			return fmt.Errorf("entity directory %s does not exist; create the entity first", voDir)
		}
	} else {
		// Standalone: own sub-package named after the VO
		pkgName = strings.ToLower(voName)
		voDir = filepath.Join(baseDomainDir, pkgName)
		if err := utils.CreateDir(voDir); err != nil {
			return fmt.Errorf("creating directory %s: %w", voDir, err)
		}
	}

	fileName := utils.ToSnakeCase(voName) + ".go"

	filePath := filepath.Join(voDir, fileName)

	if utils.FileExists(filePath) {
		return fmt.Errorf("value object file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating value object file: %s\n", filePath)

	if err := g.generateValueObjectFile(filePath, voName, pkgName, fields); err != nil {
		return err
	}

	return nil
}

// generateEntityFile generates the entity implementation
func (g *DomainGenerator) generateEntityFile(filePath, entityName, pkgName string, fields []Field) error {
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

	data := map[string]any{
		"EntityName":  entityName,
		"PackageName": pkgName,
		"FieldDefs":   fieldDefs,
		"Imports":     imports,
	}

	content, err := g.config.templateLoader.Render("domain/entity.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render entity template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generatePortFile generates the repository port interface for an entity
func (g *DomainGenerator) generatePortFile(filePath, entityName, pkgName string) error {
	data := map[string]any{
		"PackageName": pkgName,
		"EntityName":  entityName,
	}

	content, err := g.config.templateLoader.Render("domain/port.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render port template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateValueObjectFile generates the value object implementation
func (g *DomainGenerator) generateValueObjectFile(filePath, voName, pkgName string, fields []Field) error {
	hasTimeField := false
	for _, f := range fields {
		if strings.Contains(f.Type, "time.Time") {
			hasTimeField = true
			break
		}
	}

	imports := `import (
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

	data := map[string]any{
		"VOName":      voName,
		"PackageName": pkgName,
		"FieldDefs":   fieldDefs,
		"Imports":     imports,
	}

	content, err := g.config.templateLoader.Render("domain/value_object.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render value object template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
