package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
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

// GeneratePrimary generates a primary (inbound) adapter
func (g *AdapterGenerator) GeneratePrimary(adapterType, adapterName, portName string) error {
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

	// Determine directory
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), adapterType)

	// Create directory if it doesn't exist
	if err := fileutil.CreateDir(adapterDir); err != nil {
		return err
	}

	fileName := utils.ToSnakeCase(adapterName) + ".go"
	testFileName := utils.ToSnakeCase(adapterName) + "_test.go"

	filePath := filepath.Join(adapterDir, fileName)
	testFilePath := filepath.Join(adapterDir, testFileName)

	if fileutil.FileExists(filePath) {
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

// GenerateSecondary generates a secondary (outbound) adapter
func (g *AdapterGenerator) GenerateSecondary(adapterType, adapterName, portName string) error {
	// Validate adapter type
	validTypes := map[string]bool{
		"database": true,
		"external": true,
		"cache":    true,
	}

	if !validTypes[adapterType] {
		return fmt.Errorf("invalid secondary adapter type '%s'. Valid types: database, external, cache", adapterType)
	}

	// Determine directory
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), adapterType)

	// Create directory if it doesn't exist
	if err := fileutil.CreateDir(adapterDir); err != nil {
		return err
	}

	fileName := utils.ToSnakeCase(adapterName) + ".go"
	testFileName := utils.ToSnakeCase(adapterName) + "_test.go"

	filePath := filepath.Join(adapterDir, fileName)
	testFilePath := filepath.Join(adapterDir, testFileName)

	if fileutil.FileExists(filePath) {
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
		if err := g.generateDatabaseAdapter(filePath, adapterName, portName); err != nil {
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
	data := map[string]interface{}{
		"ModuleName":  g.config.ModuleName,
		"CoreLogic":   g.config.CoreLogicDir(),
		"HandlerName": handlerName,
	}

	content, err := g.config.templateLoader.Render("adapter/http.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render HTTP adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateGRPCAdapter generates a gRPC handler adapter
func (g *AdapterGenerator) generateGRPCAdapter(filePath, handlerName string) error {
	data := map[string]interface{}{
		"ModuleName":  g.config.ModuleName,
		"CoreLogic":   g.config.CoreLogicDir(),
		"HandlerName": handlerName,
	}

	content, err := g.config.templateLoader.Render("adapter/grpc.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render gRPC adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateQueueAdapter generates a message queue consumer adapter
func (g *AdapterGenerator) generateQueueAdapter(filePath, consumerName string) error {
	data := map[string]interface{}{
		"ModuleName":   g.config.ModuleName,
		"CoreLogic":    g.config.CoreLogicDir(),
		"ConsumerName": consumerName,
	}

	content, err := g.config.templateLoader.Render("adapter/queue.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render queue adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateDatabaseAdapter generates a database repository adapter
func (g *AdapterGenerator) generateDatabaseAdapter(filePath, repoName, portName string) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"RepoName":   repoName,
	}

	content, err := g.config.templateLoader.Render("adapter/database.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render database adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateExternalAdapter generates an external service adapter
func (g *AdapterGenerator) generateExternalAdapter(filePath, serviceName, portName string) error {
	data := map[string]interface{}{
		"ServiceName": serviceName,
	}

	content, err := g.config.templateLoader.Render("adapter/external.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render external adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateCacheAdapter generates a cache adapter
func (g *AdapterGenerator) generateCacheAdapter(filePath, cacheName, portName string) error {
	data := map[string]interface{}{
		"CacheName": cacheName,
	}

	content, err := g.config.templateLoader.Render("adapter/cache.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render cache adapter template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generatePortInterface generates a port interface (if using explicit ports)
func (g *AdapterGenerator) generatePortInterface(portName, adapterName string) error {
	// This would generate the port interface in internal/core/ports/
	// For now, skip implementation as it's optional
	return nil
}

// generateAdapterTestFile generates test file for adapters
func (g *AdapterGenerator) generateAdapterTestFile(filePath, adapterName, adapterType string) error {
	data := map[string]interface{}{
		"Package":     adapterType,
		"AdapterName": adapterName,
	}

	content, err := g.config.templateLoader.Render("adapter/adapter_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render adapter test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}
