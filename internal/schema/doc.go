// Package schema compiles intermediate specifications into a semantic Abstract Syntax Tree (AST).
//
// The AST compiled by this package serves as the single, validated source of truth for
// all downstream code generation, documentation creation, and runtime validation components.
//
// # Pipeline Role
//
// Schema acts as the compiler and semantic validator of the pipeline:
//
//	[parser] -> [schema (AST Build)] -> [generator (Go/JSON Schema/Markdown)]
//	                                 -> [validator (Runtime Checks)]
//	                                 -> [runtime (Loader/Watcher/Overrides)]
//
// # Semantic Validations
//
// During the compilation of the AST (via schema.Build), the package enforces:
//   - Unique naming within namespaces (sibling name collision check between nested nodes and fields).
//   - Verification that default values match the corresponding FieldType (e.g. an integer default for a string field triggers an error).
//   - Correct compilation of regular expression strings for Pattern validators, surfacing invalid regex errors at compile time.
//
// # Example
//
//	rawSpec, _ := parser.ParseFile("metadata.yaml")
//	ast, err := schema.Build(rawSpec)
//	if err != nil {
//		log.Fatalf("invalid configuration schema spec: %v", err)
//	}
package schema
