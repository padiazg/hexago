package adapter

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/internal/analyzer"
	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/utils"
	"golang.org/x/tools/go/packages"
)

type Adapter interface {
	Generate() error
}

// AdapterGenerator generates adapter files
type AdapterGenerator struct {
	config *generator.HexagoConfig
}

// NewAdapterGenerator creates a new adapter generator
func NewAdapterGenerator(config *generator.HexagoConfig) *AdapterGenerator {
	return &AdapterGenerator{
		config: config,
	}
}

// GeneratePrimary generates a primary (inbound) adapter.
// For HTTP adapters, entityName (optional) triggers sub-package generation with
// two files: <snake_entity>.go (Config/DTOs) and handlers.go (HTTP methods).
func (g *AdapterGenerator) GeneratePrimary(adapterType, adapterName, entityName, portName string) error {
	// Validate adapter type
	validTypes := map[string]bool{
		"http":  true,
		"grpc":  true,
		"queue": true,
		"cli":   true,
	}

	if !validTypes[adapterType] {
		return fmt.Errorf("invalid primary adapter type '%s'. Valid types: http, grpc, queue, cli", adapterType)
	}

	// if adapterType == "http" && entityName != "" {
	// 	return newHttpHandlerAdapter(g.config, entityName).Generate()
	// }

	// Default: flat directory
	var adapter Adapter
	switch {
	// HTTP + entity → sub-package with two files
	case adapterType == "http":
		adapter = newHttpAdapter(g.config, adapterName, entityName)
	case adapterType == "grpc":
		adapter = newGRPCAdapter(g.config, adapterName)
	case adapterType == "queue":
		adapter = newQueueAdapter(g.config, adapterName)
	case adapterType == "cli":
		adapter = newCLIAdapter(g.config, adapterName)
	default:
		return fmt.Errorf("adapter type %s not yet implemented. Valid types: http, grpc, queue, cli", adapterType)
	}

	return adapter.Generate()
}

// GenerateSecondary generates a secondary (outbound) adapter.
// For database adapters, entityName (optional) drives the sub-package and entity wiring.
// portInfo (optional) provides method signatures for code generation.
func (g *AdapterGenerator) GenerateSecondary(adapterType, adapterName, entityName, portName string, portInfo *analyzer.PortInfo) error {
	// database and cache have dedicated templates; any other string (external,
	// tool, workspace, filesystem, ...) maps to the generic port-impl template.
	var (
		adapterDir, pkgName string
	)

	var err error
	if adapterType == "database" {
		adapterDir, pkgName, err = g.generateDatabaseAdapterFiles(entityName, adapterName)
	} else {
		adapterDir, pkgName, err = g.generateOtherAdapterFiles(adapterType, adapterName)
	}

	if err != nil {
		return err
	}

	filePath := filepath.Join(adapterDir, pkgName+".go")
	// testFilePath := filepath.Join(adapterDir, pkgName+"_test.go")

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	switch adapterType {
	case "database":
		err = g.generateDatabaseAdapter(filePath, adapterName, entityName, portName)
	case "cache":
		err = g.generateCacheAdapter(filePath, adapterName, portInfo)
	default:
		// external + any free-form type → generic secondary port impl
		err = g.generateGenericAdapter(filePath, adapterType, adapterName, portName, portInfo)
	}

	return err
}

func (g *AdapterGenerator) generateDatabaseAdapterFiles(entityName, adapterName string) (string, string, error) {
	var adapterDir, pkgName string

	// Always use sub-package for database adapters
	if entityName != "" {
		pkgName = utils.ToPlural(strings.ToLower(entityName))
	} else {
		pkgName = strings.ToLower(adapterName)
	}

	adapterDir = filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), "database", pkgName)
	if err := utils.CreateDir(adapterDir); err != nil {
		return "", "", err
	}
	return adapterDir, pkgName, nil
}

func (g *AdapterGenerator) generateOtherAdapterFiles(adapterType, adapterName string) (string, string, error) {
	var adapterDir, pkgName string
	adapterDir = filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), adapterType)
	if err := utils.CreateDir(adapterDir); err != nil {
		return "", "", err
	}
	pkgName = utils.ToSnakeCase(adapterName)

	return adapterDir, pkgName, nil
}

// generateDatabaseAdapter generates a database repository adapter
func (g *AdapterGenerator) generateDatabaseAdapter(filePath, repoName, entityName, portName string) error {
	// Ensure ErrNotFound exists in domain before generating adapter
	if err := g.EnsureDomainError("ErrNotFound", "entity not found"); err != nil {
		return err
	}

	// Derive entity-related template variables
	var resolvedEntity, pkgName, entityImportAlias string
	hasEntity := entityName != ""
	if hasEntity {
		resolvedEntity = entityName
		pkgName = utils.ToPlural(strings.ToLower(entityName))
	} else {
		resolvedEntity = ""
		pkgName = strings.ToLower(repoName)
	}
	entityImportAlias = pkgName + "Domain"

	data := map[string]any{
		"ModuleName":        g.config.Project.Module,
		"PackageName":       pkgName,
		"RepoName":          repoName,
		"HasEntity":         hasEntity,
		"EntityName":        resolvedEntity,
		"EntityPackage":     pkgName,
		"EntityImportAlias": entityImportAlias,
	}

	content, err := g.config.TemplateLoader.Render("adapter/secondary/database.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render database adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// FIXME: adapter don't get it's own folder and package
// command: hexago add adapter secondary external QRClient
// creates:
//
//	internal/adapters/secondary/external/q_r_client.go, it should create it's own folder

// generateGenericAdapter generates a generic secondary adapter (port implementation).
// Used for "external" clients and any free-form adapter type (tool, workspace,
// filesystem, ...). When a port is provided via --from-port it infers method
// signatures and emits a compile-time interface assertion.
func (g *AdapterGenerator) generateGenericAdapter(filePath, adapterType, serviceName, portName string, portInfo *analyzer.PortInfo) error {
	data := map[string]any{
		"PackageName": adapterType,
		"ServiceName": serviceName,
	}

	// Set PortImport if a port was actually discovered (via --from-port).
	if portInfo != nil {
		if portInfo.Name != "" {
			data["PortName"] = portInfo.Name
		}

		// Use the port's real package path so the import resolves regardless of
		// whether the project uses explicit ports or per-entity port files.
		portImport := portInfo.ImportPath
		if portImport == "" {
			portImport = g.config.Project.Module + "/internal/core/ports/outbound"
		}
		data["PortImport"] = fmt.Sprintf("%q", portImport)

		// Collect domain imports from method parameters
		domainAliasMap := portInfo.DomainAliasMap(g.config.Project.Module)

		// Process methods with prefixed types
		data["Methods"] = processMethodsWithPrefix(portInfo.Methods, domainAliasMap)
		data["DomainImports"] = domainAliasMap
	}

	content, err := g.config.TemplateLoader.Render("adapter/secondary/generic.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render generic adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateCacheAdapter generates a cache adapter
func (g *AdapterGenerator) generateCacheAdapter(filePath, cacheName string, portInfo *analyzer.PortInfo) error {
	// Ensure ErrNotFound exists in domain before generating adapter
	if err := g.EnsureDomainError("ErrNotFound", "entity not found"); err != nil {
		return err
	}

	data := map[string]any{
		"CacheName": cacheName,
	}

	if portInfo != nil {
		if portInfo.Name != "" {
			data["PortName"] = portInfo.Name
		}
		portImport := portInfo.ImportPath
		if portImport == "" {
			portImport = g.config.Project.Module + "/internal/core/ports/outbound"
		}
		data["PortImport"] = fmt.Sprintf("%q", portImport)
	}

	content, err := g.config.TemplateLoader.Render("adapter/secondary/cache.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render cache adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// EnsureDomainError ensures an error exists in domain/errors.go.
// If the file doesn't exist, create it with the error.
// If it exists, use go/packages to check if error is already defined.
func (g *AdapterGenerator) EnsureDomainError(errorName, errorMessage string) error {
	errorsFile := filepath.Join(g.config.OutputDir, "internal", "core", "domain", "errors.go")

	if !utils.FileExists(errorsFile) {
		return g.createErrorsFile(errorsFile, errorName, errorMessage)
	}

	if g.isErrorDefined(errorsFile, errorName) {
		return nil
	}

	return g.appendErrorToFile(errorsFile, errorName, errorMessage)
}

// createErrorsFile creates a new domain/errors.go file with the given error.
func (g *AdapterGenerator) createErrorsFile(filePath, errorName, errorMessage string) error {
	data := map[string]any{
		"ErrorName":        errorName,
		"ErrorMessage":     errorMessage,
		"ErrorDescription": strings.ToLower(errorMessage),
	}

	content, err := g.config.TemplateLoader.Render("domain/errors.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render errors template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// isErrorDefined checks if an error with the given name is already defined in the file.
func (g *AdapterGenerator) isErrorDefined(filePath, errorName string) bool {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes,
	}

	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		return false
	}

	if len(pkgs) == 0 {
		return false
	}

	for _, syn := range pkgs[0].Syntax {
		for _, decl := range syn.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}

			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)

				for _, ident := range vs.Names {
					if ident.Name == errorName {
						return true
					}
				}
			}
		}
	}

	return false
}

// appendErrorToFile appends a new error to an existing errors.go file.
func (g *AdapterGenerator) appendErrorToFile(filePath, errorName, errorMessage string) error {
	content, err := utils.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read errors file: %w", err)
	}

	newError := fmt.Sprintf("\n// %s is returned when %s.\nvar %s = errors.New(\"%s\")",
		errorName, strings.ToLower(errorMessage), errorName, errorMessage)

	newContent := strings.TrimSpace(content) + newError + "\n"

	return utils.WriteFile(filePath, []byte(newContent))
}

// processMethodsWithPrefix adds import alias prefix to types that need it.
func processMethodsWithPrefix(methods []analyzer.MethodInfo, aliasMap map[string]string) []analyzer.MethodInfo {
	result := make([]analyzer.MethodInfo, len(methods))
	for i, method := range methods {
		newParams := make([]analyzer.ParamInfo, len(method.Params))
		for j, param := range method.Params {
			newParams[j] = analyzer.ParamInfo{
				Name:       param.Name,
				Type:       prefixType(param.Type, param.ImportPath, aliasMap),
				ImportPath: param.ImportPath,
			}
		}
		newReturns := make([]analyzer.ParamInfo, len(method.Returns))
		for j, ret := range method.Returns {
			newReturns[j] = analyzer.ParamInfo{
				Name:       ret.Name,
				Type:       prefixType(ret.Type, ret.ImportPath, aliasMap),
				ImportPath: ret.ImportPath,
			}
		}
		result[i] = analyzer.MethodInfo{
			Name:    method.Name,
			Params:  newParams,
			Returns: newReturns,
		}
	}
	return result
}

// AdapterInboundPackageRelPath returns the relative package path for a primary (inbound) adapter,
// matching the directory layout used by GeneratePrimary.
func (g *AdapterGenerator) AdapterInboundPackageRelPath(adapterType, adapterName, entityName string) string {
	if adapterType == "http" && entityName != "" {
		pkgName := utils.ToPlural(strings.ToLower(entityName))
		return filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", pkgName)
	}
	return filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), adapterType)
}

// AdapterOutboundPackageRelPath returns the relative package path for a secondary (outbound) adapter,
// matching the directory layout used by GenerateSecondary.
func (g *AdapterGenerator) AdapterOutboundPackageRelPath(adapterType, adapterName, entityName string) string {
	if adapterType == "database" {
		var pkgName string
		if entityName != "" {
			pkgName = utils.ToPlural(strings.ToLower(entityName))
		} else {
			pkgName = strings.ToLower(adapterName)
		}
		return filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), "database", pkgName)
	}
	return filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), adapterType)
}

// prefixType adds the import alias prefix to a type if it's from an external package.
func prefixType(typeStr, importPath string, aliasMap map[string]string) string {
	if importPath == "" {
		return typeStr
	}
	alias, ok := aliasMap[importPath]
	if !ok {
		return typeStr
	}
	// Handle pointer types
	if strings.HasPrefix(typeStr, "*") {
		return "*" + alias + "." + strings.TrimPrefix(typeStr, "*")
	}
	// Handle slice types
	if strings.HasPrefix(typeStr, "[]") {
		elem := strings.TrimPrefix(typeStr, "[]")
		return "[]" + alias + "." + elem
	}
	// Handle other types
	return alias + "." + typeStr
}
