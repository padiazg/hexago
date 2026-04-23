package analyzer

import (
	"go/token"
)

// PackageInfo contains basic information about a loaded Go package.
type PackageInfo struct {
	Name       string
	ImportPath string
	FileSet    *token.FileSet
}

// MethodInfo represents a single method signature.
type MethodInfo struct {
	Name    string
	Params  []ParamInfo
	Returns []ParamInfo
}

// ParamInfo represents a function parameter or return value.
type ParamInfo struct {
	Name       string
	Type       string
	ImportPath string // Full import path for types from external packages
}

// DomainStruct represents a discovered struct in the domain layer.
type DomainStruct struct {
	Name       string
	Package    string
	ImportPath string
	Fields     []FieldInfo
}

// FieldInfo represents a struct field.
type FieldInfo struct {
	Name string
	Type string
}
