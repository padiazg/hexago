package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/padiazg/hexago/pkg/fileutil"
)

// MigrationGenerator generates database migration files
type MigrationGenerator struct {
	config *ProjectConfig
}

// NewMigrationGenerator creates a new migration generator
func NewMigrationGenerator(config *ProjectConfig) *MigrationGenerator {
	return &MigrationGenerator{
		config: config,
	}
}

// Generate creates migration files with sequential numbering
func (g *MigrationGenerator) Generate(migrationName string) (int, error) {
	// Create migrations directory if it doesn't exist
	migrationsDir := "migrations"
	if err := fileutil.CreateDir(migrationsDir); err != nil {
		return 0, err
	}

	// Get next migration number
	migrationNumber, err := g.getNextMigrationNumber(migrationsDir)
	if err != nil {
		return 0, err
	}

	// Generate file names
	upFile := fmt.Sprintf("%06d_%s.up.sql", migrationNumber, migrationName)
	downFile := fmt.Sprintf("%06d_%s.down.sql", migrationNumber, migrationName)

	upPath := filepath.Join(migrationsDir, upFile)
	downPath := filepath.Join(migrationsDir, downFile)

	fmt.Printf("📝 Creating migration files:\n")
	fmt.Printf("   UP:   %s\n", upPath)
	fmt.Printf("   DOWN: %s\n", downPath)

	// Generate UP migration
	if err := g.generateUpMigration(upPath, migrationName); err != nil {
		return 0, err
	}

	// Generate DOWN migration
	if err := g.generateDownMigration(downPath, migrationName); err != nil {
		return 0, err
	}

	// Generate or update migration manager (first time only)
	if err := g.ensureMigrationManager(); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to ensure migration manager: %v\n", err)
	}

	// Update Makefile with migration commands (first time only)
	if err := g.ensureMakefileMigrationCommands(); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠️  Warning: failed to update Makefile: %v\n", err)
	}

	return migrationNumber, nil
}

// getNextMigrationNumber finds the next sequential migration number
func (g *MigrationGenerator) getNextMigrationNumber(migrationsDir string) (int, error) {
	// Pattern to match migration files: NNNNNN_name.up.sql
	pattern := regexp.MustCompile(`^(\d{6})_.*\.up\.sql$`)

	maxNumber := 0

	// Read directory
	entries, err := fileutil.ReadDir(migrationsDir)
	if err != nil {
		// Directory doesn't exist or is empty - start at 1
		return 1, nil
	}

	// Find highest number
	for _, entry := range entries {
		if matches := pattern.FindStringSubmatch(entry); len(matches) > 1 {
			num, err := strconv.Atoi(matches[1])
			if err == nil && num > maxNumber {
				maxNumber = num
			}
		}
	}

	return maxNumber + 1, nil
}

// generateUpMigration creates the UP migration file
func (g *MigrationGenerator) generateUpMigration(filePath, migrationName string) error {
	data := map[string]interface{}{
		"MigrationName": migrationName,
		"Timestamp":     "now", // Could use time.Now() for actual timestamp
	}

	content, err := globalTemplateLoader.Render("migration/up.sql.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render UP migration template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// generateDownMigration creates the DOWN migration file
func (g *MigrationGenerator) generateDownMigration(filePath, migrationName string) error {
	data := map[string]interface{}{
		"MigrationName": migrationName,
		"Timestamp":     "now",
	}

	content, err := globalTemplateLoader.Render("migration/down.sql.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render DOWN migration template: %w", err)
	}

	return fileutil.WriteFile(filePath, content)
}

// ensureMigrationManager creates the migration manager if it doesn't exist
func (g *MigrationGenerator) ensureMigrationManager() error {
	dbDir := filepath.Join("internal", "infrastructure", "database")
	managerPath := filepath.Join(dbDir, "migrator.go")

	// If manager already exists, don't overwrite
	if fileutil.FileExists(managerPath) {
		return nil
	}

	// Create directory
	if err := fileutil.CreateDir(dbDir); err != nil {
		return err
	}

	fmt.Printf("📝 Creating migration manager: %s\n", managerPath)

	data := map[string]interface{}{
		"ModuleName": g.config.ModuleName,
	}

	content, err := globalTemplateLoader.Render("migration/migrator.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render migrator template: %w", err)
	}

	return fileutil.WriteFile(managerPath, content)
}

// ensureMakefileMigrationCommands adds migration commands to Makefile
func (g *MigrationGenerator) ensureMakefileMigrationCommands() error {
	// For now, just inform the user to add manually
	// Full implementation would parse and update Makefile
	fmt.Printf("\nℹ️  Add these commands to your Makefile:\n")
	fmt.Printf(`
migrate-up: ## Run database migrations
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
	@migrate -path migrations -database "$(DB_URL)" down 1

migrate-version: ## Show current migration version
	@migrate -path migrations -database "$(DB_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@migrate -path migrations -database "$(DB_URL)" force $(VERSION)

# Add DB_URL to your environment or Makefile:
# DB_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
`)

	return nil
}
