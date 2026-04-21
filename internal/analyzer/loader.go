package analyzer

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

// LoadProject loads all Go packages in a project directory.
// Walks through ./internal/core/... to find packages.
func LoadProject(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedName,
		Dir: dir,
	}

	pkgs, err := packages.Load(cfg, "./internal/core/...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	var valid []*packages.Package
	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			valid = append(valid, pkg)
		} else {
			printPackageErrors(pkg)
		}
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid packages found in %s/internal/core", dir)
	}

	return valid, nil
}

// LoadSinglePackage loads a single package by import path.
func LoadSinglePackage(dir, importPath string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedName,
		Dir: dir,
	}

	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", importPath, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s not found", importPath)
	}

	return pkgs[0], nil
}

// HasProject returns true if dir contains a Go project.
func HasProject(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// printPackageErrors prints errors for a package (for debugging).
func printPackageErrors(pkg *packages.Package) {
	for _, err := range pkg.Errors {
		fmt.Fprintf(os.Stderr, "Warning: package %s: %v\n", pkg.Name, err)
	}
}
