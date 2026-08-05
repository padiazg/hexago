package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/utils"
)

// ProjectGenerator handles the generation of new projects
type ProjectGenerator struct {
	config      HexagoConfig
	projectPath string
}

// NewProjectGenerator creates a new ProjectGenerator
func NewProjectGenerator(config HexagoConfig) ProjectGenerator {
	return ProjectGenerator{
		config: config,
	}
}

// Generate creates the complete project structure
func (g *ProjectGenerator) Generate() error {
	// var projectPath string
	if g.config.InPlace {
		g.projectPath = g.config.OutputDir
	} else {
		g.projectPath = filepath.Join(g.config.OutputDir, g.config.Project.Name)
		// Check if directory already exists (in-place always uses an existing dir)
		if utils.FileExists(g.projectPath) {
			return fmt.Errorf("directory %s already exists", g.projectPath)
		}
	}

	fmt.Printf("🚀 Generating project %s...\n", g.config.Project.Name)

	// Create base directory (no-op when in-place, dir already exists)
	if err := utils.CreateDir(g.projectPath); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate directory structure
	if err := g.generateDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate files from templates
	if err := g.generateFiles(); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Generate migration manager if migrations are enabled
	if g.config.Features.WithMigrations {
		migGen := NewMigrationGenerator(&g.config, g.projectPath)
		if err := migGen.EnsureMigrationManager(); err != nil {
			fmt.Printf("⚠️  Warning: failed to create migration manager: %v\n", err)
		}
	}

	// Write .hexago.yaml to persist init-time settings.
	// Written before go mod tidy so the config survives even if dependency
	// resolution fails (tidy is the step most likely to abort generation).
	if err := g.saveHexagoConfig(); err != nil {
		fmt.Printf("⚠️  Warning: failed to write .hexago.yaml: %v\n", err)
		// non-fatal — project is still fully usable
	}

	// Initialize go.mod
	if err := g.initGoModule(); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	// Run go mod tidy
	if err := g.runGoModTidy(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	// Format generated code
	if err := g.formatCode(); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to format code: %v\n", err)
	}

	g.printSuccess()
	return nil
}

// generateDirectoryStructure creates the directory structure
func (g *ProjectGenerator) generateDirectoryStructure() error {
	fmt.Println("📁 Creating directory structure...")

	dirs := []string{
		"cmd",
		"internal/core/domain",
		fmt.Sprintf("internal/core/%s", g.config.CoreLogicDir()),
		fmt.Sprintf("internal/adapters/%s/http", g.config.AdapterInboundDir()),
		fmt.Sprintf("internal/adapters/%s/database", g.config.AdapterOutboundDir()),
		"internal/config",
		"pkg/logger",
		"pkg/version",
	}

	// Add optional directories
	if g.config.Structure.ExplicitPorts {
		dirs = append(dirs,
			"internal/core/ports/inbound",
			"internal/core/ports/outbound",
		)
	}

	if g.config.Features.WithObservability {
		dirs = append(dirs, "internal/observability")
	}

	if g.config.Features.WithWorkers {
		dirs = append(dirs, "internal/workers")
	}

	if g.config.Features.WithMigrations {
		dirs = append(dirs, "migrations")
	}

	// Create all directories
	return utils.CreateDirs(g.projectPath, dirs)
}

// generateFiles generates all files from templates
func (g *ProjectGenerator) generateFiles() error {
	fmt.Println("📝 Generating files...")
	queue := []string{
		mainTemplate,
		rootTemplate,
		versionCmdTemplate,
	}

	switch g.config.Project.Type {

	// Long-running service/daemon: run command + processor
	case "service":
		queue = append(queue, runTemplate, processorTemplate)

	// Batch CLI: no run command, subcommands added via `hexago add adapter primary cli`
	case "cli":
		// no run/processor

	// Generate pkg/httpserver and adapter wiring (http-server type only)
	case "http-server":
		queue = append(queue, runTemplate)
		queue = append(queue, []string{
			servicesStubTemplate,
			httpServerInterfaceTemplate,
			httpServerFileTemplate,
			httpAdapterTemplate,
			httpPingTemplate,
		}...)

	}

	queue = append(queue, []string{
		configTemplate,        // Generate config
		loggerTemplate,        // Generate logger
		versionTemplate,       // Generate version package
		versionSplashTemplate, // Generate splash
		versionTestTemplate,   // Generate version test
		makefileTemplate,      // Generate Makefile
		gitignoreTemplate,     // Generate .gitignore
		readmeTemplate,        // Generate README
	}...)

	// Docker files
	if g.config.Features.WithDocker {
		queue = append(queue, []string{
			dockerFileTemplate, // Generate Dockerfile
			composeTemplate,    // Generate compose.yaml
		}...)
	}

	// Observability files
	if g.config.Features.WithObservability {
		queue = append(queue, []string{
			healthTemplate,  // Generate internal/observability/health.go
			metricsTemplate, // Generate internal/observability/metrics.go
		}...)

		// Generate standalone server for service type (http-server mounts on main HTTP server)
		if g.config.Project.Type == "service" {
			queue = append(queue, serverTemplate)
		}

		// Generate route handlers for health and metrics (http-server only)
		if g.config.Project.Type == "http-server" {
			queue = append(queue, []string{
				httpHealthTemplate,
				httpMetricsTemplate,
			}...)
		}
	}

	for _, templ := range queue {
		if err := g.generateFile(templ); err != nil {
			return err
		}
	}

	return nil
}

// initGoModule initializes the go.mod file
func (g *ProjectGenerator) initGoModule() error {
	fmt.Println("📦 Initializing go module...")

	cmd := exec.Command("go", "mod", "init", g.config.Project.Module)
	cmd.Dir = g.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}

	// Add required dependencies
	return g.addDependencies()
}

// addDependencies adds required dependencies to go.mod
func (g *ProjectGenerator) addDependencies() error {
	fmt.Println("📦 Adding dependencies...")

	for _, dep := range g.dependenciesList() {
		cmd := exec.Command("go", "get", dep)
		cmd.Dir = g.projectPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: failed to add dependency %s: %v\n", dep, err)
		}
	}

	return nil
}

func (g *ProjectGenerator) dependenciesList() []string {
	// TODO: make dependency list configurable/updatable
	dependencies := []string{
		"github.com/spf13/cobra@latest",
		"github.com/spf13/viper@latest",
	}

	// Add framework-specific dependencies
	switch g.config.Project.Framework {
	case "echo":
		dependencies = append(dependencies, "github.com/labstack/echo/v4@latest")
	case "gin":
		dependencies = append(dependencies, "github.com/gin-gonic/gin@latest")
	case "chi":
		dependencies = append(dependencies, "github.com/go-chi/chi/v5@latest")
	case "fiber":
		dependencies = append(dependencies, "github.com/gofiber/fiber/v2@latest")
	}

	// Add metrics/observability dependencies
	if g.config.Features.WithMetrics || g.config.Features.WithObservability {
		dependencies = append(dependencies, "github.com/prometheus/client_golang@latest")
	}

	// Add errgroup for observability server lifecycle (service type only)
	if g.config.Features.WithObservability {
		dependencies = append(dependencies, "golang.org/x/sync@latest")
	}

	// Fiber needs the adaptor package to wrap net/http handlers
	if g.config.Project.Framework == "fiber" && g.config.Features.WithObservability {
		dependencies = append(dependencies, "github.com/gofiber/adaptor/v2@latest")
	}

	// Add database migration dependencies
	if g.config.Features.WithMigrations {
		dependencies = append(dependencies, "github.com/golang-migrate/migrate/v4@latest")
	}

	return dependencies
}

// runGoModTidy runs go mod tidy
func (g *ProjectGenerator) runGoModTidy() error {
	fmt.Println("🧹 Running go mod tidy...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = g.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

// formatCode runs go fmt on the generated code
func (g *ProjectGenerator) formatCode() error {
	fmt.Println("✨ Formatting code...")

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = g.projectPath

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// saveHexagoConfig writes .hexago.yaml with the current project settings.
func (g *ProjectGenerator) saveHexagoConfig() error {
	return SaveHexagoConfig(g.projectPath, &g.config)
}

// printSuccess prints success message with next steps
func (g *ProjectGenerator) printSuccess() {
	fmt.Println("\n✅ Project generated successfully!")
	fmt.Println("\n📚 Next steps:")
	fmt.Printf("  cd %s\n", g.config.Project.Name)
	nextCmd := "go run main.go run"
	if g.config.IsCLI() {
		nextCmd = "go run main.go version"
	}
	fmt.Println("  " + nextCmd)
	fmt.Println("\n📖 Read the README.md for more information about the project structure.")
}
