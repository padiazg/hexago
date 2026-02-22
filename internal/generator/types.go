package generator

import "time"

// ProjectConfig holds the configuration for generating a new project
type ProjectConfig struct {
	// Basic project information
	ProjectName string // e.g., "my-app"
	ModuleName  string // e.g., "github.com/user/my-app"
	OutputDir   string // Where to create the project

	// Project type and architecture choices
	ProjectType  string // "http-server", "service", "job", "cli"
	Framework    string // "echo", "gin", "chi", "fiber", "stdlib" (only for http-server)
	AdapterStyle string // "primary-secondary" or "driver-driven"
	CoreLogic    string // "services" or "usecases"

	// Metadata
	Year      int
	Author    string
	GoVersion string

	// Optional features
	WithDocker        bool
	WithExample       bool
	WithMigrations    bool
	WithMetrics       bool
	ExplicitPorts     bool // Create explicit ports/ directory
	WithWorkers       bool
	WithObservability bool

	templateLoader *TemplateLoader
}

// NewProjectConfig creates a new ProjectConfig with sensible defaults
func NewProjectConfig(projectName, moduleName string) *ProjectConfig {
	return &ProjectConfig{
		ProjectName:       projectName,
		ModuleName:        moduleName,
		OutputDir:         ".",
		ProjectType:       "http-server", // Default for backward compatibility
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
		templateLoader:    NewTemplateLoader(),
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

// IsHTTPServer returns true if project is an HTTP API server
func (c *ProjectConfig) IsHTTPServer() bool {
	return c.ProjectType == "http-server"
}

// IsService returns true if project is a long-running service/daemon
func (c *ProjectConfig) IsService() bool {
	return c.ProjectType == "service"
}

// NeedsWebFramework returns true if project needs a web framework for main logic
func (c *ProjectConfig) NeedsWebFramework() bool {
	return c.IsHTTPServer()
}
