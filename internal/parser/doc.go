// Package parser parses raw metadata specification YAML files into intermediate structure maps.
//
// The parser maps YAML namespaces and fields to strongly typed FieldSpec and NodeSpec configurations,
// performing structural validations and basic rule sanitization.
//
// # Pipeline Role
//
// The parser is the first stage of the codrao pipeline:
//
//	[YAML Spec] -> [parser] -> [schema (AST)] -> [generator/validator/runtime]
//
// It consumes the metadata file, validates syntax and structural rules, and produces a raw
// specification model. This model is then passed to internal/schema to be built into a validated
// semantic Abstract Syntax Tree (AST).
//
// # Validations Performed
//
// During parsing, the package checks for:
//   - Syntactically malformed YAML documents.
//   - Missing or unknown type definitions (only bool, int, float, string, string[], int[], map[string]string are allowed).
//   - Invalid attribute combinations (e.g. specifying "min"/"max" constraints on bool or string fields,
//     specifying "enum" constraints on bools, or using regex "pattern" checks on integers).
//
// # Example
//
//	rawSpec, err := parser.ParseFile("metadata.yaml")
//	if err != nil {
//		log.Fatalf("metadata parsing failed: %v", err)
//	}
//	// Pass rawSpec to internal/schema.Build...
package parser
