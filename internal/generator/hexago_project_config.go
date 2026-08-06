package generator

// HexagoProjectConfig holds basic project metadata
type HexagoProjectConfig struct {
	Name           string `yaml:"name" mapstructure:"name"`
	Module         string `yaml:"module" mapstructure:"module"`
	Type           string `yaml:"type" mapstructure:"type"`
	Framework      string `yaml:"framework,omitempty" mapstructure:"framework"`
	DatabaseDriver string `yaml:"database_driver,omitempty" mapstructure:"database_driver"`
	GoVersion      string `yaml:"go_version" mapstructure:"go_version"`
	Author         string `yaml:"author,omitempty" mapstructure:"author"`
	Year           int    `yaml:"year,omitempty" mapstructure:"year"`
}

func (c *HexagoProjectConfig) Validate() {
	if c.Framework == "" {
		c.Framework = "stdlib"
	}
}
