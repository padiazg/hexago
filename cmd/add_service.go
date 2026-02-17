/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/spf13/cobra"
)

var (
	serviceDescription string
)

// addServiceCmd represents the add service command
var addServiceCmd = &cobra.Command{
	Use:   "service <name>",
	Short: "Add a new service/usecase to the project",
	Long: `Add a new service (or usecase) to the internal/core directory.

This generates a service file with:
  - Input/Output structs
  - Service struct with dependencies
  - Constructor function
  - Execute method
  - Test file with basic structure

The service will be placed in internal/core/services/ (or usecases/
depending on your project configuration).

Example:
  hexago add service CreateUser
  hexago add service GetUserByID --description "Retrieves a user by ID"
  hexago add service SendEmail`,
	Args: cobra.ExactArgs(1),
	RunE: runAddService,
}

func init() {
	addCmd.AddCommand(addServiceCmd)

	addServiceCmd.Flags().StringVarP(&serviceDescription, "description", "d", "", "Service description")
}

func runAddService(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Validate service name
	if err := validateComponentName(serviceName); err != nil {
		return err
	}

	// Detect current project configuration
	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w\nMake sure you're in a hexagonal architecture project directory", err)
	}

	fmt.Printf("📦 Adding service: %s\n", serviceName)
	fmt.Printf("   Project: %s\n", config.ProjectName)
	fmt.Printf("   Module: %s\n", config.ModuleName)
	fmt.Printf("   Logic dir: %s\n\n", config.CoreLogic)

	// Generate service
	gen := generator.NewServiceGenerator(config)
	if err := gen.Generate(serviceName, serviceDescription); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	fmt.Println("\n✅ Service added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Implement the business logic in the Execute method\n")
	fmt.Printf("  2. Add any required dependencies to the constructor\n")
	fmt.Printf("  3. Write tests in the generated test file\n")

	return nil
}

func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	if strings.ContainsAny(name, " /\\:*?\"<>|.") {
		return fmt.Errorf("component name contains invalid characters")
	}

	// Should start with uppercase letter for Go conventions
	if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
		return fmt.Errorf("component name should start with uppercase letter (Go convention)")
	}

	return nil
}
