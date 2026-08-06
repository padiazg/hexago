package adapter

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/utils"
)

var _ Adapter = (*queueAdapter)(nil)

type queueAdapter struct {
	config *generator.HexagoConfig
	name   string
}

func newQueueAdapter(config *generator.HexagoConfig, name string) *queueAdapter {
	return &queueAdapter{
		config: config,
		name:   name,
	}
}

func (g *queueAdapter) Generate() error {
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "queue")

	if err := utils.CreateDir(adapterDir); err != nil {
		return err
	}

	filePath := filepath.Join(adapterDir, utils.ToSnakeCase(g.name)+".go")

	if utils.FileExists(filePath) {
		return fmt.Errorf("adapter file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating adapter file: %s\n", filePath)

	data := map[string]any{
		"ModuleName":   g.config.Project.Module,
		"CoreLogic":    g.config.CoreLogicDir(),
		"ConsumerName": g.name,
	}

	content, err := g.config.TemplateLoader.Render("adapter/queue.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render queue adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
