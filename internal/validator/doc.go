// Package validator validates unmarshaled configuration maps against semantic AST constraints.
//
// Unlike JSON Schema validation (which happens externally), this package executes directly
// in the Go application runtime, matching unmarshaled data with AST nodes.
//
// # Pipeline Role
//
// The validator checks loaded configurations for structural and value compliance at startup or reload:
//
//	[runtime (Load config map)] -> [validator.Validate] -> [Errors or Launch App]
//
// # Checks Enforced
//
// The validator walks the AST in parallel with the raw config map (`map[string]any`), asserting:
//   - Required Fields: Checks that the key is physically present in the map, avoiding false passes on zero values.
//   - Numeric Range Bounds: Verifies integer and float values satisfy min/max constraints.
//   - Enum Membership: Asserts string values match one of the declared options.
//   - Regular Expressions: Validates string fields against compiled regex patterns.
//
// # Example
//
//	var rawConfig map[string]any
//	yaml.Unmarshal(data, &rawConfig)
//
//	errs := validator.Validate(ast, rawConfig)
//	if len(errs) > 0 {
//		for _, err := range errs {
//			fmt.Println("Validation error:", err.Error())
//		}
//	}
package validator
