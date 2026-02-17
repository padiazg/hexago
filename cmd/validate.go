/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/spf13/cobra"
)

var (
	validateFix bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate hexagonal architecture compliance",
	Long: `Validate that the project follows hexagonal architecture principles.

Checks performed:
  ✓ Core domain has no external dependencies
  ✓ Services/UseCases only depend on domain and ports
  ✓ Adapters don't import from other adapters
  ✓ Proper package organization
  ✓ Naming conventions
  ✓ Dependency direction (inward only)

Example:
  hexago validate
  hexago validate --fix  # Attempt to fix issues (future)`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "Attempt to fix issues automatically (not yet implemented)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	if validateFix {
		return fmt.Errorf("--fix flag not yet implemented")
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w\nMake sure you're in a hexagonal architecture project directory", err)
	}

	fmt.Printf("🔍 Validating project: %s\n", config.ProjectName)
	fmt.Printf("   Module: %s\n", config.ModuleName)
	fmt.Printf("   Adapter style: %s\n", config.AdapterStyle)
	fmt.Printf("   Core logic: %s\n\n", config.CoreLogic)

	// Run validation
	validator := generator.NewValidator(config)
	result := validator.Validate()

	// Print results
	printValidationResult(result)

	// Exit with error if validation failed
	if result.HasErrors() {
		return fmt.Errorf("validation failed with %d error(s)", result.ErrorCount())
	}

	return nil
}

func printValidationResult(result *generator.ValidationResult) {
	fmt.Println("📋 Validation Results:")

	// Print successes
	for _, check := range result.Successes {
		fmt.Printf("✓ %s\n", check)
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Println()
		for _, warning := range result.Warnings {
			fmt.Printf("⚠️  %s\n", warning)
		}
	}

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Println()
		for _, err := range result.Errors {
			fmt.Printf("✗ %s\n", err)
		}
	}

	// Summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   ✓ Passed: %d\n", len(result.Successes))
	fmt.Printf("   ⚠️  Warnings: %d\n", len(result.Warnings))
	fmt.Printf("   ✗ Errors: %d\n", len(result.Errors))

	if result.HasErrors() {
		fmt.Printf("\n❌ Validation FAILED\n")
	} else if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Validation passed with warnings\n")
	} else {
		fmt.Printf("\n✅ Validation PASSED\n")
	}
}
