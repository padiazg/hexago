package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const HexagoConfigFile = ".hexago.yaml"

const hexagoConfigHeader = `# .hexago.yaml - HexaGo project configuration
# Created by ` + "`hexago init`" + `. Edit with care.

`

// HexagoConfig is the top-level structure for .hexago.yaml.
// It is the single runtime config type used throughout the codebase.
type HexagoConfig struct {
	TemplateLoader *TemplateLoader       `yaml:"-" mapstructure:"-"`
	OutputDir      string                `yaml:"-" mapstructure:"-"`
	Structure      HexagoStructureConfig `yaml:"structure" mapstructure:"structure"`
	Project        HexagoProjectConfig   `yaml:"project" mapstructure:"project"`
	Features       HexagoFeaturesConfig  `yaml:"features" mapstructure:"features"`
	InPlace        bool                  `yaml:"-" mapstructure:"-"`
}

// HexagoStructureConfig holds architecture naming conventions
type HexagoStructureConfig struct {
	AdapterStyle  string `yaml:"adapter_style" mapstructure:"adapter_style"`
	CoreLogic     string `yaml:"core_logic" mapstructure:"core_logic"`
	ExplicitPorts bool   `yaml:"explicit_ports" mapstructure:"explicit_ports"`
}

// HexagoFeaturesConfig holds optional feature flags
type HexagoFeaturesConfig struct {
	WithDocker        bool `yaml:"with_docker" mapstructure:"with_docker"`
	WithObservability bool `yaml:"with_observability" mapstructure:"with_observability"`
	WithMigrations    bool `yaml:"with_migrations" mapstructure:"with_migrations"`
	WithWorkers       bool `yaml:"with_workers" mapstructure:"with_workers"`
	WithMetrics       bool `yaml:"with_metrics" mapstructure:"with_metrics"`
	WithExample       bool `yaml:"with_example" mapstructure:"with_example"`
	WithTests         bool `yaml:"with_tests" mapstructure:"with_tests"`
}

// NewHexagoConfig creates a new HexagoConfig with sensible defaults
func NewHexagoConfig(
	projectName, moduleName, outputDir, projectType, framework, adapterStyle, coreLogic string,
	withDocker, withExample, withMigrations, withMetrics, explicitPorts, withWorkers, withObservability, withTests bool,
) HexagoConfig {
	config := HexagoConfig{
		Project: HexagoProjectConfig{
			Name:           projectName,
			Module:         moduleName,
			Type:           projectType,
			Framework:      framework,
			DatabaseDriver: "postgres",
			GoVersion:      "1.21",
			Year:           time.Now().Year(),
		},
		Structure: HexagoStructureConfig{
			AdapterStyle:  adapterStyle,
			CoreLogic:     coreLogic,
			ExplicitPorts: explicitPorts,
		},
		Features: HexagoFeaturesConfig{
			WithDocker:        withDocker,
			WithExample:       withExample,
			WithMigrations:    withMigrations,
			WithMetrics:       withMetrics,
			WithWorkers:       withWorkers,
			WithObservability: withObservability,
			WithTests:         withTests,
		},
		OutputDir:      outputDir,
		TemplateLoader: NewTemplateLoader(),
	}

	config.Validate()

	return config
}

// AdapterInboundDir returns the directory name for inbound adapters
func (c *HexagoConfig) AdapterInboundDir() string {
	if c.Structure.AdapterStyle == "driver-driven" {
		return "driver"
	}
	return "primary"
}

// AdapterOutboundDir returns the directory name for outbound adapters
func (c *HexagoConfig) AdapterOutboundDir() string {
	if c.Structure.AdapterStyle == "driver-driven" {
		return "driven"
	}
	return "secondary"
}

// CoreLogicDir returns the directory name for business logic
func (c *HexagoConfig) CoreLogicDir() string {
	return c.Structure.CoreLogic
}

// IsHTTPServer returns true if project is an HTTP API server
func (c *HexagoConfig) IsHTTPServer() bool {
	return c.Project.Type == "http-server"
}

// IsService returns true if project is a long-running service/daemon
func (c *HexagoConfig) IsService() bool {
	return c.Project.Type == "service"
}

// IsCLI returns true if project is a batch CLI with subcommands
func (c *HexagoConfig) IsCLI() bool {
	return c.Project.Type == "cli"
}

// NeedsWebFramework returns true if project needs a web framework for main logic
func (c *HexagoConfig) NeedsWebFramework() bool {
	return c.IsHTTPServer()
}

// -- Project field getters --

// ModuleName returns the Go module name (e.g. "github.com/user/my-app")
func (c *HexagoConfig) ModuleName() string { return c.Project.Module }

// ProjectName returns the project name
func (c *HexagoConfig) ProjectName() string { return c.Project.Name }

// Year returns the copyright year
func (c *HexagoConfig) Year() int { return c.Project.Year }

// Author returns the project author name
func (c *HexagoConfig) Author() string { return c.Project.Author }

// GoVersion returns the Go version used for building
func (c *HexagoConfig) GoVersion() string { return c.Project.GoVersion }

// -- Structure field getters --

// AdapterStyle returns the adapter naming style (primary-secondary|driver-driven)
func (c *HexagoConfig) AdapterStyle() string { return c.Structure.AdapterStyle }

// CoreLogic returns the core logic directory name (services|usecases)
func (c *HexagoConfig) CoreLogic() string { return c.Structure.CoreLogic }

// -- Feature flag getters --

func (c *HexagoConfig) WithDocker() bool        { return c.Features.WithDocker }
func (c *HexagoConfig) WithObservability() bool { return c.Features.WithObservability }
func (c *HexagoConfig) WithMigrations() bool    { return c.Features.WithMigrations }
func (c *HexagoConfig) WithWorkers() bool       { return c.Features.WithWorkers }
func (c *HexagoConfig) WithMetrics() bool       { return c.Features.WithMetrics }
func (c *HexagoConfig) WithExample() bool       { return c.Features.WithExample }
func (c *HexagoConfig) WithTests() bool         { return c.Features.WithTests }

func (c *HexagoConfig) Validate() error {
	c.Project.Validate()

	if err := c.validateModuleName(); err != nil {
		return err
	}

	if err := c.validateProjectType(); err != nil {
		return err
	}

	if c.Project.Type == "http-server" {
		if err := c.validateFramework(); err != nil {
			return err
		}
	}

	if err := c.validateAdapterStyle(); err != nil {
		return err
	}

	if err := c.validateCoreLogic(); err != nil {
		return err
	}

	return nil
}

func (c *HexagoConfig) validateModuleName() error {
	if c.Project.Module == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if !strings.Contains(c.Project.Module, "/") && !strings.Contains(c.Project.Module, ".") {
		fmt.Printf("⚠️  Warning: module name '%s' doesn't follow Go module naming convention (domain.com/user/project)\n", c.Project.Module)
	}

	return nil
}

func (c *HexagoConfig) validateProjectType() error {
	validTypes := map[string]bool{
		"http-server": true,
		"service":     true,
		"cli":         true,
	}

	if !validTypes[c.Project.Type] {
		return fmt.Errorf("invalid project type '%s'. Valid options: http-server, service, cli", c.Project.Type)
	}

	return nil
}

func (c *HexagoConfig) validateFramework() error {
	validFrameworks := map[string]bool{
		"echo":   true,
		"gin":    true,
		"chi":    true,
		"fiber":  true,
		"stdlib": true,
	}

	if !validFrameworks[c.Project.Framework] {
		return fmt.Errorf("invalid framework '%s'. Valid options: echo, gin, chi, fiber, stdlib", c.Project.Framework)
	}

	return nil
}

func (c *HexagoConfig) validateAdapterStyle() error {
	validStyles := map[string]bool{
		"primary-secondary": true,
		"driver-driven":     true,
	}

	if !validStyles[c.Structure.AdapterStyle] {
		return fmt.Errorf("invalid adapter style '%s'. Valid options: primary-secondary, driver-driven", c.Structure.AdapterStyle)
	}

	return nil
}

func (c *HexagoConfig) validateCoreLogic() error {
	validNames := map[string]bool{
		"services": true,
		"usecases": true,
	}

	if !validNames[c.Structure.CoreLogic] {
		return fmt.Errorf("invalid core logic name '%s'. Valid options: services, usecases", c.Structure.CoreLogic)
	}

	return nil
}

// LoadHexagoConfig reads and parses {dir}/.hexago.yaml.
// Returns an error if the file does not exist or cannot be parsed.
func LoadHexagoConfig(dir string) (*HexagoConfig, error) {
	v := viper.New()
	v.SetConfigName(".hexago")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read %s: %w", HexagoConfigFile, err)
	}

	var cfg HexagoConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", HexagoConfigFile, err)
	}

	cfg.TemplateLoader = NewTemplateLoader()
	cfg.Project.Validate()

	return &cfg, nil
}

// LoadInitConfig loads project configuration for the init command using viper.
// Priority: CLI flag > .hexago.yaml > defaults.
// Only returns an error if the config file exists but cannot be parsed.
func LoadInitConfig(v *viper.Viper, dir string, cmd *cobra.Command) (*HexagoConfig, error) {
	setInitDefaults(v)

	v.SetConfigName(".hexago")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read %s: %w", HexagoConfigFile, err)
		}
	}

	bindInitFlags(v, cmd)

	var cfg HexagoConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.OutputDir = dir
	cfg.TemplateLoader = NewTemplateLoader()

	return &cfg, nil
}

func setInitDefaults(v *viper.Viper) {
	v.SetDefault("project.type", "http-server")
	v.SetDefault("project.framework", "stdlib")
	v.SetDefault("project.database_driver", "postgres")
	v.SetDefault("project.go_version", "1.21")
	v.SetDefault("project.year", time.Now().Year())
	v.SetDefault("structure.adapter_style", "primary-secondary")
	v.SetDefault("structure.core_logic", "services")
	v.SetDefault("structure.explicit_ports", false)
	v.SetDefault("features.with_docker", false)
	v.SetDefault("features.with_example", false)
	v.SetDefault("features.with_migrations", false)
	v.SetDefault("features.with_metrics", false)
	v.SetDefault("features.with_workers", false)
	v.SetDefault("features.with_observability", false)
	v.SetDefault("features.with_tests", false)
}

func bindInitFlags(v *viper.Viper, cmd *cobra.Command) {
	bindings := []struct{ viperKey, flagName string }{
		{"project.module", "module"},
		{"project.type", "project-type"},
		{"project.framework", "framework"},
		{"project.database_driver", "db-driver"},
		{"structure.adapter_style", "adapter-style"},
		{"structure.core_logic", "core-logic"},
		{"structure.explicit_ports", "explicit-ports"},
		{"features.with_docker", "with-docker"},
		{"features.with_example", "with-example"},
		{"features.with_migrations", "with-migrations"},
		{"features.with_metrics", "with-metrics"},
		{"features.with_workers", "with-workers"},
		{"features.with_observability", "with-observability"},
		{"features.with_tests", "with-tests"},
	}
	for _, b := range bindings {
		_ = v.BindPFlag(b.viperKey, cmd.Flags().Lookup(b.flagName))
	}
}

// SaveHexagoConfig serializes cfg and writes it to {dir}/.hexago.yaml,
// prepending a comment header.
func SaveHexagoConfig(dir string, cfg *HexagoConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", HexagoConfigFile, err)
	}

	content := hexagoConfigHeader + string(data)

	path := filepath.Join(dir, HexagoConfigFile)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", HexagoConfigFile, err)
	}

	return nil
}
