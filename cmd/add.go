/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add components to an existing hexagonal architecture project",
	Long: `Add various components to an existing project following hexagonal architecture.

Available subcommands:
  service    - Add a business logic service/usecase
  domain     - Add domain entities or value objects
  adapter    - Add adapters (primary/secondary or driver/driven)
  worker     - Add a background worker
  migration  - Add a database migration

Example:
  hexago add service CreateUser
  hexago add domain entity User
  hexago add adapter primary http UserHandler
  hexago add adapter secondary database UserRepository
  hexago add worker EmailWorker
  hexago add migration create_users_table`,
}

func init() {
	rootCmd.AddCommand(addCmd)
}
