package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/pkg/utils"
)

// ProjectDetector detects existing project configuration
type ProjectDetector struct {
	projectPath string
}

// NewProjectDetector creates a new project detector
func NewProjectDetector(projectPath string) *ProjectDetector {
	return &ProjectDetector{
		projectPath: projectPath,
	}
}

// DetectConfig detects the project configuration from existing structure
func (d *ProjectDetector) DetectConfig() (*ProjectConfig, error) {
	// Try .hexago.yaml first — it has the full picture
	if hexCfg, err := LoadHexagoConfig(d.projectPath); err == nil {
		cfg := hexCfg.ToProjectConfig()
		// Always override with actual project values
		cfg.ProjectName = filepath.Base(d.projectPath)
		cfg.OutputDir = d.projectPath
		return cfg, nil
	}

	// Fall back to filesystem heuristics (legacy / non-hexago projects)

	// Verify we're in a Go project
	if !d.isGoProject() {
		return nil, fmt.Errorf("not a Go project (go.mod not found)")
	}

	// Verify hexagonal structure exists
	if !d.hasHexagonalStructure() {
		return nil, fmt.Errorf("not a hexagonal architecture project (internal/core not found)")
	}

	config := &ProjectConfig{}

	// Detect module name from go.mod
	moduleName, err := d.detectModuleName()
	if err != nil {
		return nil, err
	}
	config.ModuleName = moduleName

	// Detect project name from directory
	config.ProjectName = filepath.Base(d.projectPath)

	// Detect adapter style (primary-secondary vs driver-driven)
	config.AdapterStyle = d.detectAdapterStyle()

	// Detect core logic naming (services vs usecases)
	config.CoreLogic = d.detectCoreLogic()

	// Check for explicit ports
	config.ExplicitPorts = d.hasExplicitPorts()

	// Check for observability
	config.WithObservability = d.hasObservability()

	// Set output directory
	config.OutputDir = d.projectPath

	return config, nil
}

// isGoProject checks if go.mod exists
func (d *ProjectDetector) isGoProject() bool {
	goModPath := filepath.Join(d.projectPath, "go.mod")
	return utils.FileExists(goModPath)
}

// hasHexagonalStructure checks if internal/core exists
func (d *ProjectDetector) hasHexagonalStructure() bool {
	corePath := filepath.Join(d.projectPath, "internal", "core")
	return utils.FileExists(corePath)
}

// detectModuleName reads module name from go.mod
func (d *ProjectDetector) detectModuleName() (string, error) {
	goModPath := filepath.Join(d.projectPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return after, nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// detectAdapterStyle checks which adapter naming is used
func (d *ProjectDetector) detectAdapterStyle() string {
	adaptersPath := filepath.Join(d.projectPath, "internal", "adapters")

	// Check for primary/secondary
	primaryPath := filepath.Join(adaptersPath, "primary")
	if utils.FileExists(primaryPath) {
		return "primary-secondary"
	}

	// Check for driver/driven
	driverPath := filepath.Join(adaptersPath, "driver")
	if utils.FileExists(driverPath) {
		return "driver-driven"
	}

	// Default to primary-secondary
	return "primary-secondary"
}

// detectCoreLogic checks which core logic naming is used
func (d *ProjectDetector) detectCoreLogic() string {
	corePath := filepath.Join(d.projectPath, "internal", "core")

	// Check for services
	servicesPath := filepath.Join(corePath, "services")
	if utils.FileExists(servicesPath) {
		return "services"
	}

	// Check for usecases
	usecasesPath := filepath.Join(corePath, "usecases")
	if utils.FileExists(usecasesPath) {
		return "usecases"
	}

	// Default to services
	return "services"
}

// hasExplicitPorts checks if ports directory exists
func (d *ProjectDetector) hasExplicitPorts() bool {
	portsPath := filepath.Join(d.projectPath, "internal", "core", "ports")
	return utils.FileExists(portsPath)
}

// hasObservability checks if observability directory exists
func (d *ProjectDetector) hasObservability() bool {
	obsPath := filepath.Join(d.projectPath, "internal", "observability")
	return utils.FileExists(obsPath)
}

// GetCurrentProjectConfig detects the project in the given dir.
// If dir is empty, os.Getwd() is used.
func GetCurrentProjectConfig(dir string) (*ProjectConfig, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	detector := NewProjectDetector(dir)
	return detector.DetectConfig()
}
