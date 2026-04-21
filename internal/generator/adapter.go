package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/internal/analyzer"
	"github.com/padiazg/hexago/pkg/utils"
	"golang.org/x/tools/go/packages"
)

// AdapterGenerator generates adapter files
type AdapterGenerator struct {
	config *ProjectConfig
}

// NewAdapterGenerator creates a new adapter generator
func NewAdapterGenerator(config *ProjectConfig) *AdapterGenerator {
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

	// HTTP + entity → sub-package with two files
	if adapterType == "http" && entityName != "" {
		return g.generateHTTPHandlerPackage(adapterName, entityName)
	}

	// Default: flat directory
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), adapterType)
	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	fileName := utils.ToSnakeCase(adapterName) + ".go"
	testFileName := utils.ToSnakeCase(adapterName) + "_test.go"
	filePath := filepath.Join(adapterDir, fileName)
	testFilePath := filepath.Join(adapterDir, testFileName)

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	switch adapterType {
	case "http":
		if err := g.generateHTTPAdapter(filePath, adapterName); err != nil {
			return err
		}
	case "grpc":
		if err := g.generateGRPCAdapter(filePath, adapterName); err != nil {
			return err
		}
	case "queue":
		if err := g.generateQueueAdapter(filePath, adapterName); err != nil {
			return err
		}
	default:
		return fmt.Errorf("adapter type %s not yet implemented", adapterType)
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	if err := g.generateAdapterTestFile(testFilePath, adapterName, adapterType, nil); err != nil {
		return err
	}

	return nil
}

// generateHTTPHandlerPackage generates the two-file per-entity HTTP handler sub-package.
func (g *AdapterGenerator) generateHTTPHandlerPackage(adapterName, entityName string) error {
	pkgName := utils.ToPlural(strings.ToLower(entityName))
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", pkgName)

	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	configFile := filepath.Join(adapterDir, utils.ToSnakeCase(entityName)+".go")
	handlersFile := filepath.Join(adapterDir, "handlers.go")

	if utils.FileExists(configFile) {
		return fmt.Errorf("handler file %s already exists", configFile)
	}

	entityVarName := strings.ToLower(entityName[:1]) + entityName[1:]
	servicePkgName := pkgName
	serviceImportAlias := pkgName + "Svc"
	entityImportAlias := pkgName + "Domain"
	serviceField := utils.ToTitleCase(pkgName)
	routePrefix := pkgName

	data := map[string]any{
		"ModuleName":         g.config.ModuleName,
		"CoreLogic":          g.config.CoreLogicDir(),
		"PackageName":        pkgName,
		"EntityName":         entityName,
		"EntityVarName":      entityVarName,
		"EntityPackage":      pkgName,
		"EntityImportAlias":  entityImportAlias,
		"ServicePackage":     servicePkgName,
		"ServiceImportAlias": serviceImportAlias,
		"ServiceName":        entityName,
		"ServiceField":       serviceField,
		"RoutePrefix":        routePrefix,
	}

	framework := g.config.Framework
	if framework == "" {
		framework = "chi"
	}

	fmt.Printf("📝 Creating handler config file: %s\n", configFile)
	configTmpl := fmt.Sprintf("adapter/primary/http/%s/handler_config.go.tmpl", framework)
	configContent, err := g.config.templateLoader.Render(configTmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render handler config template: %w", err)
	}
	if err := utils.WriteFile(configFile, configContent); err != nil {
		return err
	}

	fmt.Printf("📝 Creating handler methods file: %s\n", handlersFile)
	methodsTmpl := fmt.Sprintf("adapter/primary/http/%s/handler_methods.go.tmpl", framework)
	methodsContent, err := g.config.templateLoader.Render(methodsTmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render handler methods template: %w", err)
	}
	return utils.WriteFile(handlersFile, methodsContent)
}

// GenerateSecondary generates a secondary (outbound) adapter.
// For database adapters, entityName (optional) drives the sub-package and entity wiring.
// portInfo (optional) provides method signatures for code generation.
func (g *AdapterGenerator) GenerateSecondary(adapterType, adapterName, entityName, portName string, portInfo *analyzer.PortInfo) error {
	// Validate adapter type
	validTypes := map[string]bool{
		"database": true,
		"external": true,
		"cache":    true,
	}

	if !validTypes[adapterType] {
		return fmt.Errorf("invalid secondary adapter type '%s'. Valid types: database, external, cache", adapterType)
	}

	var adapterDir, filePath, testFilePath string

	if adapterType == "database" {
		// Always use sub-package for database adapters
		var pkgName string
		if entityName != "" {
			pkgName = utils.ToPlural(strings.ToLower(entityName))
		} else {
			pkgName = strings.ToLower(adapterName)
		}
		adapterDir = filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), "database", pkgName)
		if err := utils.CreateDir(adapterDir); err != nil {
			return err
		}
		filePath = filepath.Join(g.config.OutputDir, adapterDir, pkgName+".go")
		testFilePath = filepath.Join(g.config.OutputDir, adapterDir, pkgName+"_test.go")
	} else {
		adapterDir = filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), adapterType)
		if err := utils.CreateDir(adapterDir); err != nil {
			return err
		}
		filePath = filepath.Join(g.config.OutputDir, adapterDir, utils.ToSnakeCase(adapterName)+".go")
		testFilePath = filepath.Join(g.config.OutputDir, adapterDir, utils.ToSnakeCase(adapterName)+"_test.go")
	}

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	// Generate port interface if using explicit ports
	// TODO: review this step in this flow. shouldn't ports be created when creating domain entities, or manually if needed
	if g.config.ExplicitPorts && portName != "" {
		if err := g.generatePortInterface(portName, adapterName); err != nil {
			// Non-fatal - just warn
			fmt.Printf("⚠️  Warning: failed to generate port interface: %v\n", err)
		}
	}

	switch adapterType {
	case "database":
		if err := g.generateDatabaseAdapter(filePath, adapterName, entityName, portName); err != nil {
			return err
		}
	case "external":
		if err := g.generateExternalAdapter(filePath, adapterName, portName, portInfo); err != nil {
			return err
		}
	case "cache":
		if err := g.generateCacheAdapter(filePath, adapterName, portName); err != nil {
			return err
		}
	default:
		return fmt.Errorf("adapter type %s not yet implemented", adapterType)
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	// FIXME: wrong import
	// command: hexago add adapter secondary database URLRepository --entity URL
	// creates:
	//  internal/adapters/secondary/database/urls/urls.go       => package urls
	//  internal/adapters/secondary/database/urls/urls_test.go  => package database_test
	if err := g.generateAdapterTestFile(testFilePath, adapterName, adapterType, portInfo); err != nil {
		return err
	}

	return nil
}

// generateHTTPAdapter generates an HTTP handler adapter
func (g *AdapterGenerator) generateHTTPAdapter(filePath, handlerName string) error {
	data := map[string]any{
		"ModuleName":  g.config.ModuleName,
		"CoreLogic":   g.config.CoreLogicDir(),
		"HandlerName": handlerName,
	}

	content, err := g.config.templateLoader.Render("adapter/http.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render HTTP adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateGRPCAdapter generates a gRPC handler adapter
func (g *AdapterGenerator) generateGRPCAdapter(filePath, handlerName string) error {
	data := map[string]any{
		"ModuleName":  g.config.ModuleName,
		"CoreLogic":   g.config.CoreLogicDir(),
		"HandlerName": handlerName,
	}

	content, err := g.config.templateLoader.Render("adapter/grpc.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render gRPC adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateQueueAdapter generates a message queue consumer adapter
func (g *AdapterGenerator) generateQueueAdapter(filePath, consumerName string) error {
	data := map[string]any{
		"ModuleName":   g.config.ModuleName,
		"CoreLogic":    g.config.CoreLogicDir(),
		"ConsumerName": consumerName,
	}

	content, err := g.config.templateLoader.Render("adapter/queue.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render queue adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateDatabaseAdapter generates a database repository adapter
func (g *AdapterGenerator) generateDatabaseAdapter(filePath, repoName, entityName, portName string) error {
	// Ensure ErrNotFound exists in domain before generating adapter
	if err := g.EnsureDomainError("ErrNotFound", "entity not found"); err != nil {
		return err
	}

	// Derive entity-related template variables
	var resolvedEntity, pkgName, entityImportAlias string
	if entityName != "" {
		resolvedEntity = entityName
		pkgName = utils.ToPlural(strings.ToLower(entityName))
	} else {
		resolvedEntity = repoName
		pkgName = strings.ToLower(repoName)
	}
	entityImportAlias = pkgName + "Domain"

	data := map[string]any{
		"ModuleName":        g.config.ModuleName,
		"PackageName":       pkgName,
		"RepoName":          repoName,
		"EntityName":        resolvedEntity,
		"EntityPackage":     pkgName,
		"EntityImportAlias": entityImportAlias,
	}

	content, err := g.config.templateLoader.Render("adapter/database.go.tmpl", data)
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

// generateExternalAdapter generates an external service adapter
func (g *AdapterGenerator) generateExternalAdapter(filePath, serviceName, portName string, portInfo *analyzer.PortInfo) error {
	data := map[string]any{
		"ServiceName": serviceName,
		"PortName":    portName,
	}

	if portInfo != nil {
		data["Methods"] = portInfo.Methods
	}

	content, err := g.config.templateLoader.Render("adapter/external.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render external adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateCacheAdapter generates a cache adapter
func (g *AdapterGenerator) generateCacheAdapter(filePath, cacheName, portName string) error {
	// Ensure ErrNotFound exists in domain before generating adapter
	if err := g.EnsureDomainError("ErrNotFound", "entity not found"); err != nil {
		return err
	}

	data := map[string]any{
		"CacheName": cacheName,
		"PortName":  portName,
	}

	content, err := g.config.templateLoader.Render("adapter/cache.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render cache adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generatePortInterface generates a port interface (if using explicit ports)
func (g *AdapterGenerator) generatePortInterface(portName, adapterName string) error {
	// This would generate the port interface in internal/core/ports/
	// For now, skip implementation as it's optional
	return nil
}

// generateAdapterTestFile generates test file for adapters
func (g *AdapterGenerator) generateAdapterTestFile(filePath, adapterName, adapterType string, portInfo *analyzer.PortInfo) error {
	data := map[string]any{
		"Package":     adapterType,
		"AdapterName": adapterName,
	}

	if portInfo != nil {
		data["Methods"] = portInfo.Methods
		data["PortName"] = portInfo.Name
	}

	content, err := g.config.templateLoader.Render("adapter/adapter_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render adapter test template: %w", err)
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

	content, err := g.config.templateLoader.Render("domain/errors.go.tmpl", data)
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
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

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
