package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const HexagoConfigFile = ".hexago.yaml"

const hexagoConfigHeader = `# .hexago.yaml - HexaGo project configuration
# Created by ` + "`hexago init`" + `. Edit with care.

`

// HexagoConfig is the top-level structure for .hexago.yaml
type HexagoConfig struct {
	Project   HexagoProjectConfig   `yaml:"project"`
	Structure HexagoStructureConfig `yaml:"structure"`
	Features  HexagoFeaturesConfig  `yaml:"features"`
}

// HexagoProjectConfig holds basic project metadata
type HexagoProjectConfig struct {
	Name      string `yaml:"name"`
	Module    string `yaml:"module"`
	Type      string `yaml:"type"`
	Framework string `yaml:"framework,omitempty"`
	GoVersion string `yaml:"go_version"`
	Author    string `yaml:"author,omitempty"`
}

// HexagoStructureConfig holds architecture naming conventions
type HexagoStructureConfig struct {
	AdapterStyle  string `yaml:"adapter_style"`
	CoreLogic     string `yaml:"core_logic"`
	ExplicitPorts bool   `yaml:"explicit_ports"`
}

// HexagoFeaturesConfig holds optional feature flags
type HexagoFeaturesConfig struct {
	WithDocker        bool `yaml:"with_docker"`
	WithObservability bool `yaml:"with_observability"`
	WithMigrations    bool `yaml:"with_migrations"`
	WithWorkers       bool `yaml:"with_workers"`
	WithMetrics       bool `yaml:"with_metrics"`
	WithExample       bool `yaml:"with_example"`
}

// HexagoConfigFromProject maps a ProjectConfig to a HexagoConfig.
func HexagoConfigFromProject(cfg *ProjectConfig) *HexagoConfig {
	return &HexagoConfig{
		Project: HexagoProjectConfig{
			Name:      cfg.ProjectName,
			Module:    cfg.ModuleName,
			Type:      cfg.ProjectType,
			Framework: cfg.Framework,
			GoVersion: cfg.GoVersion,
			Author:    cfg.Author,
		},
		Structure: HexagoStructureConfig{
			AdapterStyle:  cfg.AdapterStyle,
			CoreLogic:     cfg.CoreLogic,
			ExplicitPorts: cfg.ExplicitPorts,
		},
		Features: HexagoFeaturesConfig{
			WithDocker:        cfg.WithDocker,
			WithObservability: cfg.WithObservability,
			WithMigrations:    cfg.WithMigrations,
			WithWorkers:       cfg.WithWorkers,
			WithMetrics:       cfg.WithMetrics,
			WithExample:       cfg.WithExample,
		},
	}
}

// ToProjectConfig maps a HexagoConfig back to a ProjectConfig.
func (h *HexagoConfig) ToProjectConfig() *ProjectConfig {
	cfg := NewProjectConfig(h.Project.Name, h.Project.Module)

	cfg.ProjectType = h.Project.Type
	cfg.Framework = h.Project.Framework
	cfg.GoVersion = h.Project.GoVersion
	cfg.Author = h.Project.Author
	cfg.Year = time.Now().Year()

	cfg.AdapterStyle = h.Structure.AdapterStyle
	cfg.CoreLogic = h.Structure.CoreLogic
	cfg.ExplicitPorts = h.Structure.ExplicitPorts

	cfg.WithDocker = h.Features.WithDocker
	cfg.WithObservability = h.Features.WithObservability
	cfg.WithMigrations = h.Features.WithMigrations
	cfg.WithWorkers = h.Features.WithWorkers
	cfg.WithMetrics = h.Features.WithMetrics
	cfg.WithExample = h.Features.WithExample

	return cfg
}

// LoadHexagoConfig reads and parses {dir}/.hexago.yaml.
// Returns an error if the file does not exist or cannot be parsed.
func LoadHexagoConfig(dir string) (*HexagoConfig, error) {
	path := filepath.Join(dir, HexagoConfigFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", HexagoConfigFile, err)
	}

	var cfg HexagoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", HexagoConfigFile, err)
	}

	return &cfg, nil
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
