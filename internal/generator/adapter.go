package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/pkg/utils"
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

	if err := g.generateAdapterTestFile(testFilePath, adapterName, adapterType); err != nil {
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
func (g *AdapterGenerator) GenerateSecondary(adapterType, adapterName, entityName, portName string) error {
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
		filePath = filepath.Join(adapterDir, pkgName+".go")
		testFilePath = filepath.Join(adapterDir, pkgName+"_test.go")
	} else {
		adapterDir = filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), adapterType)
		if err := utils.CreateDir(adapterDir); err != nil {
			return err
		}
		filePath = filepath.Join(adapterDir, utils.ToSnakeCase(adapterName)+".go")
		testFilePath = filepath.Join(adapterDir, utils.ToSnakeCase(adapterName)+"_test.go")
	}

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	// Generate port interface if using explicit ports
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
		if err := g.generateExternalAdapter(filePath, adapterName, portName); err != nil {
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

	if err := g.generateAdapterTestFile(testFilePath, adapterName, adapterType); err != nil {
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

// generateExternalAdapter generates an external service adapter
func (g *AdapterGenerator) generateExternalAdapter(filePath, serviceName, portName string) error {
	data := map[string]any{
		"ServiceName": serviceName,
	}

	content, err := g.config.templateLoader.Render("adapter/external.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render external adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateCacheAdapter generates a cache adapter
func (g *AdapterGenerator) generateCacheAdapter(filePath, cacheName, portName string) error {
	data := map[string]any{
		"CacheName": cacheName,
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
func (g *AdapterGenerator) generateAdapterTestFile(filePath, adapterName, adapterType string) error {
	data := map[string]any{
		"Package":     adapterType,
		"AdapterName": adapterName,
	}

	content, err := g.config.templateLoader.Render("adapter/adapter_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render adapter test template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
