// Package generator processes the semantic AST to produce Go code, JSON Schema documents, and Markdown files.
//
// The generator translates abstract configuration definitions into concrete files used during
// build time, editor integration, and documentation workflows.
//
// # Pipeline Role
//
// Generator is the backend compiler target of the configforge pipeline:
//
//	[schema (AST)] -> [generator] -> [generated_config.go] (Go Code)
//	                               -> [schema.json] (JSON Schema)
//	                               -> [CONFIGURATION.md] (Markdown Reference)
//
// # Outputs Generated
//
//   - Go Code (GenerateGoCode): Emits typed structures matching configuration namespaces.
//     Supports generating a Functional API (getters instead of direct field access). When functional
//     getters are requested, struct fields are suffixed with "Field" (e.g. "PortField") to prevent
//     method/field naming collisions in Go.
//   - JSON Schema (GenerateJSONSchema): Produces a Draft-07 JSON Schema document with stable key orders,
//     ideal for IDE completion and client-side form integrations.
//   - Markdown Reference (GenerateMarkdownDocs): Formats documentation tables mapping paths, defaults,
//     requirements, and constraint rules into a readable Markdown document.
package generator
