// Package cmd defines the Cobra command-line interface subcommands and persistent flags for the configforge tool.
//
// # CLI Overview
//
// The command-line interface provides developers with tools to manage specifications, generate code/schemas,
// and validate local configuration files:
//
//	configforge defaults    - Prints a list of all fields, types, and defaults.
//	configforge validate    - Validates configuration files against semantic rules.
//	configforge generate    - Generates typed Go structures and JSON Schema files.
//	configforge schema      - Prints or writes the JSON Schema document.
//	configforge docs        - Generates reference Markdown documentation.
//	configforge version     - Prints version, commit hash, and build date.
package cmd
