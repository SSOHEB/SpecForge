// Package cmd defines the Cobra command-line interface subcommands and persistent flags for the specforge tool.
//
// # CLI Overview
//
// The command-line interface provides developers with tools to manage specifications, generate code/schemas,
// and validate local configuration files:
//
//	specforge defaults    - Prints a list of all fields, types, and defaults.
//	specforge validate    - Validates configuration files against semantic rules.
//	specforge generate    - Generates typed Go structures and JSON Schema files.
//	specforge schema      - Prints or writes the JSON Schema document.
//	specforge docs        - Generates reference Markdown documentation.
//	specforge version     - Prints version, commit hash, and build date.
package cmd
