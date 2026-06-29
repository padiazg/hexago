package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type HexagoConfigCheckFn func(*testing.T, HexagoConfig)

var checkHexagoConfig = func(fns ...HexagoConfigCheckFn) []HexagoConfigCheckFn { return fns }

func checkProjectName(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Project.Name, "Project.Name = %s, expected %s", cfg.Project.Name, want)
	}
}

func checkModuleName(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Project.Module, "Project.Module = %s, expected %s", cfg.Project.Module, want)
	}
}

func checkOutputDir(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.OutputDir, "OutputDir = %s, expected %s", cfg.OutputDir, want)
	}
}

func checkProjectType(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Project.Type, "Project.Type = %s, expected %s", cfg.Project.Type, want)
	}
}

func checkFramework(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Project.Framework, "Project.Framework = %s, expected %s", cfg.Project.Framework, want)
	}
}

func checkAdapterStyle(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Structure.AdapterStyle, "Structure.AdapterStyle = %s, expected %s", cfg.Structure.AdapterStyle, want)
	}
}

func checkCoreLogic(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Structure.CoreLogic, "Structure.CoreLogic = %s, expected %s", cfg.Structure.CoreLogic, want)
	}
}

func checkWithDocker(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithDocker, "Features.WithDocker = %t, expected %t", cfg.Features.WithDocker, want)
	}
}

func checkWithExample(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithExample, "Features.WithExample = %t, expected %t", cfg.Features.WithExample, want)
	}
}

func checkWithMigrations(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithMigrations, "Features.WithMigrations = %t, expected %t", cfg.Features.WithMigrations, want)
	}
}

func checkWithMetrics(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithMetrics, "Features.WithMetrics = %t, expected %t", cfg.Features.WithMetrics, want)
	}
}

func checkExplicitPorts(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Structure.ExplicitPorts, "Structure.ExplicitPorts = %t, expected %t", cfg.Structure.ExplicitPorts, want)
	}
}

func checkWithWorkers(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithWorkers, "Features.WithWorkers = %t, expected %t", cfg.Features.WithWorkers, want)
	}
}

func checkWithObservability(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithObservability, "Features.WithObservability = %t, expected %t", cfg.Features.WithObservability, want)
	}
}

func checkWithTests(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Features.WithTests, "Features.WithTests = %t, expected %t", cfg.Features.WithTests, want)
	}
}

func checkInPlace(want bool) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.InPlace, "InPlace = %t, expected %t", cfg.InPlace, want)
	}
}

func checkGoVersion(want string) HexagoConfigCheckFn {
	return func(t *testing.T, cfg HexagoConfig) {
		t.Helper()
		assert.Equalf(t, want, cfg.Project.GoVersion, "Project.GoVersion = %s, expected %s", cfg.Project.GoVersion, want)
	}
}

func TestNewHexagoConfig(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		moduleName   string
		outputDir    string
		projectType  string
		framework    string
		adapterStyle string
		coreLogic    string
		checks       []HexagoConfigCheckFn
	}{
		{
			name:         "default values with project and module name",
			projectName:  "myapp",
			moduleName:   "github.com/user/myapp",
			outputDir:    ".",
			projectType:  "http-server",
			framework:    "stdlib",
			adapterStyle: "primary-secondary",
			coreLogic:    "services",
			checks: checkHexagoConfig(
				checkProjectName("myapp"),
				checkModuleName("github.com/user/myapp"),
				checkOutputDir("."),
				checkProjectType("http-server"),
				checkFramework("stdlib"),
				checkAdapterStyle("primary-secondary"),
				checkCoreLogic("services"),
				checkGoVersion("1.21"),
				checkWithDocker(false),
				checkWithExample(false),
				checkWithMigrations(false),
				checkWithMetrics(false),
				checkExplicitPorts(false),
				checkWithWorkers(false),
				checkWithObservability(false),
				checkWithTests(false),
				checkInPlace(false),
			),
		},
		{
			name:         "empty project name",
			projectName:  "",
			moduleName:   "github.com/user/myapp",
			outputDir:    ".",
			projectType:  "http-server",
			framework:    "stdlib",
			adapterStyle: "primary-secondary",
			coreLogic:    "services",
			checks: checkHexagoConfig(
				checkProjectName(""),
				checkModuleName("github.com/user/myapp"),
				checkOutputDir("."),
				checkProjectType("http-server"),
				checkGoVersion("1.21"),
			),
		},
		{
			name:         "empty module name",
			projectName:  "myapp",
			moduleName:   "",
			outputDir:    ".",
			projectType:  "http-server",
			framework:    "stdlib",
			adapterStyle: "primary-secondary",
			coreLogic:    "services",
			checks: checkHexagoConfig(
				checkProjectName("myapp"),
				checkModuleName(""),
				checkOutputDir("."),
				checkProjectType("http-server"),
				checkGoVersion("1.21"),
			),
		},
		{
			name:         "both project and module name empty",
			projectName:  "",
			moduleName:   "",
			outputDir:    ".",
			projectType:  "http-server",
			framework:    "stdlib",
			adapterStyle: "primary-secondary",
			coreLogic:    "services",
			checks: checkHexagoConfig(
				checkProjectName(""),
				checkModuleName(""),
				checkOutputDir("."),
				checkProjectType("http-server"),
				checkGoVersion("1.21"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := NewHexagoConfig(tt.projectName, tt.moduleName, tt.outputDir, tt.projectType, tt.framework, tt.adapterStyle, tt.coreLogic, false, false, false, false, false, false, false, false)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

func TestHexagoConfig_AdapterInboundDir(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		before func(*HexagoConfig)
	}{
		{
			name: "default adapter style returns primary",
			want: "primary",
		},
		{
			name: "driver-driven adapter style returns driver",
			want: "driver",
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = "driver-driven"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			r := s.AdapterInboundDir()
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestHexagoConfig_AdapterOutboundDir(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		before func(*HexagoConfig)
	}{
		{
			name: "default adapter style returns secondary",
			want: "secondary",
		},
		{
			name: "driver-driven adapter style returns driven",
			want: "driven",
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = "driver-driven"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			r := s.AdapterOutboundDir()
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestHexagoConfig_CoreLogicDir(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		before func(*HexagoConfig)
	}{
		{
			name: "default core logic returns services",
			want: "services",
		},
		{
			name: "usecases core logic returns usecases",
			want: "usecases",
			before: func(cfg *HexagoConfig) {
				cfg.Structure.CoreLogic = "usecases"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			r := s.CoreLogicDir()
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestHexagoConfig_IsHTTPServer(t *testing.T) {
	tests := []struct {
		name   string
		want   bool
		before func(*HexagoConfig)
	}{
		{
			name: "default project type returns true",
			want: true,
		},
		{
			name: "service project type returns false",
			want: false,
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = "service"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			r := s.IsHTTPServer()
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestHexagoConfig_IsService(t *testing.T) {
	tests := []struct {
		name   string
		want   bool
		before func(*HexagoConfig)
	}{
		{
			name: "default project type returns false",
			want: false,
		},
		{
			name: "service project type returns true",
			want: true,
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = "service"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			r := s.IsService()
			assert.Equal(t, tt.want, r)
		})
	}
}

type checkHexagoConfigValidateFn func(*testing.T, error)

var checkHexagoConfigValidate = func(fns ...checkHexagoConfigValidateFn) []checkHexagoConfigValidateFn {
	return fns
}

func checkConfigValidateError(want string) checkHexagoConfigValidateFn {
	return func(t *testing.T, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkConfigValidateError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkConfigValidateError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkConfigValidateError mismatch")
		}
	}
}

func TestHexagoConfig_validateModuleName(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "valid module name with path",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/myapp"
			},
		},
		{
			name: "valid module name with dot only",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "myapp.com"
			},
		},
		{
			name: "empty module name returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("module name cannot be empty"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = ""
			},
		},
		{
			name: "module name without separator returns nil (prints warning)",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "myapp"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.validateModuleName()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

func TestHexagoConfig_validateProjectType(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "http-server is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = "http-server"
			},
		},
		{
			name: "service is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = "service"
			},
		},
		{
			name: "invalid project type returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid project type 'cli'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = "cli"
			},
		},
		{
			name: "empty project type returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid project type ''"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Type = ""
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.validateProjectType()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

func TestHexagoConfig_validateFramework(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "stdlib is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Framework = "stdlib"
			},
		},
		{
			name: "echo is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Framework = "echo"
			},
		},
		{
			name: "invalid framework returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid framework 'invalid'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Framework = "invalid"
			},
		},
		{
			name: "empty framework returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid framework ''"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Framework = ""
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.validateFramework()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

func TestHexagoConfig_validateAdapterStyle(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "primary-secondary is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = "primary-secondary"
			},
		},
		{
			name: "driver-driven is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = "driver-driven"
			},
		},
		{
			name: "invalid adapter style returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid adapter style 'flat'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = "flat"
			},
		},
		{
			name: "empty adapter style returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid adapter style ''"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.AdapterStyle = ""
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.validateAdapterStyle()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

func TestHexagoConfig_validateCoreLogic(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "services is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.CoreLogic = "services"
			},
		},
		{
			name: "usecases is valid",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.CoreLogic = "usecases"
			},
		},
		{
			name: "invalid core logic returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid core logic name 'models'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.CoreLogic = "models"
			},
		},
		{
			name: "empty core logic returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid core logic name ''"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Structure.CoreLogic = ""
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.validateCoreLogic()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

func TestHexagoConfig_Validate(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkHexagoConfigValidateFn
		before func(*HexagoConfig)
	}{
		{
			name: "all valid http-server defaults",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "http-server"
				cfg.Project.Framework = "stdlib"
				cfg.Structure.AdapterStyle = "primary-secondary"
				cfg.Structure.CoreLogic = "services"
			},
		},
		{
			name: "valid service with non-stdlib framework",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError(""),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "service"
				cfg.Project.Framework = "gin"
				cfg.Structure.AdapterStyle = "driver-driven"
				cfg.Structure.CoreLogic = "usecases"
			},
		},
		{
			name: "empty module name returns error first",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("module name cannot be empty"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = ""
				cfg.Project.Type = "http-server"
			},
		},
		{
			name: "invalid project type returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid project type 'cli'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "cli"
			},
		},
		{
			name: "invalid framework returns error for http-server",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid framework 'unknown'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "http-server"
				cfg.Project.Framework = "unknown"
			},
		},
		{
			name: "invalid adapter style returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid adapter style 'flat'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "http-server"
				cfg.Project.Framework = "stdlib"
				cfg.Structure.AdapterStyle = "flat"
			},
		},
		{
			name: "invalid core logic returns error",
			checks: checkHexagoConfigValidate(
				checkConfigValidateError("invalid core logic name 'models'"),
			),
			before: func(cfg *HexagoConfig) {
				cfg.Project.Module = "github.com/user/app"
				cfg.Project.Type = "http-server"
				cfg.Project.Framework = "stdlib"
				cfg.Structure.AdapterStyle = "primary-secondary"
				cfg.Structure.CoreLogic = "models"
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewHexagoConfig("value", "value", ".", "http-server", "stdlib", "primary-secondary", "services", false, false, false, false, false, false, false, false)
			if tt.before != nil {
				tt.before(&s)
			}
			err := s.Validate()
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}
