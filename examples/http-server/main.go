package main

import (
	"fmt"
	"os"

	"configforge/internal/parser"
	"configforge/internal/runtime"
	"configforge/internal/schema"
	"configforge/internal/validator"
)

func main() {
	configPath := os.Getenv("CONFIGFORGE_CONFIG_PATH")
	if configPath == "" {
		configPath = "examples/http-server/config.yaml"
	}

	// 1. Load metadata and build AST for semantic rules definition
	rawMeta, err := parser.ParseFile("examples/http-server/metadata.yaml")
	if err != nil {
		fmt.Printf("failed to parse metadata: %v\n", err)
		os.Exit(1)
	}
	ast, err := schema.Build(rawMeta)
	if err != nil {
		fmt.Printf("failed to build AST: %v\n", err)
		os.Exit(1)
	}

	// 2. Load config, apply defaults, and convert to typed struct in one step
	cfg, rawConfig, err := runtime.LoadAndPrepareFile[Config](ast, configPath)
	if err != nil {
		fmt.Printf("failed to load and prepare config: %v\n", err)
		os.Exit(1)
	}

	// 3. Validate raw configuration (post-defaults) against AST rules
	valErrs := validator.Validate(ast, rawConfig)
	if len(valErrs) > 0 {
		fmt.Println("Validation failed:")
		for _, valErr := range valErrs {
			fmt.Println(valErr.Error())
		}
		os.Exit(1)
	}

	fmt.Println("config valid")
	fmt.Printf("HTTP Server Port: %d (configured)\n", cfg.Instrumentation.Http.Port)
	fmt.Printf("HTTP Server Host: %s (default)\n", cfg.Instrumentation.Http.Host)
	fmt.Printf("HTTP Server Enabled: %v (default)\n", cfg.Instrumentation.Http.Enabled)
	fmt.Printf("HTTP Server Timeout: %d (default)\n", cfg.Instrumentation.Http.Timeout)
	fmt.Printf("HTTP Server LogLevel: %s (default)\n", cfg.Instrumentation.Http.LogLevel)
	fmt.Printf("HTTP Server ApiKey: %s (configured)\n", cfg.Instrumentation.Http.ApiKey)
}
