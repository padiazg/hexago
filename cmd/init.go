/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Initialize a new hexagonal architecture project",
	Long: `Initialize a new Go project with hexagonal architecture structure.

This command creates a complete project structure following the Hexagonal
Architecture (Ports & Adapters) pattern, including:

  - Cobra CLI structure with graceful shutdown
  - Hexagonal architecture directories (core, adapters, config)
  - Configuration management with Viper
  - Logger implementation
  - Docker files (optional)
  - Makefile with common tasks
  - README with architecture documentation

Project Types:
  http-server  - HTTP API server with web framework
  service      - Long-running daemon/service (no web framework for main logic)
  cli          - Batch CLI with subcommands (no run command, no HTTP layer)

Example:
  hexago init my-api --module github.com/user/my-api --project-type http-server --framework echo
  hexago init my-service --module github.com/user/my-service --project-type service
  hexago init my-tool --module github.com/user/my-tool --project-type cli`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Required flags
	initCmd.Flags().StringP("module", "m", "", "Go module name (e.g., github.com/user/my-app)")

	// Project type and architecture choices
	initCmd.Flags().StringP("project-type", "t", "http-server", "Project type (http-server|service|cli)")
	initCmd.Flags().StringP("framework", "f", "stdlib", "Web framework for http-server (echo|gin|chi|fiber|stdlib)")
	initCmd.Flags().String("db-driver", "postgres", "Database driver for the migration scaffold (postgres|sqlite3)")
	initCmd.Flags().String("adapter-style", "primary-secondary", "Adapter naming style (primary-secondary|driver-driven)")
	initCmd.Flags().String("core-logic", "services", "Core business logic directory name (services|usecases)")

	// Optional features - all default to false for maximum flexibility
	initCmd.Flags().Bool("with-docker", false, "Generate Docker files")
	initCmd.Flags().Bool("with-example", false, "Include example code")
	initCmd.Flags().Bool("with-migrations", false, "Include database migration setup")
	initCmd.Flags().Bool("with-metrics", false, "Include Prometheus metrics")
	initCmd.Flags().Bool("explicit-ports", false, "Create explicit ports/ directory")
	initCmd.Flags().Bool("with-workers", false, "Include worker pattern setup")
	initCmd.Flags().Bool("with-observability", false, "Include observability (health checks + metrics)")
	initCmd.Flags().Bool("with-tests", false, "Enable go-testgen test generation for add commands (requires go-testgen ≥ v0.1.0)")
	initCmd.Flags().Bool("in-place", false, "Generate project files directly in the working directory (no <name> subdirectory)")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Resolve output directory (working dir flag or CWD)
	outDir := workingDir
	if outDir == "" {
		var wdErr error
		outDir, wdErr = os.Getwd()
		if wdErr != nil {
			return fmt.Errorf("failed to get current directory: %w", wdErr)
		}
	}

	// Load config via viper: flag > .hexago.yaml > defaults
	v := viper.New()
	config, err := generator.LoadInitConfig(v, outDir, cmd)
	if err != nil {
		return err
	}

	config.Project.Name = projectName
	config.OutputDir = outDir
	config.InPlace, _ = cmd.Flags().GetBool("in-place")

	// Default module name if still empty
	if config.Project.Module == "" {
		config.Project.Module = projectName
		fmt.Printf("ℹ️  No module name provided, using: %s\n", config.Project.Module)
	}

	if err := config.Validate(); err != nil {
		return err
	}

	// Validate framework (only required for http-server)
	if config.Project.Type != "http-server" && config.Project.Framework != "stdlib" {
		fmt.Printf("⚠️  Warning: --framework is ignored for project type '%s' (only used for http-server)\n", config.Project.Type)
	}

	// Print configuration
	printProjectInfo(config)

	// Generate project
	gen := generator.NewProjectGenerator(*config)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	return nil
}

func printProjectInfo(config *generator.HexagoConfig) {
	fmt.Println("\n📋 Project Configuration:")
	fmt.Printf("  Name:              %s\n", config.Project.Name)
	fmt.Printf("  Module:            %s\n", config.Project.Module)
	fmt.Printf("  Project Type:      %s\n", config.Project.Type)
	if config.IsHTTPServer() {
		fmt.Printf("  Framework:         %s\n", config.Project.Framework)
	}
	fmt.Printf("  Adapter Style:     %s\n", config.Structure.AdapterStyle)
	fmt.Printf("  Core Logic:        %s\n", config.Structure.CoreLogic)
	fmt.Printf("  Docker:            %v\n", config.Features.WithDocker)
	fmt.Printf("  Observability:     %v\n", config.Features.WithObservability)
	fmt.Printf("  Migrations:        %v\n", config.Features.WithMigrations)
	fmt.Printf("  Workers:           %v\n", config.Features.WithWorkers)
	fmt.Printf("  Example Code:      %v\n", config.Features.WithExample)
	fmt.Println()
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, " /\\:*?\"<>|") {
		return fmt.Errorf("project name contains invalid characters")
	}

	// Check if directory already exists
	if err := validateDirectoryNotExists(name); err != nil {
		return err
	}

	return nil
}

func validateDirectoryNotExists(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check with fileutil
	// For now, we'll let the generator handle this check
	_ = absPath

	return nil
}
