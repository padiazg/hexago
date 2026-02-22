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
	moduleName        string
	projectType       string
	framework         string
	adapterStyle      string
	coreLogic         string
	withDocker        bool
	withExample       bool
	withMigrations    bool
	withMetrics       bool
	explicitPorts     bool
	withWorkers       bool
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

Project Types:
  http-server  - HTTP API server with web framework
  service      - Long-running daemon/service (no web framework for main logic)

Example:
  hexago init my-api --module github.com/user/my-api --project-type http-server --framework echo
  hexago init my-service --module github.com/user/my-service --project-type service`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Required flags
	initCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Go module name (e.g., github.com/user/my-app)")

	// Project type and architecture choices
	initCmd.Flags().StringVarP(&projectType, "project-type", "t", "http-server", "Project type (http-server|service)")
	initCmd.Flags().StringVarP(&framework, "framework", "f", "stdlib", "Web framework for http-server (echo|gin|chi|fiber|stdlib)")
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

	// Load .hexago.yaml from CWD as a defaults layer (flags > yaml > hardcoded defaults)
	if hexCfg, err := generator.LoadHexagoConfig("."); err == nil {
		fmt.Println("ℹ️  Loading defaults from .hexago.yaml")
		pc := hexCfg.ToProjectConfig()
		if !cmd.Flags().Changed("module") && pc.ModuleName != "" {
			moduleName = pc.ModuleName
		}
		if !cmd.Flags().Changed("project-type") && pc.ProjectType != "" {
			projectType = pc.ProjectType
		}
		if !cmd.Flags().Changed("framework") && pc.Framework != "" {
			framework = pc.Framework
		}
		if !cmd.Flags().Changed("adapter-style") && pc.AdapterStyle != "" {
			adapterStyle = pc.AdapterStyle
		}
		if !cmd.Flags().Changed("core-logic") && pc.CoreLogic != "" {
			coreLogic = pc.CoreLogic
		}
		if !cmd.Flags().Changed("with-docker") {
			withDocker = pc.WithDocker
		}
		if !cmd.Flags().Changed("with-example") {
			withExample = pc.WithExample
		}
		if !cmd.Flags().Changed("with-migrations") {
			withMigrations = pc.WithMigrations
		}
		if !cmd.Flags().Changed("with-metrics") {
			withMetrics = pc.WithMetrics
		}
		if !cmd.Flags().Changed("explicit-ports") {
			explicitPorts = pc.ExplicitPorts
		}
		if !cmd.Flags().Changed("with-workers") {
			withWorkers = pc.WithWorkers
		}
		if !cmd.Flags().Changed("with-observability") {
			withObservability = pc.WithObservability
		}
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

	// Validate project type
	if err := validateProjectType(projectType); err != nil {
		return err
	}

	// Validate framework (only required for http-server)
	if projectType == "http-server" {
		if err := validateFramework(framework); err != nil {
			return err
		}
	} else if framework != "stdlib" {
		// Warn if framework specified for non-http-server projects
		fmt.Printf("⚠️  Warning: --framework is ignored for project type '%s' (only used for http-server)\n", projectType)
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
	config.ProjectType = projectType
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

func validateProjectType(pt string) error {
	validTypes := map[string]bool{
		"http-server": true,
		"service":     true,
	}

	if !validTypes[pt] {
		return fmt.Errorf("invalid project type '%s'. Valid options: http-server, service", pt)
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
	fmt.Printf("  Project Type:      %s\n", config.ProjectType)
	if config.IsHTTPServer() {
		fmt.Printf("  Framework:         %s\n", config.Framework)
	}
	fmt.Printf("  Adapter Style:     %s\n", config.AdapterStyle)
	fmt.Printf("  Core Logic:        %s\n", config.CoreLogic)
	fmt.Printf("  Docker:            %v\n", config.WithDocker)
	fmt.Printf("  Observability:     %v\n", config.WithObservability)
	fmt.Printf("  Migrations:        %v\n", config.WithMigrations)
	fmt.Printf("  Workers:           %v\n", config.WithWorkers)
	fmt.Printf("  Example Code:      %v\n", config.WithExample)
	fmt.Println()
}
