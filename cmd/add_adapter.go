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
	adapterPort string
)

// addAdapterCmd represents the add adapter command
var addAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Add adapters (primary/secondary or driver/driven)",
	Long: `Add adapter implementations for external interfaces.

Adapters are divided into:
  - Primary/Driver: Inbound adapters (HTTP handlers, gRPC, CLI)
  - Secondary/Driven: Outbound adapters (repositories, external services)

Example:
  hexago add adapter primary http UserHandler
  hexago add adapter secondary database UserRepository`,
}

// addAdapterPrimaryCmd adds primary (inbound) adapters
var addAdapterPrimaryCmd = &cobra.Command{
	Use:   "primary <type> <name>",
	Short: "Add a primary (inbound) adapter",
	Long: `Add a primary/driver adapter that receives requests from external sources.

Types:
  http   - HTTP handler
  grpc   - gRPC handler
  queue  - Message queue consumer

Example:
  hexago add adapter primary http UserHandler
  hexago add adapter primary grpc OrderService`,
	Args: cobra.ExactArgs(2),
	RunE: runAddAdapterPrimary,
}

// addAdapterSecondaryCmd adds secondary (outbound) adapters
var addAdapterSecondaryCmd = &cobra.Command{
	Use:   "secondary <type> <name>",
	Short: "Add a secondary (outbound) adapter",
	Long: `Add a secondary/driven adapter for outbound communication.

Types:
  database  - Database repository
  external  - External service client
  cache     - Cache adapter

Example:
  hexago add adapter secondary database UserRepository
  hexago add adapter secondary external EmailService`,
	Args: cobra.ExactArgs(2),
	RunE: runAddAdapterSecondary,
}

func init() {
	addCmd.AddCommand(addAdapterCmd)
	addAdapterCmd.AddCommand(addAdapterPrimaryCmd)
	addAdapterCmd.AddCommand(addAdapterSecondaryCmd)

	// Flags
	addAdapterPrimaryCmd.Flags().StringVarP(&adapterPort, "port", "p", "", "Port interface name (if using explicit ports)")
	addAdapterSecondaryCmd.Flags().StringVarP(&adapterPort, "port", "p", "", "Port interface name (if using explicit ports)")
}

func runAddAdapterPrimary(cmd *cobra.Command, args []string) error {
	adapterType := args[0]
	adapterName := args[1]

	if err := validateComponentName(adapterName); err != nil {
		return err
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding primary adapter: %s (%s)\n", adapterName, adapterType)
	fmt.Printf("   Project: %s\n", config.ProjectName)
	fmt.Printf("   Adapter dir: %s\n\n", config.AdapterInboundDir())

	gen := generator.NewAdapterGenerator(config)
	if err := gen.GeneratePrimary(adapterType, adapterName, adapterPort); err != nil {
		return fmt.Errorf("failed to generate adapter: %w", err)
	}

	fmt.Println("\n✅ Primary adapter added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Implement the adapter methods\n")
	fmt.Printf("  2. Wire up dependencies in the DI container\n")
	fmt.Printf("  3. Add routes/endpoints as needed\n")

	return nil
}

func runAddAdapterSecondary(cmd *cobra.Command, args []string) error {
	adapterType := args[0]
	adapterName := args[1]

	if err := validateComponentName(adapterName); err != nil {
		return err
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding secondary adapter: %s (%s)\n", adapterName, adapterType)
	fmt.Printf("   Project: %s\n", config.ProjectName)
	fmt.Printf("   Adapter dir: %s\n\n", config.AdapterOutboundDir())

	gen := generator.NewAdapterGenerator(config)
	if err := gen.GenerateSecondary(adapterType, adapterName, adapterPort); err != nil {
		return fmt.Errorf("failed to generate adapter: %w", err)
	}

	fmt.Println("\n✅ Secondary adapter added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Implement the port interface methods\n")
	fmt.Printf("  2. Add database queries or external API calls\n")
	fmt.Printf("  3. Wire up dependencies in the DI container\n")

	return nil
}
