package analyzer

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// FindDomainStructs discovers all structs in the given packages.
func FindDomainStructs(pkgs []*packages.Package) []DomainStruct {
	var structs []DomainStruct

	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			if _, ok := obj.Type().Underlying().(*types.Struct); ok {
				domainStruct := extractDomainStruct(pkg, name, obj.Type().(*types.Named))
				structs = append(structs, domainStruct)
			}
		}
	}

	return structs
}

// FindDomainStructByName finds a specific struct by name.
func FindDomainStructByName(pkgs []*packages.Package, name string) (*DomainStruct, error) {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()
		obj := scope.Lookup(name)

		if obj == nil {
			continue
		}

		if _, ok := obj.Type().Underlying().(*types.Struct); ok {
			domainStruct := extractDomainStruct(pkg, name, obj.Type().(*types.Named))
			return &domainStruct, nil
		}
	}

	return nil, fmt.Errorf("struct %q not found", name)
}

// extractDomainStruct converts a types.Named to DomainStruct.
func extractDomainStruct(pkg *packages.Package, name string, named *types.Named) DomainStruct {
	var fields []FieldInfo

	if structType, ok := named.Underlying().(*types.Struct); ok {
		fields = make([]FieldInfo, 0, structType.NumFields())

		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)

			fields = append(fields, FieldInfo{
				Name: field.Name(),
				Type: typeToString(field.Type()),
			})
		}
	}

	return DomainStruct{
		Name:       name,
		Package:    pkg.Name,
		ImportPath: pkg.PkgPath,
		Fields:     fields,
	}
}
