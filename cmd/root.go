/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hexago",
	Short: "HexaGo - Hexagonal Architecture Scaffolding CLI",
	Long: `HexaGo is an opinionated CLI tool to scaffold applications
following the Hexagonal Architecture (Ports & Adapters) pattern.

It helps developers maintain proper separation of concerns and avoid conceptual
confusion when building Go applications.

Features:
  - Framework support: Echo, Gin, Chi, Fiber, and stdlib
  - Docker ready with multi-stage builds
  - Graceful shutdown with context-based cancellation
  - Background workers using goroutines and channels
  - Database migrations with golang-migrate
  - OpenAPI/Swagger documentation
  - Opinionated structure enforcing hexagonal architecture

Example:
  hexago init my-app --module github.com/user/my-app --framework echo
  hexago add service CreateUser
  hexago add adapter primary http UserHandler`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
