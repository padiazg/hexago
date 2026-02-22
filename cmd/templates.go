/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/fileutil"
	"github.com/spf13/cobra"
)

// templatesCmd represents the templates parent command
var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage HexaGo code generation templates",
	Long: `Inspect and customize the templates used by HexaGo to generate code.

Templates are loaded from multiple sources in priority order:
  1. Binary-local   - templates/ directory next to the hexago binary
  2. Project-local  - .hexago/templates/ in the current project
  3. User-global    - ~/.hexago/templates/ in your home directory
  4. Embedded       - built-in templates compiled into the binary

Use subcommands to list, inspect, export, validate, or reset templates.`,
}

// templatesListCmd lists all templates and marks overrides
var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available templates",
	Long:  `List all templates built into HexaGo, marking any that have local or global overrides active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		loader := generator.NewTemplateLoader()

		names, err := loader.List()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}
		sort.Strings(names)

		// Group by first path component
		groups := make(map[string][]string)
		var groupOrder []string
		for _, name := range names {
			parts := strings.SplitN(name, "/", 2)
			group := parts[0]
			if _, ok := groups[group]; !ok {
				groupOrder = append(groupOrder, group)
			}
			groups[group] = append(groups[group], name)
		}
		sort.Strings(groupOrder)

		fmt.Printf("Available templates (%d total):\n\n", len(names))
		for _, group := range groupOrder {
			fmt.Printf("  %s/\n", group)
			for _, name := range groups[group] {
				source, err := loader.Which(name)
				if err != nil {
					continue
				}
				// Detect override: anything that isn't the embedded source
				if strings.HasPrefix(source, "embedded") {
					fmt.Printf("    %s\n", filepath.Base(name))
				} else {
					// Extract source label (first word before space)
					label := strings.SplitN(source, " ", 2)[0]
					fmt.Printf("    %-44s <- %s\n", filepath.Base(name), label)
				}
			}
		}

		fmt.Println()
		fmt.Println("Use 'hexago templates which <name>' for the full override path.")
		fmt.Println("Use 'hexago templates export <name>' to start customizing a template.")
		return nil
	},
}

// templatesWhichCmd shows which source wins for a given template
var templatesWhichCmd = &cobra.Command{
	Use:   "which <name>",
	Short: "Show which source provides a template",
	Long:  `Show the winning source (embedded, project-local, user-global, or binary-local) for a given template name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		loader := generator.NewTemplateLoader()

		source, err := loader.Which(name)
		if err != nil {
			return err
		}

		fmt.Printf("%s -> %s\n", name, source)
		return nil
	},
}

// templatesExportCmd copies a template to an override location
var templatesExportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export a template to a local or global override location",
	Long: `Copy a built-in template to your project-local (.hexago/templates/) or
user-global (~/.hexago/templates/) directory for customization.

Once exported, HexaGo will use your customized version instead of the built-in one.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		global, _ := cmd.Flags().GetBool("global")
		loader := generator.NewTemplateLoader()

		if !loader.Exists(name) {
			return fmt.Errorf("template not found: %s", name)
		}

		if err := loader.Export(name, global); err != nil {
			return err
		}

		var destPath string
		if global {
			destPath = filepath.Join(fileutil.HomeDir(), ".hexago", "templates", name)
		} else {
			destPath = filepath.Join(".hexago", "templates", name)
		}

		fmt.Printf("Template exported to: %s\n", destPath)
		fmt.Printf("Edit it and re-run your hexago commands to use the customized version.\n")
		return nil
	},
}

// templatesExportAllCmd exports all templates at once
var templatesExportAllCmd = &cobra.Command{
	Use:   "export-all",
	Short: "Export all templates to a local or global override location",
	Long: `Copy all built-in templates to your project-local (.hexago/templates/) or
user-global (~/.hexago/templates/) directory for customization.

Templates that already have an override are skipped unless --force is provided.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		global, _ := cmd.Flags().GetBool("global")
		force, _ := cmd.Flags().GetBool("force")
		loader := generator.NewTemplateLoader()

		names, err := loader.List()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}
		sort.Strings(names)

		var baseDir string
		if global {
			baseDir = filepath.Join(fileutil.HomeDir(), ".hexago", "templates")
		} else {
			baseDir = filepath.Join(".hexago", "templates")
		}

		var exported, skipped int
		for _, name := range names {
			destPath := filepath.Join(baseDir, name)
			if !force && fileutil.FileExists(destPath) {
				skipped++
				continue
			}
			if err := loader.Export(name, global); err != nil {
				fmt.Printf("  ✗ %s: %v\n", name, err)
				continue
			}
			fmt.Printf("  ✓ %s\n", name)
			exported++
		}

		fmt.Printf("\nExported %d template(s) to %s", exported, baseDir)
		if skipped > 0 {
			fmt.Printf(" (%d skipped — already exist, use --force to overwrite)", skipped)
		}
		fmt.Println()
		return nil
	},
}

// templatesValidateCmd checks template syntax
var templatesValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a template file for syntax errors",
	Long:  `Parse a template file and report any syntax errors. Useful after editing an exported template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		loader := generator.NewTemplateLoader()

		if err := loader.Validate(path); err != nil {
			fmt.Printf("✗ %s\n  %v\n", path, err)
			return err
		}

		fmt.Printf("✓ %s — template syntax is valid\n", path)
		return nil
	},
}

// templatesResetCmd removes a custom template override
var templatesResetCmd = &cobra.Command{
	Use:   "reset <name>",
	Short: "Remove a custom template override",
	Long: `Delete a template override from your project-local (.hexago/templates/) or
user-global (~/.hexago/templates/) directory. HexaGo will revert to using the built-in template.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		global, _ := cmd.Flags().GetBool("global")
		loader := generator.NewTemplateLoader()

		if err := loader.Reset(name, global); err != nil {
			return err
		}

		scope := "project-local"
		if global {
			scope = "user-global"
		}
		fmt.Printf("Removed %s override for: %s\n", scope, name)
		return nil
	},
}

func init() {
	// Register parent with root
	rootCmd.AddCommand(templatesCmd)

	// Register subcommands
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesWhichCmd)
	templatesCmd.AddCommand(templatesExportCmd)
	templatesCmd.AddCommand(templatesExportAllCmd)
	templatesCmd.AddCommand(templatesValidateCmd)
	templatesCmd.AddCommand(templatesResetCmd)

	// Flags — declared per-subcommand to avoid shared variable races
	templatesExportCmd.Flags().Bool("global", false, "Export to user-global override directory (~/.hexago/templates/)")
	templatesExportAllCmd.Flags().Bool("global", false, "Export to user-global override directory (~/.hexago/templates/)")
	templatesExportAllCmd.Flags().Bool("force", false, "Overwrite templates that already have an override")
	templatesResetCmd.Flags().Bool("global", false, "Remove from user-global override directory (~/.hexago/templates/)")
}
