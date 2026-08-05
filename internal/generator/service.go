package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/padiazg/hexago/internal/analyzer"
	"github.com/padiazg/hexago/pkg/utils"
)

// ServiceEntry holds the metadata for one service used in the aggregator template.
type ServiceEntry struct {
	Package       string // e.g. "categories"
	Alias         string // e.g. "categoriesSvc"
	DomainAlias   string // e.g. "categoriesDomain"
	RepoField     string // e.g. "CategoriesRepository"
	RepoInterface string // e.g. "CategoryRepository"
	ServiceField  string // e.g. "Categories"
	ServiceType   string // e.g. "CategoryService"
	HasEntity     bool   // true when service is bound to a domain entity (has repo dependency)
}

// ServiceGenerator generates service/usecase files
type ServiceGenerator struct {
	config *HexagoConfig
}

// NewServiceGenerator creates a new service generator
func NewServiceGenerator(config *HexagoConfig) *ServiceGenerator {
	return &ServiceGenerator{
		config: config,
	}
}

// Generate creates a new service file in its own sub-package.
// entityName (optional) is the domain entity this service manages; when provided
// the sub-package name is derived from it (e.g. "Category" → "categories").
// When omitted, serviceName itself is used as the package name.
// portInfo (optional) provides method signatures for code generation.
func (g *ServiceGenerator) Generate(serviceName, entityName, description string, portInfo *analyzer.PortInfo) error {
	baseServiceDir := filepath.Join(g.config.OutputDir, "internal", "core", g.config.CoreLogicDir())
	if !utils.FileExists(baseServiceDir) {
		return fmt.Errorf("directory %s does not exist. Are you in a hexagonal project?", baseServiceDir)
	}

	// Derive package name and entity name
	hasEntity := entityName != ""
	var pkgName, resolvedEntity string
	if hasEntity {
		pkgName = utils.ToPlural(strings.ToLower(entityName))
		resolvedEntity = entityName
	} else {
		pkgName = strings.ToLower(serviceName)
		resolvedEntity = ""
	}

	serviceDir := filepath.Join("internal", "core", g.config.CoreLogicDir(), pkgName)
	if err := utils.CreateDir(serviceDir); err != nil {
		return fmt.Errorf("creating directory %s: %w", serviceDir, err)
	}

	fileName := pkgName + ".go"

	filePath := filepath.Join(g.config.OutputDir, serviceDir, fileName)

	if utils.FileExists(filePath) {
		return fmt.Errorf("service file %s already exists", filePath)
	}

	fmt.Printf("📝 Creating service file: %s\n", filePath)

	if err := g.generateServiceFile(filePath, serviceName, resolvedEntity, pkgName, description, hasEntity, portInfo); err != nil {
		return err
	}

	// Entity-bound services are generated with uuid.NewString() and need the dep.
	if hasEntity {
		g.ensureUUIDDep()
	}

	if err := g.upsertAggregator(baseServiceDir); err != nil {
		// Non-fatal: aggregator update failure should not block the service generation
		fmt.Printf("⚠️  Warning: failed to update services aggregator: %v\n", err)
	}

	return nil
}

// ensureUUIDDep runs `go get github.com/google/uuid` (non-fatal) since
// entity-bound services are generated with uuid.NewString().
func (g *ServiceGenerator) ensureUUIDDep() {
	cmd := exec.Command("go", "get", "github.com/google/uuid")
	cmd.Dir = g.config.OutputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("⚠️  Warning: could not add dependency github.com/google/uuid: %s\n", strings.TrimSpace(string(out)))
	}
}

// generateServiceFile generates the service implementation file
func (g *ServiceGenerator) generateServiceFile(filePath, serviceName, entityName, pkgName, description string, hasEntity bool, portInfo *analyzer.PortInfo) error {
	desc := description
	if desc == "" {
		if hasEntity {
			desc = fmt.Sprintf("handles %s operations", entityName)
		} else {
			desc = fmt.Sprintf("implements %s logic", serviceName)
		}
	}

	entityImportAlias := pkgName + "Domain"

	data := map[string]any{
		"CoreLogic":         g.config.CoreLogicDir(),
		"ModuleName":        g.config.Project.Module,
		"ServiceName":       serviceName,
		"PackageName":       pkgName,
		"HasEntity":         hasEntity,
		"EntityName":        entityName,
		"EntityPackage":     pkgName,
		"EntityImportAlias": entityImportAlias,
		"Description":       desc,
	}

	if portInfo != nil {
		data["Methods"] = portInfo.Methods
		data["PortName"] = portInfo.Name
	}

	content, err := g.config.templateLoader.Render("service/service.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render service template: %w", err)
	}

	return utils.WriteFile(filePath, content)
}

// upsertAggregator scans all service sub-packages and regenerates services.go.
func (g *ServiceGenerator) upsertAggregator(baseServiceDir string) error {
	entries, err := os.ReadDir(baseServiceDir)
	if err != nil {
		return fmt.Errorf("reading service dir: %w", err)
	}

	var serviceEntries []ServiceEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()
		srcFile := filepath.Join(baseServiceDir, pkgName, pkgName+".go")
		entityName, hasEntity, err := g.extractServiceInfo(srcFile)
		if err != nil {
			continue // not a service package — skip silently
		}
		serviceEntries = append(serviceEntries, ServiceEntry{
			Package:       pkgName,
			Alias:         pkgName + "Svc",
			DomainAlias:   pkgName + "Domain",
			RepoField:     utils.ToTitleCase(pkgName) + "Repository",
			RepoInterface: entityName + "Repository",
			ServiceField:  utils.ToTitleCase(pkgName),
			ServiceType:   serviceTypeName(entityName, hasEntity),
			HasEntity:     hasEntity,
		})
	}

	if len(serviceEntries) == 0 {
		return nil
	}

	aggregatorPath := filepath.Join(baseServiceDir, "services.go")
	data := map[string]any{
		"ModuleName": g.config.Project.Module,
		"CoreLogic":  g.config.CoreLogicDir(),
		"Entries":    serviceEntries,
	}

	content, err := g.config.templateLoader.Render("service/services_aggregator.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render aggregator template: %w", err)
	}

	fmt.Printf("📝 Updating services aggregator: %s\n", aggregatorPath)
	return utils.WriteFile(aggregatorPath, content)
}

// serviceTypeName returns the generated service struct name for the given entity.
// Entity-bound services generate "XxxService"; plain use cases generate "Xxx".
func serviceTypeName(entityName string, hasEntity bool) string {
	if hasEntity {
		return entityName + "Service"
	}
	return entityName
}

// extractServiceInfo scans a service Go file for the first `type XxxService struct`
// (or `type Xxx struct` for plain use cases) declaration and returns the entity
// name ("Xxx") plus whether the service is entity-bound (i.e. it imports from
// internal/core/domain/).
func (g *ServiceGenerator) extractServiceInfo(filePath string) (entityName string, hasEntity bool, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", false, err
	}
	re := regexp.MustCompile(`type (\w+?)(?:Service)? struct`)
	for _, m := range re.FindAllSubmatch(content, -1) {
		if len(m) < 2 {
			continue
		}
		name := string(m[1])
		if strings.HasSuffix(name, "Input") || strings.HasSuffix(name, "Output") || name == "Config" {
			continue
		}
		entityName = name
		break
	}
	if entityName == "" {
		return "", false, fmt.Errorf("no XxxService struct found in %s", filePath)
	}
	hasEntity = strings.Contains(string(content), `/internal/core/domain/`)
	return entityName, hasEntity, nil
}
