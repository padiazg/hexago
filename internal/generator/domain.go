package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/padiazg/hexago/pkg/utils"
)

// Field represents a struct field
type Field struct {
	Name string
	Type string
}

// builtinTypeNames are Go types that need no import.
var builtinTypeNames = map[string]bool{
	"string": true, "bool": true, "byte": true, "rune": true,
	"error": true, "any": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"uintptr": true, "float32": true, "float64": true,
	"complex64": true, "complex128": true,
}

// fieldBaseType strips slice/pointer/map wrappers from a field type, returning
// the underlying named type (e.g. "[]*uuid.UUID" → "uuid.UUID").
func fieldBaseType(t string) string {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "[]")
	for strings.HasPrefix(t, "*") {
		t = strings.TrimPrefix(t, "*")
	}
	if strings.HasPrefix(t, "map[") {
		if i := strings.Index(t, "]"); i >= 0 {
			t = strings.TrimPrefix(t[i+1:], "*")
		}
	}
	return t
}

// resolveDomainImports scans internal/core/domain (including sub-packages) and
// returns a map of exported type name → module import path.
func resolveDomainImports(module string) map[string]string {
	domainDir := filepath.Join("internal", "core", "domain")
	index := map[string]string{}
	_ = filepath.WalkDir(domainDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		parsePath(path, module, func(i, importPath string) { index[i] = importPath })
		return nil
	})

	return index
}

func parsePath(path string, module string, addFn func(string, string)) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return
	}

	importPath := builImportpath(path, module)
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if ok && ts.Name.IsExported() {
				addFn(ts.Name.Name, importPath)
				// index[ts.Name.Name] = importPath
			}
		}
	}
}

func builImportpath(path string, module string) string {
	pkgRel := filepath.ToSlash(filepath.Dir(path))
	pkgRel = strings.TrimPrefix(pkgRel, filepath.ToSlash("internal/core/domain"))
	pkgRel = strings.TrimPrefix(pkgRel, "/")
	importPath := module + "/internal/core/domain/" + pkgRel
	return importPath
}

// packageAlias returns the default package name for an import path
// (the last path segment).
func packageAlias(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

// collectImports returns the set of import paths needed by the given fields,
// resolving custom domain types against the type index. baseImports are always
// included. Imports living in selfPkg are skipped.
func collectImports(selfPkg string, baseImports []string, fields []Field, index map[string]string) map[string]bool {
	imports := map[string]bool{}
	for _, b := range baseImports {
		imports[b] = true
	}
	for _, f := range fields {
		base := fieldBaseType(f.Type)
		switch {
		case base == "" || builtinTypeNames[base]:
			// no import needed
		case strings.HasPrefix(base, "time."):
			imports["time"] = true
		case strings.HasPrefix(base, "uuid."):
			imports["github.com/google/uuid"] = true
		case base == "RawMessage" || base == "Number" || strings.HasPrefix(base, "json."):
			imports["encoding/json"] = true
		default:
			if p, ok := index[base]; ok && p != selfPkg {
				imports[p] = true
			}
		}
	}
	return imports
}

// qualifyField rewrites a field type so bare references to custom domain types
// living in other packages are prefixed with their package alias.
// e.g. "MoneyExample" → "moneyexample.MoneyExample", "[]MoneyExample" → "[]moneyexample.MoneyExample".
func qualifyField(f Field, index map[string]string, selfPkg string) Field {
	t := strings.TrimSpace(f.Type)
	qualifyNamed := func(named string) string {
		if base := fieldBaseType(named); base != "" {
			if p, ok := index[base]; ok && p != selfPkg {
				return strings.Replace(named, base, packageAlias(p)+"."+base, 1)
			}
		}
		return named
	}
	switch {
	case strings.HasPrefix(t, "[]"):
		if i := strings.Index(t, "]"); i >= 0 {
			f.Type = t[:i+1] + qualifyNamed(t[i+1:])
		}
	case strings.HasPrefix(t, "map["):
		if i := strings.Index(t, "]"); i >= 0 {
			f.Type = t[:i+1] + qualifyNamed(t[i+1:])
		}
	default:
		f.Type = qualifyNamed(t)
	}

	return f
}

// renderImportBlock renders a sorted import(...) block (or "") from a set.
func renderImportBlock(imports map[string]bool) string {
	if len(imports) == 0 {
		return ""
	}
	paths := make([]string, 0, len(imports))
	for p := range imports {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var sb strings.Builder
	sb.WriteString("import (\n")
	for _, p := range paths {
		sb.WriteString("\t")
		sb.WriteString(strconv.Quote(p))
		sb.WriteString("\n")
	}
	sb.WriteString(")")
	return sb.String()
}

// DomainGenerator generates domain entities and value objects
type DomainGenerator struct {
	config *HexagoConfig
}

// NewDomainGenerator creates a new domain generator
func NewDomainGenerator(config *HexagoConfig) *DomainGenerator {
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

// ensureExternalDeps runs `go get` for any third-party import the generated
// file references but the project does not yet depend on (e.g. google/uuid).
// Failures are non-fatal: the file is still generated.
func (g *DomainGenerator) ensureExternalDeps(imports map[string]bool) {
	for p := range imports {
		first := p
		if i := strings.Index(p, "/"); i >= 0 {
			first = p[:i]
		}
		if !strings.Contains(first, ".") {
			continue // stdlib
		}
		if strings.HasPrefix(p, g.config.Project.Module+"/") {
			continue // module-internal
		}
		cmd := exec.Command("go", "get", p)
		cmd.Dir = g.config.OutputDir
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Warning: could not add dependency %s: %s\n", p, strings.TrimSpace(string(out)))
		}
	}
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

	selfPkgImport := g.config.Project.Module + "/internal/core/domain/" + pkgName
	index := resolveDomainImports(g.config.Project.Module)
	qualified := make([]Field, len(fields))
	for i, f := range fields {
		qualified[i] = qualifyField(f, index, selfPkgImport)
	}
	imports := collectImports(selfPkgImport, nil, fields, index)

	// Generate field definitions
	fieldDefs := ""
	if len(qualified) > 0 {
		for _, field := range qualified {
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
		"Imports":           renderImportBlock(imports),
		"FieldDefs":         fieldDefs,
		"HasTimeField":      hasTimeField,
		"ConstructorParams": constructorParams(qualified),
		"ConstructorInit":   constructorInit(qualified),
	}

	content, err := g.config.TemplateLoader.Render("domain/entity.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render entity template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}
	g.ensureExternalDeps(imports)
	return nil
}

// generatePortFile generates the repository port interface for an entity
func (g *DomainGenerator) generatePortFile(filePath, entityName, pkgName string) error {
	data := map[string]any{
		"PackageName": pkgName,
		"EntityName":  entityName,
	}

	content, err := g.config.TemplateLoader.Render("domain/port.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render port template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateValueObjectFile generates the value object implementation
func (g *DomainGenerator) generateValueObjectFile(filePath, voName, pkgName string, fields []Field) error {
	selfPkgImport := g.config.Project.Module + "/internal/core/domain/" + pkgName
	index := resolveDomainImports(g.config.Project.Module)
	qualified := make([]Field, len(fields))
	for i, f := range fields {
		qualified[i] = qualifyField(f, index, selfPkgImport)
	}
	// Value objects always render a String() method using fmt.
	imports := collectImports(selfPkgImport, []string{"fmt"}, fields, index)
	importBlock := renderImportBlock(imports)

	// Generate field definitions
	fieldDefs := ""
	if len(qualified) > 0 {
		for _, field := range qualified {
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
		"Imports":           importBlock,
		"ConstructorParams": constructorParams(qualified),
		"ConstructorInit":   constructorInit(qualified),
	}

	content, err := g.config.TemplateLoader.Render("domain/value_object.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render value object template: %w", err)
	}

	if err := utils.WriteFile(filePath, content); err != nil {
		return err
	}
	g.ensureExternalDeps(imports)
	return nil
}
