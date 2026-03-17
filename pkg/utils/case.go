package utils

import "strings"

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

// toTitleCase converts a string to title case (simple implementation)
func ToTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
