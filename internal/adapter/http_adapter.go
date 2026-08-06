package adapter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/utils"
)

func newHttpAdapter(config *generator.HexagoConfig, name, entityName string) Adapter {
	if entityName != "" {
		return newHttpHandlerAdapter(config, entityName)
	}

	return newHttpFlatAdapter(config, name)
}

var _ Adapter = (*httpFlatAdapter)(nil)

type httpFlatAdapter struct {
	config *generator.HexagoConfig
	name   string
}

func newHttpFlatAdapter(config *generator.HexagoConfig, name string) *httpFlatAdapter {
	return &httpFlatAdapter{
		config: config,
		name:   name,
	}
}

func (g *httpFlatAdapter) Generate() error {
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http")

	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	filePath := filepath.Join(adapterDir, utils.ToSnakeCase(g.name)+".go")

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	data := map[string]any{
		"ModuleName":  g.config.Project.Module,
		"CoreLogic":   g.config.CoreLogicDir(),
		"HandlerName": g.name,
	}

	content, err := g.config.TemplateLoader.Render("adapter/primary/http.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render HTTP adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

var _ Adapter = (*httpHandlerAdapter)(nil)

type httpHandlerAdapter struct {
	config     *generator.HexagoConfig
	entityName string
}

func newHttpHandlerAdapter(config *generator.HexagoConfig, entityName string) *httpHandlerAdapter {
	return &httpHandlerAdapter{
		config:     config,
		entityName: entityName,
	}
}

func (g *httpHandlerAdapter) Generate() error {
	pkgName := utils.ToPlural(strings.ToLower(g.entityName))
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", pkgName)

	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	configFile := filepath.Join(adapterDir, utils.ToSnakeCase(g.entityName)+".go")
	handlersFile := filepath.Join(adapterDir, "handlers.go")

	if utils.FileExists(configFile) {
		return fmt.Errorf("handler file %s already exists", configFile)
	}

	entityVarName := strings.ToLower(g.entityName[:1]) + g.entityName[1:]
	servicePkgName := pkgName
	serviceImportAlias := pkgName + "Svc"
	entityImportAlias := pkgName + "Domain"
	serviceField := utils.ToTitleCase(pkgName)
	routePrefix := pkgName

	data := map[string]any{
		"ModuleName":         g.config.Project.Module,
		"CoreLogic":          g.config.CoreLogicDir(),
		"PackageName":        pkgName,
		"EntityName":         g.entityName,
		"EntityVarName":      entityVarName,
		"EntityPackage":      pkgName,
		"EntityImportAlias":  entityImportAlias,
		"ServicePackage":     servicePkgName,
		"ServiceImportAlias": serviceImportAlias,
		"ServiceName":        g.entityName,
		"ServiceField":       serviceField,
		"RoutePrefix":        routePrefix,
	}

	framework := g.config.Project.Framework

	fmt.Printf("📝 Creating handler config file: %s\n", configFile)
	configTmpl := fmt.Sprintf("adapter/primary/http/%s/handler_config.go.tmpl", framework)
	configContent, err := g.config.TemplateLoader.Render(configTmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render handler config template: %w", err)
	}
	if err := utils.WriteFile(configFile, configContent); err != nil {
		return err
	}

	fmt.Printf("📝 Creating handler methods file: %s\n", handlersFile)
	methodsTmpl := fmt.Sprintf("adapter/primary/http/%s/handler_methods.go.tmpl", framework)
	methodsContent, err := g.config.TemplateLoader.Render(methodsTmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render handler methods template: %w", err)
	}
	return utils.WriteFile(handlersFile, methodsContent)
}
