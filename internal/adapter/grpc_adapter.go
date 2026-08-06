package adapter

import (
	"fmt"
	"path/filepath"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/padiazg/hexago/pkg/utils"
)

var _ Adapter = (*grpcAdapter)(nil)

type grpcAdapter struct {
	config *generator.HexagoConfig
	name   string
}

func newGRPCAdapter(config *generator.HexagoConfig, name string) *grpcAdapter {
	return &grpcAdapter{
		config: config,
		name:   name,
	}
}

func (g *grpcAdapter) Generate() error {
	adapterDir := filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "grpc")

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

	content, err := g.config.TemplateLoader.Render("adapter/primary/grpc.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render gRPC adapter template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}
