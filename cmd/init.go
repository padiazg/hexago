/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/spf13/cobra"
)

var (
	moduleName       string
	framework        string
	adapterStyle     string
	coreLogic        string
	withDocker       bool
	withExample      bool
	withMigrations   bool
	withMetrics      bool
	explicitPorts    bool
	withWorkers      bool
	withObservability bool
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

Example:
  hexago init my-app --module github.com/user/my-app
  hexago init blog-api --module github.com/user/blog-api --framework echo
  hexago init iot-service --adapter-style driver-driven --core-logic usecases`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Required flags
	initCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Go module name (e.g., github.com/user/my-app)")

	// Framework and architecture choices
	initCmd.Flags().StringVarP(&framework, "framework", "f", "stdlib", "Web framework (echo|gin|chi|fiber|stdlib)")
	initCmd.Flags().StringVar(&adapterStyle, "adapter-style", "primary-secondary", "Adapter naming style (primary-secondary|driver-driven)")
	initCmd.Flags().StringVar(&coreLogic, "core-logic", "services", "Core business logic directory name (services|usecases)")

	// Optional features - all default to false for maximum flexibility
	initCmd.Flags().BoolVar(&withDocker, "with-docker", false, "Generate Docker files")
	initCmd.Flags().BoolVar(&withExample, "with-example", false, "Include example code")
	initCmd.Flags().BoolVar(&withMigrations, "with-migrations", false, "Include database migration setup")
	initCmd.Flags().BoolVar(&withMetrics, "with-metrics", false, "Include Prometheus metrics")
	initCmd.Flags().BoolVar(&explicitPorts, "explicit-ports", false, "Create explicit ports/ directory")
	initCmd.Flags().BoolVar(&withWorkers, "with-workers", false, "Include worker pattern setup")
	initCmd.Flags().BoolVar(&withObservability, "with-observability", false, "Include observability (health checks + metrics)")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Generate module name if not provided
	if moduleName == "" {
		moduleName = projectName
		fmt.Printf("ℹ️  No module name provided, using: %s\n", moduleName)
	}

	// Validate module name
	if err := validateModuleName(moduleName); err != nil {
		return err
	}

	// Validate framework
	if err := validateFramework(framework); err != nil {
		return err
	}

	// Validate adapter style
	if err := validateAdapterStyle(adapterStyle); err != nil {
		return err
	}

	// Validate core logic name
	if err := validateCoreLogic(coreLogic); err != nil {
		return err
	}

	// Create project configuration
	config := generator.NewProjectConfig(projectName, moduleName)
	config.Framework = framework
	config.AdapterStyle = adapterStyle
	config.CoreLogic = coreLogic
	config.WithDocker = withDocker
	config.WithExample = withExample
	config.WithMigrations = withMigrations
	config.WithMetrics = withMetrics
	config.ExplicitPorts = explicitPorts
	config.WithWorkers = withWorkers
	config.WithObservability = withObservability

	// Print configuration
	printProjectInfo(config)

	// Generate project
	gen := generator.NewProjectGenerator(config)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	return nil
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

func validateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Basic validation - could be more strict
	if !strings.Contains(name, "/") && !strings.Contains(name, ".") {
		fmt.Printf("⚠️  Warning: module name '%s' doesn't follow Go module naming convention (domain.com/user/project)\n", name)
	}

	return nil
}

func validateFramework(fw string) error {
	validFrameworks := map[string]bool{
		"echo":   true,
		"gin":    true,
		"chi":    true,
		"fiber":  true,
		"stdlib": true,
	}

	if !validFrameworks[fw] {
		return fmt.Errorf("invalid framework '%s'. Valid options: echo, gin, chi, fiber, stdlib", fw)
	}

	return nil
}

func validateAdapterStyle(style string) error {
	validStyles := map[string]bool{
		"primary-secondary": true,
		"driver-driven":     true,
	}

	if !validStyles[style] {
		return fmt.Errorf("invalid adapter style '%s'. Valid options: primary-secondary, driver-driven", style)
	}

	return nil
}

func validateCoreLogic(name string) error {
	validNames := map[string]bool{
		"services": true,
		"usecases": true,
	}

	if !validNames[name] {
		return fmt.Errorf("invalid core logic name '%s'. Valid options: services, usecases", name)
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

func printProjectInfo(config *generator.ProjectConfig) {
	fmt.Println("\n📋 Project Configuration:")
	fmt.Printf("  Name:              %s\n", config.ProjectName)
	fmt.Printf("  Module:            %s\n", config.ModuleName)
	fmt.Printf("  Framework:         %s\n", config.Framework)
	fmt.Printf("  Adapter Style:     %s\n", config.AdapterStyle)
	fmt.Printf("  Core Logic:        %s\n", config.CoreLogic)
	fmt.Printf("  Docker:            %v\n", config.WithDocker)
	fmt.Printf("  Observability:     %v\n", config.WithObservability)
	fmt.Printf("  Migrations:        %v\n", config.WithMigrations)
	fmt.Printf("  Workers:           %v\n", config.WithWorkers)
	fmt.Printf("  Example Code:      %v\n", config.WithExample)
	fmt.Println()
}
