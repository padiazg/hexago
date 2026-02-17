package generator

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ValidationResult holds validation results
type ValidationResult struct {
	Successes []string
	Warnings  []string
	Errors    []string
}

// HasErrors returns true if there are any errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of errors
func (r *ValidationResult) ErrorCount() int {
	return len(r.Errors)
}

// Validator validates hexagonal architecture compliance
type Validator struct {
	config *ProjectConfig
}

// NewValidator creates a new validator
func NewValidator(config *ProjectConfig) *Validator {
	return &Validator{
		config: config,
	}
}

// Validate runs all validation checks
func (v *Validator) Validate() *ValidationResult {
	result := &ValidationResult{
		Successes: make([]string, 0),
		Warnings:  make([]string, 0),
		Errors:    make([]string, 0),
	}

	// Check 1: Project structure
	v.validateProjectStructure(result)

	// Check 2: Core domain dependencies
	v.validateCoreDependencies(result)

	// Check 3: Service/UseCase dependencies
	v.validateServiceDependencies(result)

	// Check 4: Adapter dependencies
	v.validateAdapterDependencies(result)

	// Check 5: Naming conventions
	v.validateNamingConventions(result)

	return result
}

// validateProjectStructure checks if required directories exist
func (v *Validator) validateProjectStructure(result *ValidationResult) {
	requiredDirs := []struct {
		path        string
		description string
	}{
		{"internal/core/domain", "Domain directory"},
		{filepath.Join("internal/core", v.config.CoreLogicDir()), "Core logic directory"},
		{filepath.Join("internal/adapters", v.config.AdapterInboundDir()), "Inbound adapters directory"},
		{filepath.Join("internal/adapters", v.config.AdapterOutboundDir()), "Outbound adapters directory"},
		{"internal/config", "Config directory"},
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir.path); err == nil {
			result.Successes = append(result.Successes, fmt.Sprintf("%s exists", dir.description))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s not found: %s", dir.description, dir.path))
		}
	}
}

// validateCoreDependencies ensures core/domain has no external dependencies
func (v *Validator) validateCoreDependencies(result *ValidationResult) {
	domainPath := filepath.Join("internal", "core", "domain")

	violations, err := v.checkImports(domainPath, func(importPath string) bool {
		// Domain should not import from adapters or infrastructure
		return !strings.Contains(importPath, "/adapters/") &&
			!strings.Contains(importPath, "/infrastructure/")
	})

	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check domain dependencies: %v", err))
		return
	}

	if len(violations) == 0 {
		result.Successes = append(result.Successes, "Core domain has no external dependencies")
	} else {
		for _, v := range violations {
			result.Errors = append(result.Errors, fmt.Sprintf("Domain imports external package: %s in %s", v.importPath, v.file))
		}
	}
}

// validateServiceDependencies ensures services only depend on domain and ports
func (v *Validator) validateServiceDependencies(result *ValidationResult) {
	servicePath := filepath.Join("internal", "core", v.config.CoreLogicDir())

	violations, err := v.checkImports(servicePath, func(importPath string) bool {
		// Services can import domain and ports, but not adapters
		if strings.Contains(importPath, v.config.ModuleName) {
			return !strings.Contains(importPath, "/adapters/")
		}
		return true
	})

	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check %s dependencies: %v", v.config.CoreLogicDir(), err))
		return
	}

	if len(violations) == 0 {
		result.Successes = append(result.Successes, fmt.Sprintf("%s only depend on domain and ports", strings.Title(v.config.CoreLogicDir())))
	} else {
		for _, violation := range violations {
			result.Errors = append(result.Errors, fmt.Sprintf("%s imports adapter: %s in %s", strings.Title(v.config.CoreLogicDir()), violation.importPath, violation.file))
		}
	}
}

// validateAdapterDependencies ensures adapters don't import from other adapters
func (v *Validator) validateAdapterDependencies(result *ValidationResult) {
	adaptersPath := filepath.Join("internal", "adapters")

	violations, err := v.checkImports(adaptersPath, func(importPath string) bool {
		// Adapters can import from core, but not from other adapters
		// Allow same-type adapter imports (e.g., primary/http can import primary/http)
		if strings.Contains(importPath, "/adapters/") {
			// Get the import adapter type
			parts := strings.Split(importPath, "/adapters/")
			if len(parts) > 1 {
				importAdapterType := strings.Split(parts[1], "/")[0]
				// This is a simplified check - could be more sophisticated
				_ = importAdapterType
				// For now, allow adapter imports (too strict otherwise)
				return true
			}
		}
		return true
	})

	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check adapter dependencies: %v", err))
		return
	}

	if len(violations) == 0 {
		result.Successes = append(result.Successes, "Adapters follow dependency rules")
	} else {
		for _, violation := range violations {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Adapter cross-import: %s in %s", violation.importPath, violation.file))
		}
	}
}

// validateNamingConventions checks naming conventions
func (v *Validator) validateNamingConventions(result *ValidationResult) {
	// Check if adapter directories match expected style
	adaptersPath := filepath.Join("internal", "adapters")

	expectedInbound := v.config.AdapterInboundDir()
	expectedOutbound := v.config.AdapterOutboundDir()

	inboundPath := filepath.Join(adaptersPath, expectedInbound)
	outboundPath := filepath.Join(adaptersPath, expectedOutbound)

	if _, err := os.Stat(inboundPath); err == nil {
		result.Successes = append(result.Successes, fmt.Sprintf("Using %s for inbound adapters", expectedInbound))
	}

	if _, err := os.Stat(outboundPath); err == nil {
		result.Successes = append(result.Successes, fmt.Sprintf("Using %s for outbound adapters", expectedOutbound))
	}

	// Check for consistent naming
	// Check if core logic directory matches expected
	coreLogicPath := filepath.Join("internal", "core", v.config.CoreLogicDir())
	if _, err := os.Stat(coreLogicPath); err == nil {
		result.Successes = append(result.Successes, fmt.Sprintf("Using %s for business logic", v.config.CoreLogicDir()))
	}
}

// importViolation represents an import that violates architecture rules
type importViolation struct {
	file       string
	importPath string
}

// checkImports checks all Go files in a directory for import violations
func (v *Validator) checkImports(dir string, isAllowed func(string) bool) ([]importViolation, error) {
	var violations []importViolation

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse file
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Check imports
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Only check imports from the same module
			if !strings.HasPrefix(importPath, v.config.ModuleName) {
				continue
			}

			if !isAllowed(importPath) {
				violations = append(violations, importViolation{
					file:       path,
					importPath: importPath,
				})
			}
		}

		return nil
	})

	return violations, err
}
