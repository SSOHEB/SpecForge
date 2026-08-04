package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/SSOHEB/SpecForge/internal/parser"
	"github.com/SSOHEB/SpecForge/internal/runtime"
	"github.com/SSOHEB/SpecForge/internal/schema"
	"github.com/SSOHEB/SpecForge/internal/validator"
)

func main() {
	configPath := os.Getenv("SPECFORGE_CONFIG_PATH")
	if configPath == "" {
		configPath = "examples/postgres/config.yaml"
	}

	// 1. Load metadata and build AST
	rawMeta, err := parser.ParseFile("examples/postgres/metadata.yaml")
	if err != nil {
		fmt.Printf("failed to parse metadata: %v\n", err)
		os.Exit(1)
	}
	ast, err := schema.Build(rawMeta)
	if err != nil {
		fmt.Printf("failed to build AST: %v\n", err)
		os.Exit(1)
	}

	// 2. Load config, apply defaults & overrides
	cfg, rawConfig, err := runtime.LoadAndPrepareFile[Config](ast, configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 3. Validate raw config
	valErrs := validator.Validate(ast, rawConfig)
	if len(valErrs) > 0 {
		fmt.Println("Validation failed:")
		for _, valErr := range valErrs {
			fmt.Println(valErr.Error())
		}
		os.Exit(1)
	}

	fmt.Println("config valid")
	fmt.Println("Resolved Postgres Configuration (Functional API):")
	fmt.Printf("  Host: %s\n", cfg.Postgres().Host())
	fmt.Printf("  Port: %d\n", cfg.Postgres().Port())
	fmt.Printf("  Database: %s\n", cfg.Postgres().Database())
	fmt.Printf("  User: %s\n", cfg.Postgres().User())
	fmt.Printf("  SSL Mode: %s\n", cfg.Postgres().SslMode())
	fmt.Printf("  Max Open Conns: %d\n", cfg.Postgres().MaxOpenConns())
	fmt.Printf("  Max Idle Conns: %d\n", cfg.Postgres().MaxIdleConns())
	fmt.Printf("  Max Lifetime: %d\n", cfg.Postgres().ConnMaxLifetime())
	fmt.Printf("  Retry Attempts: %d\n", cfg.Postgres().RetryAttempts())
	fmt.Printf("  Retry Backoff: %d ms\n", cfg.Postgres().RetryBackoffMs())

	// Reconnection retry check: We wrap a TCP connection attempt to Postgres in a loop
	// governed by the config's RetryAttempts and RetryBackoffMs values.
	attempts := cfg.Postgres().RetryAttempts()
	backoff := time.Duration(cfg.Postgres().RetryBackoffMs()) * time.Millisecond
	address := fmt.Sprintf("%s:%d", cfg.Postgres().Host(), cfg.Postgres().Port())
	dialTimeout := 2 * time.Second

	fmt.Printf("\nAttempting connection to Postgres at %s...\n", address)
	for attempt := 0; attempt <= attempts; attempt++ {
		conn, err := net.DialTimeout("tcp", address, dialTimeout)
		if err == nil {
			conn.Close()
			fmt.Println("Connection succeeded: Postgres is reachable!")
			return
		}

		fmt.Printf("Attempt %d/%d failed: %v\n", attempt+1, attempts+1, err)
		if attempt < attempts {
			fmt.Printf("Retrying in %v...\n", backoff)
			time.Sleep(backoff)
		}
	}

	fmt.Println("All connection attempts failed: Postgres is unreachable.")
}
