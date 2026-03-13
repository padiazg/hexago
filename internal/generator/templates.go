package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/padiazg/hexago/pkg/utils"
)

const (
	makefileTemplate            string = "makefile"
	gitignoreTemplate           string = "gitignore"
	mainTemplate                string = "main"
	runTemplate                 string = "run"
	rootTemplate                string = "root"
	processorTemplate           string = "processor"
	configTemplate              string = "config"
	loggerTemplate              string = "logger"
	httpServerInterfaceTemplate string = "http-server-interface"
	httpServerFileTemplate      string = "http-server-file"
	httpServerHandlerTemplate   string = "http-server-handler"
	httpAdapterTemplate         string = "http-adapter"
	httpPingTemplate            string = "http-ping"
	httpHealthTemplate          string = "http-health"
	httpMetricsTemplate         string = "http-metrics"
	readmeTemplate              string = "readme"
	dockerFileTemplate          string = "dockerfile"
	composeTemplate             string = "compose"
	healthTemplate              string = "health"
	metricsTemplate             string = "metrics"
)

type templateItem struct {
	source string
	target string
}

type templateFn func(g *ProjectGenerator) templateItem

var templateMap = map[string]templateFn{
	makefileTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "misc/makefile.tmpl",
			target: "Makefile",
		}
	},
	gitignoreTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "misc/gitignore.tmpl",
			target: ".gitignore",
		}
	},
	mainTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "project/main.go.tmpl",
			target: "main.go",
		}
	},
	runTemplate: func(g *ProjectGenerator) templateItem {
		var templateName string
		switch g.config.ProjectType {
		case "http-server":
			templateName = "cmd/run_http_server.go.tmpl"
		case "service":
			templateName = "cmd/run_service.go.tmpl"
		default:
			// Fallback to http-server for backward compatibility
			templateName = "cmd/run.go.tmpl"
		}

		return templateItem{
			source: templateName,
			target: filepath.Join("cmd", "run.go"),
		}
	},
	rootTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "cmd/root.go.tmpl",
			target: filepath.Join("cmd", "root.go"),
		}
	},
	processorTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "service/processor.go.tmpl",
			target: filepath.Join("internal", "core", g.config.CoreLogicDir(), "processor.go"),
		}
	},
	configTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "project/config.go.tmpl",
			target: filepath.Join("internal", "config", "config.go"),
		}
	},
	loggerTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "project/logger.go.tmpl",
			target: filepath.Join("pkg", "logger", "logger.go"),
		}
	},
	httpServerInterfaceTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "pkg/server/server_interface.go.tmpl",
			target: filepath.Join("pkg", "server", "server.go"),
		}
	},
	httpServerFileTemplate: func(g *ProjectGenerator) templateItem {
		// TODO: manage this at the config package
		framework := g.config.Framework
		if framework == "" {
			framework = "stdlib"
		}

		return templateItem{
			source: fmt.Sprintf("pkg/httpserver/http_server_%s.go.tmpl", framework),
			target: filepath.Join("pkg", "httpserver", "server.go"),
		}
	},
	// httpServerHandlerTemplate: func(g *ProjectGenerator) templateItem {
	// 	framework := g.config.Framework
	// 	if framework == "" {
	// 		framework = "stdlib"
	// 	}

	// 	return templateItem{
	// 		source: fmt.Sprintf("pkg/httpserver/http_server_handler_%s.go.tmpl", framework),
	// 		target: filepath.Join("pkg", "httpserver", "handler.go"),
	// 	}
	// },
	httpAdapterTemplate: func(g *ProjectGenerator) templateItem {
		framework := g.config.Framework
		if framework == "" {
			framework = "stdlib"
		}

		return templateItem{
			source: fmt.Sprintf("adapter/primary/http/%s/http_adapter.go.tmpl", framework),
			target: filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", "http.go"),
		}
	},
	httpPingTemplate: func(g *ProjectGenerator) templateItem {
		framework := g.config.Framework
		if framework == "" {
			framework = "stdlib"
		}

		return templateItem{
			source: fmt.Sprintf("adapter/primary/http/%s/http_ping.go.tmpl", framework),
			target: filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", "ping", "ping.go"),
		}
	},
	httpHealthTemplate: func(g *ProjectGenerator) templateItem {
		framework := g.config.Framework
		if framework == "" {
			framework = "stdlib"
		}

		return templateItem{
			source: fmt.Sprintf("adapter/primary/http/%s/http_health.go.tmpl", framework),
			target: filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", "health", "health.go"),
		}
	},
	httpMetricsTemplate: func(g *ProjectGenerator) templateItem {
		framework := g.config.Framework
		if framework == "" {
			framework = "stdlib"
		}

		return templateItem{
			source: fmt.Sprintf("adapter/primary/http/%s/http_metrics.go.tmpl", framework),
			target: filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", "metrics", "metrics.go"),
		}
	},
	readmeTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "misc/readme.md.tmpl",
			target: "README.md",
		}
	},
	dockerFileTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "misc/dockerfile.tmpl",
			target: "Dockerfile",
		}
	},
	composeTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "misc/compose.yaml.tmpl",
			target: "compose.yaml",
		}
	},
	healthTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "observability/health.go.tmpl",
			target: filepath.Join("internal", "observability", "health.go"),
		}
	},
	metricsTemplate: func(g *ProjectGenerator) templateItem {
		return templateItem{
			source: "observability/metrics.go.tmpl",
			target: filepath.Join("internal", "observability", "metrics.go"),
		}
	},
}

// generatefile generates a given file
func (g *ProjectGenerator) generateFile(name string) error {
	item, ok := templateMap[name]
	if !ok {
		return fmt.Errorf("undefined %s template", name)
	}

	templ := item(g)

	content, err := g.config.templateLoader.Render(templ.source, g.config)
	if err != nil {
		return fmt.Errorf("failed to render %s template: %w", templ.source, err)
	}

	return utils.WriteFile(filepath.Join(g.projectPath, templ.target), content)
}

// renderTemplate renders a template with the given data
func (g *ProjectGenerator) renderTemplate(tmplStr string, data interface{}) ([]byte, error) {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"upper": func(s string) string {
			// Simple uppercase - can be enhanced
			return s
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]-32) + s[1:]
		},
	}).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
