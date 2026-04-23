package utils

import (
	"go/token"
	"strings"
)

// ToPlural returns a simple English plural of s (sufficient for typical entity names).
func ToPlural(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.HasSuffix(s, "y"):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

// toSnakeCase converts PascalCase to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToTitleCase converts a string to title case (simple implementation)
func ToTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// LcFirst lowercases the first letter of s.
func LcFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// SafeParamName returns a safe Go parameter name from a field name.
// If the lowercased name is a reserved keyword, it appends "Val".
func SafeParamName(name string) string {
	param := LcFirst(name)
	if token.IsKeyword(param) {
		return param + "Val"
	}
	return param
}

// ZeroValueFor returns the Go zero-value literal for a given type string.
func ZeroValueFor(typ string) string {
	switch typ {
	case "string":
		return `""`
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "0"
	case "bool":
		return "false"
	case "time.Time":
		return "time.Time{}"
	case "[]byte":
		return "nil"
	default:
		if strings.HasPrefix(typ, "*") || strings.HasPrefix(typ, "[]") || strings.HasPrefix(typ, "map[") {
			return "nil"
		}
		return typ + "{}"
	}
}
