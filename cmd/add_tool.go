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
	toolDescription string
)

// addToolCmd represents the add tool command
var addToolCmd = &cobra.Command{
	Use:   "tool <type> <name>",
	Short: "Add infrastructure tools and utilities",
	Long: `Add infrastructure tools and utilities to the project.

Tool types:
  logger     - Custom logger implementation
  validator  - Input validation utilities
  mapper     - DTO mapping utilities
  middleware - HTTP middleware (auth, logging, rate limiting, etc.)

Examples:
  hexago add tool logger StructuredLogger
  hexago add tool validator RequestValidator
  hexago add tool mapper UserMapper
  hexago add tool middleware AuthMiddleware`,
	Args: cobra.ExactArgs(2),
	RunE: runAddTool,
}

func init() {
	addCmd.AddCommand(addToolCmd)

	addToolCmd.Flags().StringVarP(&toolDescription, "description", "d", "", "Tool description")
}

func runAddTool(cmd *cobra.Command, args []string) error {
	toolType := args[0]
	toolName := args[1]

	// Validate tool type
	validTypes := []string{"logger", "validator", "mapper", "middleware"}
	if !contains(validTypes, toolType) {
		return fmt.Errorf("invalid tool type '%s'. Valid types: %v", toolType, validTypes)
	}

	// Validate tool name
	if err := validateComponentName(toolName); err != nil {
		return err
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w\nMake sure you're in a hexagonal architecture project directory", err)
	}

	fmt.Printf("📦 Adding %s tool: %s\n", toolType, toolName)
	fmt.Printf("   Project: %s\n", config.ProjectName)
	if toolDescription != "" {
		fmt.Printf("   Description: %s\n", toolDescription)
	}
	fmt.Println()

	// Generate tool
	gen := generator.NewToolGenerator(config)
	if err := gen.Generate(toolType, toolName, toolDescription); err != nil {
		return fmt.Errorf("failed to generate tool: %w", err)
	}

	fmt.Println("\n✅ Tool added successfully!")
	fmt.Printf("\n📝 Files created:\n")
	fmt.Printf("   - internal/infrastructure/%s/%s\n", toolType, toSnakeCase(toolName)+".go")
	fmt.Printf("   - internal/infrastructure/%s/%s\n", toolType, toSnakeCase(toolName)+"_test.go")

	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Implement the %s logic\n", toolName)
	fmt.Printf("  2. Write unit tests\n")
	fmt.Printf("  3. Use the tool in your services or adapters\n")

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
