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
	entityFields string
)

// addDomainCmd represents the add domain command
var addDomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Add domain entities or value objects",
	Long: `Add domain layer components (entities or value objects) to internal/core/domain.

Available subcommands:
  entity        - Add a domain entity
  valueobject   - Add a value object

Example:
  hexago add domain entity User --fields "id:string,name:string,email:string"
  hexago add domain valueobject Email`,
}

// addDomainEntityCmd represents adding a domain entity
var addDomainEntityCmd = &cobra.Command{
	Use:   "entity <name>",
	Short: "Add a new domain entity",
	Long: `Add a new domain entity to internal/core/domain.

Entities are objects with unique identity that persist through time.
They contain business logic and validation rules.

Example:
  hexago add domain entity User --fields "id:string,name:string,email:string,createdAt:time.Time"
  hexago add domain entity Order
  hexago add domain entity Product --fields "id:string,name:string,price:float64"`,
	Args: cobra.ExactArgs(1),
	RunE: runAddDomainEntity,
}

// addDomainValueObjectCmd represents adding a value object
var addDomainValueObjectCmd = &cobra.Command{
	Use:   "valueobject <name>",
	Short: "Add a new value object",
	Long: `Add a new value object to internal/core/domain.

Value objects are immutable objects defined by their attributes.
They don't have unique identity and are compared by value.

Example:
  hexago add domain valueobject Email
  hexago add domain valueobject Address --fields "street:string,city:string,zipCode:string"
  hexago add domain valueobject Money --fields "amount:float64,currency:string"`,
	Args: cobra.ExactArgs(1),
	RunE: runAddDomainValueObject,
}

func init() {
	addCmd.AddCommand(addDomainCmd)
	addDomainCmd.AddCommand(addDomainEntityCmd)
	addDomainCmd.AddCommand(addDomainValueObjectCmd)

	// Flags for entity
	addDomainEntityCmd.Flags().StringVarP(&entityFields, "fields", "f", "", "Comma-separated field definitions (name:type)")

	// Flags for value object
	addDomainValueObjectCmd.Flags().StringVarP(&entityFields, "fields", "f", "", "Comma-separated field definitions (name:type)")
}

func runAddDomainEntity(cmd *cobra.Command, args []string) error {
	entityName := args[0]

	if err := validateComponentName(entityName); err != nil {
		return err
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding domain entity: %s\n", entityName)
	fmt.Printf("   Project: %s\n\n", config.ProjectName)

	// Parse fields
	fields, err := parseFields(entityFields)
	if err != nil {
		return fmt.Errorf("failed to parse fields: %w", err)
	}

	// Generate entity
	gen := generator.NewDomainGenerator(config)
	if err := gen.GenerateEntity(entityName, fields); err != nil {
		return fmt.Errorf("failed to generate entity: %w", err)
	}

	fmt.Println("\n✅ Domain entity added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Add business logic methods to the entity\n")
	fmt.Printf("  2. Add validation rules\n")
	fmt.Printf("  3. Write tests for domain logic\n")

	return nil
}

func runAddDomainValueObject(cmd *cobra.Command, args []string) error {
	voName := args[0]

	if err := validateComponentName(voName); err != nil {
		return err
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding value object: %s\n", voName)
	fmt.Printf("   Project: %s\n\n", config.ProjectName)

	// Parse fields
	fields, err := parseFields(entityFields)
	if err != nil {
		return fmt.Errorf("failed to parse fields: %w", err)
	}

	// Generate value object
	gen := generator.NewDomainGenerator(config)
	if err := gen.GenerateValueObject(voName, fields); err != nil {
		return fmt.Errorf("failed to generate value object: %w", err)
	}

	fmt.Println("\n✅ Value object added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Ensure immutability (no setter methods)\n")
	fmt.Printf("  2. Implement validation in constructor\n")
	fmt.Printf("  3. Implement Equals method for value comparison\n")

	return nil
}

// parseFields parses field definitions from string
// Format: "name:type,name:type"
func parseFields(fieldsStr string) ([]generator.Field, error) {
	if fieldsStr == "" {
		return []generator.Field{}, nil
	}

	var fields []generator.Field
	parts := strings.Split(fieldsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fieldParts := strings.Split(part, ":")
		if len(fieldParts) != 2 {
			return nil, fmt.Errorf("invalid field format '%s'. Expected 'name:type'", part)
		}

		name := strings.TrimSpace(fieldParts[0])
		typeName := strings.TrimSpace(fieldParts[1])

		if name == "" || typeName == "" {
			return nil, fmt.Errorf("field name and type cannot be empty in '%s'", part)
		}

		// Capitalize first letter for Go convention
		if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
			name = strings.ToUpper(name[:1]) + name[1:]
		}

		fields = append(fields, generator.Field{
			Name: name,
			Type: typeName,
		})
	}

	return fields, nil
}
