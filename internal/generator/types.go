package generator

import "time"

// ProjectConfig holds the configuration for generating a new project
type ProjectConfig struct {
	// Basic project information
	ProjectName string // e.g., "my-app"
	ModuleName  string // e.g., "github.com/user/my-app"
	OutputDir   string // Where to create the project

	// Framework and architecture choices
	Framework    string // "echo", "gin", "chi", "fiber", "stdlib"
	AdapterStyle string // "primary-secondary" or "driver-driven"
	CoreLogic    string // "services" or "usecases"

	// Optional features
	WithDocker      bool
	WithExample     bool
	WithMigrations  bool
	WithMetrics     bool
	ExplicitPorts   bool // Create explicit ports/ directory
	WithWorkers     bool
	WithObservability bool

	// Metadata
	GoVersion string
	Author    string
	Year      int
}

// NewProjectConfig creates a new ProjectConfig with sensible defaults
func NewProjectConfig(projectName, moduleName string) *ProjectConfig {
	return &ProjectConfig{
		ProjectName:       projectName,
		ModuleName:        moduleName,
		OutputDir:         ".",
		Framework:         "stdlib",
		AdapterStyle:      "primary-secondary",
		CoreLogic:         "services",
		WithDocker:        false,
		WithExample:       false,
		WithMigrations:    false,
		WithMetrics:       false,
		ExplicitPorts:     false,
		WithWorkers:       false,
		WithObservability: false,
		GoVersion:         "1.21",
		Author:            "",
		Year:              time.Now().Year(),
	}
}

// AdapterInboundDir returns the directory name for inbound adapters
func (c *ProjectConfig) AdapterInboundDir() string {
	if c.AdapterStyle == "driver-driven" {
		return "driver"
	}
	return "primary"
}

// AdapterOutboundDir returns the directory name for outbound adapters
func (c *ProjectConfig) AdapterOutboundDir() string {
	if c.AdapterStyle == "driver-driven" {
		return "driven"
	}
	return "secondary"
}

// CoreLogicDir returns the directory name for business logic
func (c *ProjectConfig) CoreLogicDir() string {
	return c.CoreLogic
}
