package generator

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/utils"
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
	if err := utils.CreateDir(workersDir); err != nil {
		return err
	}

	fileName := utils.ToSnakeCase(workerName) + ".go"

	filePath := filepath.Join(workersDir, fileName)

	if utils.FileExists(filePath) {
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

	// Generate or update worker manager
	if err := g.ensureWorkerManager(workersDir); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to ensure worker manager: %v\n", err)
	}

	return nil
}

// generateQueueWorker generates a queue-based worker
func (g *WorkerGenerator) generateQueueWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]any{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
		"Workers":    config.Workers,
		"QueueSize":  config.QueueSize,
	}

	content, err := g.config.templateLoader.Render("worker/queue.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render queue worker template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generatePeriodicWorker generates a periodic worker
func (g *WorkerGenerator) generatePeriodicWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]any{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
		"Interval":   config.Interval,
	}

	content, err := g.config.templateLoader.Render("worker/periodic.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render periodic worker template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// generateEventWorker generates an event-driven worker
func (g *WorkerGenerator) generateEventWorker(filePath, workerName string, config WorkerConfig) error {
	data := map[string]any{
		"ModuleName": g.config.ModuleName,
		"WorkerName": workerName,
	}

	content, err := g.config.templateLoader.Render("worker/event.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render event worker template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// ensureWorkerManager creates or updates the worker manager
func (g *WorkerGenerator) ensureWorkerManager(workersDir string) error {
	managerPath := filepath.Join(workersDir, "manager.go")

	// If manager already exists, don't overwrite
	if utils.FileExists(managerPath) {
		fmt.Printf("ℹ️  Worker manager already exists: %s\n", managerPath)
		return nil
	}

	fmt.Printf("📝 Creating worker manager: %s\n", managerPath)

	data := map[string]any{
		"ModuleName": g.config.ModuleName,
	}

	content, err := g.config.templateLoader.Render("worker/manager.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render worker manager template: %w", err)
	}

	return utils.WriteFile(managerPath, content)
}
