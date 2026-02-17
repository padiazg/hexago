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
	migrationType string
)

// addMigrationCmd represents the add migration command
var addMigrationCmd = &cobra.Command{
	Use:   "migration <name>",
	Short: "Add a database migration",
	Long: `Add a database migration file using golang-migrate format.

Generates sequentially numbered up and down migration files:
  - migrations/000001_<name>.up.sql
  - migrations/000001_<name>.down.sql

Migration types:
  sql (default) - SQL migration files
  go            - Go-based migrations (future)

Example:
  hexago add migration create_users_table
  hexago add migration add_email_index
  hexago add migration alter_products_table`,
	Args: cobra.ExactArgs(1),
	RunE: runAddMigration,
}

func init() {
	addCmd.AddCommand(addMigrationCmd)

	addMigrationCmd.Flags().StringVarP(&migrationType, "type", "t", "sql", "Migration type (sql|go)")
}

func runAddMigration(cmd *cobra.Command, args []string) error {
	migrationName := args[0]

	// Validate migration name (should be snake_case)
	if migrationName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	// Validate type
	if migrationType != "sql" && migrationType != "go" {
		return fmt.Errorf("invalid migration type '%s'. Valid types: sql, go", migrationType)
	}

	if migrationType == "go" {
		return fmt.Errorf("go migrations not yet implemented. Use --type sql")
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding migration: %s\n", migrationName)
	fmt.Printf("   Project: %s\n", config.ProjectName)
	fmt.Printf("   Type: %s\n\n", migrationType)

	// Generate migration
	gen := generator.NewMigrationGenerator(config)
	migrationNumber, err := gen.Generate(migrationName)
	if err != nil {
		return fmt.Errorf("failed to generate migration: %w", err)
	}

	fmt.Println("\n✅ Migration added successfully!")
	fmt.Printf("\n📝 Files created:\n")
	fmt.Printf("   - migrations/%06d_%s.up.sql\n", migrationNumber, migrationName)
	fmt.Printf("   - migrations/%06d_%s.down.sql\n", migrationNumber, migrationName)
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Edit the .up.sql file with your schema changes\n")
	fmt.Printf("  2. Edit the .down.sql file to reverse those changes\n")
	fmt.Printf("  3. Run migrations:\n")
	fmt.Printf("     make migrate-up\n")
	fmt.Printf("  4. To rollback:\n")
	fmt.Printf("     make migrate-down\n")

	return nil
}
