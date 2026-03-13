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
	config      *ProjectConfig
	projectPath string
}

// NewProjectGenerator creates a new ProjectGenerator
func NewProjectGenerator(config *ProjectConfig) *ProjectGenerator {
	return &ProjectGenerator{
		config: config,
	}
}

// Generate creates the complete project structure
func (g *ProjectGenerator) Generate() error {
	// var projectPath string
	if g.config.InPlace {
		g.projectPath = g.config.OutputDir
	} else {
		g.projectPath = filepath.Join(g.config.OutputDir, g.config.ProjectName)
		// Check if directory already exists (in-place always uses an existing dir)
		if utils.FileExists(g.projectPath) {
			return fmt.Errorf("directory %s already exists", g.projectPath)
		}
	}

	fmt.Printf("🚀 Generating project %s...\n", g.config.ProjectName)

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

	// Write .hexago.yaml to persist init-time settings
	if err := g.saveHexagoConfig(); err != nil {
		fmt.Printf("⚠️  Warning: failed to write .hexago.yaml: %v\n", err)
		// non-fatal — project is still fully usable
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
	}

	// Add optional directories
	if g.config.ExplicitPorts {
		dirs = append(dirs,
			"internal/core/ports/inbound",
			"internal/core/ports/outbound",
		)
	}

	if g.config.WithObservability {
		dirs = append(dirs, "internal/observability")
	}

	if g.config.WithWorkers {
		dirs = append(dirs, "internal/workers")
	}

	if g.config.WithMigrations {
		dirs = append(dirs, "migrations")
	}

	// Create all directories
	return utils.CreateDirs(g.projectPath, dirs)
}

// generateFiles generates all files from templates
func (g *ProjectGenerator) generateFiles() error {
	fmt.Println("📝 Generating files...")

	// Generate main.go
	if err := g.generateFile(mainTemplate); err != nil {
		return err
	}

	// Generate cmd/root.go
	if err := g.generateFile(rootTemplate); err != nil {
		return err
	}

	// Generate cmd/run.go
	if err := g.generateFile(runTemplate); err != nil {
		return err
	}

	switch g.config.ProjectType {

	// Generate processor for service type
	case "service":
		if err := g.generateFile(processorTemplate); err != nil {
			return err
		}

		// Generate pkg/httpserver and adapter wiring (http-server type only)
	case "http-server":
		if err := g.generateFile(httpServerInterfaceTemplate); err != nil {
			return err
		}

		if err := g.generateFile(httpServerFileTemplate); err != nil {
			return err
		}

		if err := g.generateFile(httpAdapterTemplate); err != nil {
			return err
		}

		if err := g.generateFile(httpPingTemplate); err != nil {
			return err
		}
	}

	// Generate config
	if err := g.generateFile(configTemplate); err != nil {
		return err
	}

	// Generate logger
	if err := g.generateFile(loggerTemplate); err != nil {
		return err
	}

	// Generate Makefile
	if err := g.generateFile(makefileTemplate); err != nil {
		return err
	}

	// Generate .gitignore
	if err := g.generateFile(gitignoreTemplate); err != nil {
		return err
	}

	// Generate README
	if err := g.generateFile(readmeTemplate); err != nil {
		return err
	}

	// Optional files
	if g.config.WithDocker {
		// Generate Dockerfile
		if err := g.generateFile(dockerFileTemplate); err != nil {
			return err
		}
		// Generate compose.yaml
		if err := g.generateFile(composeTemplate); err != nil {
			return err
		}
	}

	if g.config.WithObservability {
		// Generate internal/observability/health.go
		if err := g.generateFile(healthTemplate); err != nil {
			return err
		}
		// Generate internal/observability/metrics.go
		if err := g.generateFile(metricsTemplate); err != nil {
			return err
		}
		// Generate route handlers for health and metrics (http-server only)
		if g.config.ProjectType == "http-server" {
			if err := g.generateFile(httpHealthTemplate); err != nil {
				return err
			}
			if err := g.generateFile(httpMetricsTemplate); err != nil {
				return err
			}
		}
	}

	return nil
}

// initGoModule initializes the go.mod file
func (g *ProjectGenerator) initGoModule() error {
	fmt.Println("📦 Initializing go module...")

	cmd := exec.Command("go", "mod", "init", g.config.ModuleName)
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
// TODO: make dependency list configurable/updatable
func (g *ProjectGenerator) addDependencies() error {
	fmt.Println("📦 Adding dependencies...")

	dependencies := []string{
		"github.com/spf13/cobra@latest",
		"github.com/spf13/viper@latest",
	}

	// Add framework-specific dependencies
	switch g.config.Framework {
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
	if g.config.WithMetrics || g.config.WithObservability {
		dependencies = append(dependencies, "github.com/prometheus/client_golang@latest")
	}

	// Fiber needs the adaptor package to wrap net/http handlers
	if g.config.Framework == "fiber" && g.config.WithObservability {
		dependencies = append(dependencies, "github.com/gofiber/adaptor/v2@latest")
	}

	for _, dep := range dependencies {
		cmd := exec.Command("go", "get", dep)
		cmd.Dir = g.projectPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: failed to add dependency %s: %v\n", dep, err)
		}
	}

	return nil
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
	cfg := HexagoConfigFromProject(g.config)
	return SaveHexagoConfig(g.projectPath, cfg)
}

// printSuccess prints success message with next steps
func (g *ProjectGenerator) printSuccess() {
	fmt.Println("\n✅ Project generated successfully!")
	fmt.Println("\n📚 Next steps:")
	fmt.Printf("  cd %s\n", g.config.ProjectName)
	fmt.Println("  go run main.go run")
	fmt.Println("\n📖 Read the README.md for more information about the project structure.")
}
