package adapter

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/utils"
)

var _ Adapter = (*cliAdapter)(nil)

type cliAdapter struct {
	config *generator.HexagoConfig
	name   string
}

func newCLIAdapter(config *generator.HexagoConfig, name string) *cliAdapter {
	return &cliAdapter{
		config: config,
		name:   name,
	}
}

func (g *cliAdapter) Generate() error {
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "cli")

	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	filePath := filepath.Join(adapterDir, utils.ToSnakeCase(g.name)+".go")

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	if len(g.name) < 2 {
		return fmt.Errorf("cli adapter name is too short")
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	useName := utils.ToSnakeCase(g.name)
	data := map[string]any{
		"CommandName": g.name,
		"UseName":     useName,
	}

	content, err := g.config.TemplateLoader.Render("adapter/primary/cli.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render cli adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
