package generator

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkProjectGeneratorprintSuccessFn func(*testing.T, string)

var checkProjectGeneratorprintSuccess = func(fns ...checkProjectGeneratorprintSuccessFn) []checkProjectGeneratorprintSuccessFn { return fns }

func TestProjectGenerator_printSuccess(t *testing.T) {
	captureStdOut := func(f func()) string {
		var (
			orig    = os.Stdout
			r, w, _ = os.Pipe()
		)

		os.Stdout = w
		f()
		os.Stdout = orig
		w.Close()
		out, _ := io.ReadAll(r)
		return string(out)
	}

	contains := func(sub string) checkProjectGeneratorprintSuccessFn {
		return func(t *testing.T, got string) {
			t.Helper()
			if !strings.Contains(got, sub) {
				t.Errorf("output should contain %q, got:\n%s", sub, got)
			}
		}
	}

	notContains := func(sub string) checkProjectGeneratorprintSuccessFn {
		return func(t *testing.T, got string) {
			t.Helper()
			if strings.Contains(got, sub) {
				t.Errorf("output should NOT contain %q, got:\n%s", sub, got)
			}
		}
	}

	tests := []struct {
		name        string
		module      string
		projectName string
		projType    string
		checks      []checkProjectGeneratorprintSuccessFn
	}{
		{
			name:        "http-server project",
			projectName: "my-service",
			projType:    "http-server",
			checks: checkProjectGeneratorprintSuccess(
				contains("✅ Project generated successfully!"),
				contains("📚 Next steps:"),
				contains("cd my-service"),
				contains("go run main.go run"),
				contains("📖 Read the README.md for more information about the project structure."),
				notContains("version"),
			),
		},
		{
			name:        "cli project",
			projectName: "my-cli-tool",
			projType:    "cli",
			checks: checkProjectGeneratorprintSuccess(
				contains("✅ Project generated successfully!"),
				contains("cd my-cli-tool"),
				contains("go run main.go version"),
				notContains("go run main.go run"),
			),
		},
		{
			name:        "project with special name",
			projectName: "user-api",
			projType:    "http-server",
			checks: checkProjectGeneratorprintSuccess(
				contains("cd user-api"),
				contains("go run main.go run"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &ProjectGenerator{
				config: HexagoConfig{
					Project: HexagoProjectConfig{
						Name: tt.projectName,
						Type: tt.projType,
					},
				},
			}

			got := captureStdOut(func() {
				s.printSuccess()
			})

			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}

type checkProjectGeneratordependenciesListFn func(*testing.T, []string)

var checkProjectGeneratordependenciesList = func(fns ...checkProjectGeneratordependenciesListFn) []checkProjectGeneratordependenciesListFn {
	return fns
}

func TestProjectGenerator_dependenciesList(t *testing.T) {
	contains := func(want string) checkProjectGeneratordependenciesListFn {
		return func(t *testing.T, s []string) {
			t.Helper()
			assert.Contains(t, s, want)
		}
	}

	notContains := func(want string) checkProjectGeneratordependenciesListFn {
		return func(t *testing.T, s []string) {
			t.Helper()
			assert.NotContains(t, s, want)
		}
	}

	tests := []struct {
		name           string
		framework      string
		withMetrics    bool
		withObs        bool
		withMigrations bool
		withFiberObs   bool
		checks         []checkProjectGeneratordependenciesListFn
	}{
		{
			name:      "base dependencies only",
			framework: "echo",
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/spf13/cobra@latest"),
				contains("github.com/spf13/viper@latest"),
				contains("github.com/labstack/echo/v4@latest"),
				notContains("prometheus"),
				notContains("golang-migrate"),
				notContains("x/sync"),
			),
		},
		{
			name:      "gin framework",
			framework: "gin",
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/gin-gonic/gin@latest"),
				notContains("echo"),
				notContains("chi"),
				notContains("fiber"),
			),
		},
		{
			name:      "chi framework",
			framework: "chi",
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/go-chi/chi/v5@latest"),
				notContains("echo"),
				notContains("gin"),
			),
		},
		{
			name:      "fiber framework",
			framework: "fiber",
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/gofiber/fiber/v2@latest"),
				notContains("echo"),
				notContains("gin"),
				notContains("chi"),
			),
		},
		{
			name:        "with metrics",
			framework:   "echo",
			withMetrics: true,
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/prometheus/client_golang@latest"),
			),
		},
		{
			name:      "with observability",
			framework: "echo",
			withObs:   true,
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/prometheus/client_golang@latest"),
				contains("golang.org/x/sync@latest"),
			),
		},
		{
			name:           "with migrations",
			framework:      "echo",
			withMigrations: true,
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/golang-migrate/migrate/v4@latest"),
			),
		},
		{
			name:         "fiber with observability needs adaptor",
			framework:    "fiber",
			withObs:      true,
			withFiberObs: true,
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/gofiber/fiber/v2@latest"),
				contains("golang.org/x/sync@latest"),
				contains("github.com/prometheus/client_golang@latest"),
				contains("github.com/gofiber/adaptor/v2@latest"),
			),
		},
		{
			name:           "fiber without observability has no adaptor",
			framework:      "fiber",
			withMigrations: true,
			checks: checkProjectGeneratordependenciesList(
				contains("github.com/gofiber/fiber/v2@latest"),
				contains("github.com/golang-migrate/migrate/v4@latest"),
				notContains("gofiber/adaptor"),
				notContains("x/sync"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &ProjectGenerator{
				config: HexagoConfig{
					Project: HexagoProjectConfig{
						Framework: tt.framework,
					},
					Features: HexagoFeaturesConfig{
						WithMetrics:       tt.withMetrics,
						WithObservability: tt.withObs,
						WithMigrations:    tt.withMigrations,
					},
				},
			}
			r := s.dependenciesList()
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}
