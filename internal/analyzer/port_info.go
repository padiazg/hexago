package analyzer

import "strings"

// PortInfo represents a discovered interface (port) in the domain/services layer.
type PortInfo struct {
	Name       string
	Package    string
	ImportPath string
	Methods    []MethodInfo
}

// Domains returns a list of domains from the port
func (pi *PortInfo) Domains(moduleName string) []string {
	// Collect domain imports from method parameters
	var domainImports []string

	for _, method := range pi.Methods {
		for _, param := range method.Params {
			if param.ImportPath != "" && param.ImportPath != moduleName {
				domainImports = append(domainImports, param.ImportPath)
			}
		}

		for _, ret := range method.Returns {
			if ret.ImportPath != "" && ret.ImportPath != moduleName {
				domainImports = append(domainImports, ret.ImportPath)
			}
		}
	}

	return domainImports
}

// DomainAliasMap returns an alias map for the domains from the port
func (pi *PortInfo) DomainAliasMap(moduleName string) map[string]string {
	domainImports := pi.Domains(moduleName)
	domainAliasMap := make(map[string]string) // importPath -> alias

	if len(domainImports) > 0 {
		seen := make(map[string]bool)
		for _, imp := range domainImports {
			if !seen[imp] {
				seen[imp] = true
				// Use last path component as alias (e.g., "joke" from ".../domain/joke")
				alias := getLastPathComponent(imp)
				domainAliasMap[imp] = alias
			}
		}
	}

	return domainAliasMap
}

// getLastPathComponent returns the last component of a import path.
// e.g., "github.com/user/project/internal/core/domain/joke" -> "joke"
func getLastPathComponent(importPath string) string {
	parts := strings.Split(importPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return importPath
}
