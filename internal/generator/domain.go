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

// constructorParams returns a comma-separated parameter list for a constructor.
// e.g. [{Name:"Id", Type:"string"}] → "id string"
func constructorParams(fields []Field) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = utils.SafeParamName(f.Name) + " " + f.Type
	}
	return strings.Join(parts, ", ")
}

// constructorInit returns the struct field initialization block body (indented with two tabs).
// e.g. [{Name:"Id", Type:"string"}] → "\t\tId: id,\n"
func constructorInit(fields []Field) string {
	var sb strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&sb, "\t\t%s: %s,\n", f.Name, utils.SafeParamName(f.Name))
	}
	return sb.String()
}

// constructorTestArgs returns constructor call arguments for test files using zero-value
// literals with inline type comments, ready to be embedded inside a function call.
// Returns "" when fields is empty (call becomes `New...()`).
// func constructorTestArgs(fields []Field) string {
// 	if len(fields) == 0 {
// 		return ""
// 	}
// 	var sb strings.Builder
// 	for _, f := range fields {
// 		fmt.Fprintf(&sb, "\n\t\t\t\t%s, // %s %s", utils.ZeroValueFor(f.Type), utils.LcFirst(f.Name), f.Type)
// 	}
// 	sb.WriteString("\n\t\t\t")
// 	return sb.String()
// }

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
	}

	data := map[string]any{
		"EntityName":        entityName,
		"PackageName":       pkgName,
		"FieldDefs":         fieldDefs,
		"HasTimeField":      hasTimeField,
		"ConstructorParams": constructorParams(fields),
		"ConstructorInit":   constructorInit(fields),
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
		"VOName":            voName,
		"PackageName":       pkgName,
		"FieldDefs":         fieldDefs,
		"Imports":           imports,
		"ConstructorParams": constructorParams(fields),
		"ConstructorInit":   constructorInit(fields),
	}

	content, err := g.config.templateLoader.Render("domain/value_object.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render value object template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
