package schema

import (
	"strings"
	"unicode"
)

// ToGoName converts snake_case, kebab-case, or dotted YAML keys to PascalCase Go names.
func ToGoName(s string) string {
	var parts []string
	var current []rune
	for _, r := range s {
		if r == '_' || r == '-' || r == '.' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	var result []string
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		result = append(result, string(runes))
	}
	return strings.Join(result, "")
}
