package validator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/SSOHEB/SpecForge/internal/schema"
)

// Validate checks the raw configuration map against the semantic rules of the AST.
// It collects all validation errors in one pass, sorts them by path, and returns them.
func Validate(ast *schema.AST, raw map[string]any) ValidationErrors {
	var errs ValidationErrors
	if ast == nil || ast.Root == nil {
		return errs
	}

	validateNode(ast.Root, raw, &errs)

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Path < errs[j].Path
	})

	return errs
}

func validateNode(node *schema.Node, raw map[string]any, errs *ValidationErrors) {
	for _, f := range node.Fields {
		val, exists := raw[f.YAMLKey]
		if !exists || val == nil {
			if f.Required {
				*errs = append(*errs, ValidationError{
					Path: strings.Join(f.Path, "."),
					Msg:  "required field is missing",
					Rule: "required",
				})
			}
			continue
		}

		validateField(f, val, errs)
	}

	for _, child := range node.Children {
		val, exists := raw[child.YAMLKey]
		if !exists || val == nil {
			flagAllRequiredFields(child, errs)
			continue
		}

		childMap := convertToMapStringAny(val)
		if childMap == nil {
			flagAllRequiredFields(child, errs)
			continue
		}

		validateNode(child, childMap, errs)
	}
}

func validateField(f *schema.Field, val any, errs *ValidationErrors) {
	dottedPath := strings.Join(f.Path, ".")

	switch f.Type {
	case schema.TypeInt, schema.TypeFloat:
		var numVal float64
		hasNum := false
		switch v := val.(type) {
		case int:
			numVal = float64(v)
			hasNum = true
		case int64:
			numVal = float64(v)
			hasNum = true
		case float64:
			numVal = v
			hasNum = true
		}

		if hasNum {
			if f.Min != nil && numVal < *f.Min {
				*errs = append(*errs, ValidationError{
					Path: dottedPath,
					Msg:  fmt.Sprintf("value %v is less than minimum %v", val, *f.Min),
					Rule: "min",
				})
			}
			if f.Max != nil && numVal > *f.Max {
				*errs = append(*errs, ValidationError{
					Path: dottedPath,
					Msg:  fmt.Sprintf("value %v is greater than maximum %v", val, *f.Max),
					Rule: "max",
				})
			}
		}

	case schema.TypeString:
		strVal, ok := val.(string)
		if ok {
			if len(f.Enum) > 0 {
				found := false
				for _, e := range f.Enum {
					if strVal == e {
						found = true
						break
					}
				}
				if !found {
					*errs = append(*errs, ValidationError{
						Path: dottedPath,
						Msg:  fmt.Sprintf("value %q is not one of the allowed enum values: %v", strVal, f.Enum),
						Rule: "enum",
					})
				}
			}

			if f.Pattern != "" {
				matched, err := regexp.MatchString(f.Pattern, strVal)
				if err != nil || !matched {
					*errs = append(*errs, ValidationError{
						Path: dottedPath,
						Msg:  fmt.Sprintf("value %q does not match pattern %q", strVal, f.Pattern),
						Rule: "pattern",
					})
				}
			}
		}
	}
}

func flagAllRequiredFields(node *schema.Node, errs *ValidationErrors) {
	for _, f := range node.Fields {
		if f.Required {
			*errs = append(*errs, ValidationError{
				Path: strings.Join(f.Path, "."),
				Msg:  "required field is missing",
				Rule: "required",
			})
		}
	}
	for _, child := range node.Children {
		flagAllRequiredFields(child, errs)
	}
}

func convertToMapStringAny(val any) map[string]any {
	if m, ok := val.(map[string]any); ok {
		return m
	}
	if m, ok := val.(map[any]any); ok {
		res := make(map[string]any)
		for k, v := range m {
			res[fmt.Sprintf("%v", k)] = v
		}
		return res
	}
	return nil
}
