package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// ProjectGenerator handles the generation of new projects
type ProjectGenerator struct {
	config *ProjectConfig
}

// NewProjectGenerator creates a new ProjectGenerator
func NewProjectGenerator(config *ProjectConfig) *ProjectGenerator {
	return &ProjectGenerator{
		config: config,
	}
}

// Generate creates the complete project structure
func (g *ProjectGenerator) Generate() error {
	projectPath := filepath.Join(g.config.OutputDir, g.config.ProjectName)

	// Check if directory already exists
	if fileutil.FileExists(projectPath) {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	fmt.Printf("🚀 Generating project %s...\n", g.config.ProjectName)

	// Create base directory
	if err := fileutil.CreateDir(projectPath); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate directory structure
	if err := g.generateDirectoryStructure(projectPath); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate files from templates
	if err := g.generateFiles(projectPath); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Initialize go.mod
	if err := g.initGoModule(projectPath); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	// Run go mod tidy
	if err := g.runGoModTidy(projectPath); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	// Format generated code
	if err := g.formatCode(projectPath); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to format code: %v\n", err)
	}

	g.printSuccess(projectPath)
	return nil
}

// generateDirectoryStructure creates the directory structure
func (g *ProjectGenerator) generateDirectoryStructure(projectPath string) error {
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
	return fileutil.CreateDirs(projectPath, dirs)
}

// generateFiles generates all files from templates
func (g *ProjectGenerator) generateFiles(projectPath string) error {
	fmt.Println("📝 Generating files...")

	// Generate main.go
	if err := g.generateMainFile(projectPath); err != nil {
		return err
	}

	// Generate cmd/root.go
	if err := g.generateRootCommand(projectPath); err != nil {
		return err
	}

	// Generate cmd/run.go
	if err := g.generateRunCommand(projectPath); err != nil {
		return err
	}

	// Generate internal/adapters/{inbound}/http/server.go (http-server type only)
	if g.config.ProjectType == "http-server" {
		if err := g.generateHTTPServerInterface(projectPath); err != nil {
			return err
		}

		if err := g.generateHTTPServerFile(projectPath); err != nil {
			return err
		}
	}

	// Generate config
	if err := g.generateConfig(projectPath); err != nil {
		return err
	}

	// Generate logger
	if err := g.generateLogger(projectPath); err != nil {
		return err
	}

	// Generate Makefile
	if err := g.generateMakefile(projectPath); err != nil {
		return err
	}

	// Generate .gitignore
	if err := g.generateGitignore(projectPath); err != nil {
		return err
	}

	// Generate README
	if err := g.generateReadme(projectPath); err != nil {
		return err
	}

	// Optional files
	if g.config.WithDocker {
		if err := g.generateDockerFiles(projectPath); err != nil {
			return err
		}
	}

	if g.config.WithObservability {
		if err := g.generateObservability(projectPath); err != nil {
			return err
		}
	}

	return nil
}

// initGoModule initializes the go.mod file
func (g *ProjectGenerator) initGoModule(projectPath string) error {
	fmt.Println("📦 Initializing go module...")

	cmd := exec.Command("go", "mod", "init", g.config.ModuleName)
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}

	// Add required dependencies
	return g.addDependencies(projectPath)
}

// addDependencies adds required dependencies to go.mod
func (g *ProjectGenerator) addDependencies(projectPath string) error {
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

	// Add metrics dependencies
	if g.config.WithMetrics {
		dependencies = append(dependencies, "github.com/prometheus/client_golang@latest")
	}

	for _, dep := range dependencies {
		cmd := exec.Command("go", "get", dep)
		cmd.Dir = projectPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: failed to add dependency %s: %v\n", dep, err)
		}
	}

	return nil
}

// runGoModTidy runs go mod tidy
func (g *ProjectGenerator) runGoModTidy(projectPath string) error {
	fmt.Println("🧹 Running go mod tidy...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

// formatCode runs go fmt on the generated code
func (g *ProjectGenerator) formatCode(projectPath string) error {
	fmt.Println("✨ Formatting code...")

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = projectPath

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// printSuccess prints success message with next steps
func (g *ProjectGenerator) printSuccess(projectPath string) {
	fmt.Println("\n✅ Project generated successfully!")
	fmt.Println("\n📚 Next steps:")
	fmt.Printf("  cd %s\n", g.config.ProjectName)
	fmt.Println("  go run main.go run")
	fmt.Println("\n📖 Read the README.md for more information about the project structure.")
}
