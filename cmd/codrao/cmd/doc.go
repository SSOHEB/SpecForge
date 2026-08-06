// Package cmd defines the Cobra command-line interface subcommands and persistent flags for the codrao tool.
//
// # CLI Overview
//
// The command-line interface provides developers with tools to manage specifications, generate code/schemas,
// and validate local configuration files:
//
//	codrao defaults    - Prints a list of all fields, types, and defaults.
//	codrao validate    - Validates configuration files against semantic rules.
//	codrao generate    - Generates typed Go structures and JSON Schema files.
//	codrao schema      - Prints or writes the JSON Schema document.
//	codrao docs        - Generates reference Markdown documentation.
//	codrao version     - Prints version, commit hash, and build date.
package cmd
