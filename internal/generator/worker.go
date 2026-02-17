package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Type      string // queue, periodic, event
	Interval  string // for periodic workers
	Workers   int    // number of goroutines for queue workers
	QueueSize int    // queue size for queue workers
}

// WorkerGenerator generates worker files
type WorkerGenerator struct {
	config *ProjectConfig
}

// NewWorkerGenerator creates a new worker generator
func NewWorkerGenerator(config *ProjectConfig) *WorkerGenerator {
	return &WorkerGenerator{
		config: config,
	}
}

// Generate creates worker files
func (g *WorkerGenerator) Generate(workerName string, workerConfig WorkerConfig) error {
	// Create workers directory if it doesn't exist
	workersDir := filepath.Join("internal", "workers")
	if err := fileutil.CreateDir(workersDir); err != nil {
		return err
	}

	fileName := toSnakeCase(workerName) + ".go"
	testFileName := toSnakeCase(workerName) + "_test.go"

	filePath := filepath.Join(workersDir, fileName)
	testFilePath := filepath.Join(workersDir, testFileName)

	if fileutil.FileExists(filePath) {
		return fmt.Errorf("worker file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating worker file: %s\n", filePath)

	// Generate worker based on type
	switch workerConfig.Type {
	case "queue":
		if err := g.generateQueueWorker(filePath, workerName, workerConfig); err != nil {
			return err
		}
	case "periodic":
		if err := g.generatePeriodicWorker(filePath, workerName, workerConfig); err != nil {
			return err
		}
	case "event":
		if err := g.generateEventWorker(filePath, workerName, workerConfig); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported worker type: %s", workerConfig.Type)
	}

	fmt.Printf("📝 Creating test file: %s\n", testFilePath)

	// Generate test file
	if err := g.generateWorkerTestFile(testFilePath, workerName); err != nil {
		return err
	}

	// Generate or update worker manager
	if err := g.ensureWorkerManager(workersDir); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to ensure worker manager: %v\n", err)
	}

	return nil
}

// generateQueueWorker generates a queue-based worker
func (g *WorkerGenerator) generateQueueWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
		"Workers":    config.Workers,
		"QueueSize":  config.QueueSize,
	}

	content, err := globalTemplateLoader.Render("worker/queue.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render queue worker template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generatePeriodicWorker generates a periodic worker
func (g *WorkerGenerator) generatePeriodicWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
		"Interval":   config.Interval,
	}

	content, err := globalTemplateLoader.Render("worker/periodic.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render periodic worker template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateEventWorker generates an event-driven worker
func (g *WorkerGenerator) generateEventWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
	}

	content, err := globalTemplateLoader.Render("worker/event.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render event worker template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateWorkerTestFile generates test file for worker
func (g *WorkerGenerator) generateWorkerTestFile(filePath, workerName string) error {
	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
	}

	content, err := globalTemplateLoader.Render("worker/worker_test.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render worker test template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// ensureWorkerManager creates or updates the worker manager
func (g *WorkerGenerator) ensureWorkerManager(workersDir string) error {
	managerPath := filepath.Join(workersDir, "manager.go")

	// If manager already exists, don't overwrite
	if fileutil.FileExists(managerPath) {
		fmt.Printf("ℹ️  Worker manager already exists: %s\n", managerPath)
		return nil
	}

	fmt.Printf("📝 Creating worker manager: %s\n", managerPath)

	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
	}

	content, err := globalTemplateLoader.Render("worker/manager.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render worker manager template: %w", err)
	}

	return fileutil.WriteFile(managerPath, content)
}
