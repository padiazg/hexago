package analyzer

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// FindInterfaces discovers all interfaces (ports) in the given packages.
func FindInterfaces(pkgs []*packages.Package) ([]PortInfo, error) {
	var ports []PortInfo

	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				portInfo := extractPortInfo(pkg, name, iface)
				ports = append(ports, portInfo)
			}
		}
	}

	return ports, nil
}

// FindInterfaceByName finds a specific interface by name.
func FindInterfaceByName(pkgs []*packages.Package, name string) (*PortInfo, error) {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()
		obj := scope.Lookup(name)

		if obj == nil {
			continue
		}

		if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
			portInfo := extractPortInfo(pkg, name, iface)
			return &portInfo, nil
		}
	}

	return nil, fmt.Errorf("interface %q not found", name)
}

// extractPortInfo converts a types.Interface to PortInfo.
func extractPortInfo(pkg *packages.Package, name string, iface *types.Interface) PortInfo {
	methods := make([]MethodInfo, 0, iface.NumMethods())

	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		sig := method.Type().(*types.Signature)

		methodInfo := MethodInfo{
			Name:    method.Name(),
			Params:  extractParams(sig.Params()),
			Returns: extractParams(sig.Results()),
		}
		methods = append(methods, methodInfo)
	}

	return PortInfo{
		Name:       name,
		Package:    pkg.Name,
		ImportPath: pkg.PkgPath,
		Methods:    methods,
	}
}

// extractParams extracts parameters from a *types.Tuple.
func extractParams(tuple *types.Tuple) []ParamInfo {
	if tuple == nil {
		return nil
	}

	params := make([]ParamInfo, 0, tuple.Len())

	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)

		params = append(params, ParamInfo{
			Name: param.Name(),
			Type: typeToString(param.Type()),
		})
	}

	return params
}

// typeToString converts a types.Type to a human-readable string.
func typeToString(t types.Type) string {
	if t == nil {
		return ""
	}

	switch v := t.(type) {
	case *types.Named:
		obj := v.Obj()
		objName := obj.Name()

		// Handle commonly used types with nil-safe checks
		if obj.Pkg() != nil {
			pkgPath := obj.Pkg().Path()
			switch objName {
			case "Context":
				// Check if it's context.Context
				if pkgPath == "context" {
					return "context.Context"
				}
			case "Error":
				if pkgPath == "errors" {
					return "error"
				}
			case "User", "URL", "Category", "Product", "Order":
				// For domain types, check if they're from our project
				// Return simple name - user will add import in template
				return objName
			}
		}
		return objName
	case *types.Pointer:
		return "*" + typeToString(v.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", v.Len(), typeToString(v.Elem()))
	case *types.Slice:
		return "[]" + typeToString(v.Elem())
	case *types.Map:
		return "map[" + typeToString(v.Key()) + "]" + typeToString(v.Elem())
	case *types.Chan:
		return v.String()
	case *types.Signature:
		return v.String()
	case *types.Basic:
		return v.Name()
	default:
		return t.String()
	}
}
