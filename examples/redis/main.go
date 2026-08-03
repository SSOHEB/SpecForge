package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"configforge/internal/parser"
	"configforge/internal/runtime"
	"configforge/internal/schema"
	"configforge/internal/validator"
)

func main() {
	configPath := os.Getenv("CONFIGFORGE_CONFIG_PATH")
	if configPath == "" {
		configPath = "examples/redis/config.yaml"
		if _, err := os.Stat("config.yaml"); err == nil {
			configPath = "config.yaml"
		}
	}

	metadataFile := "examples/redis/metadata.yaml"
	if _, err := os.Stat("metadata.yaml"); err == nil {
		metadataFile = "metadata.yaml"
	}

	// 1. Load metadata and build AST
	rawMeta, err := parser.ParseFile(metadataFile)
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
	fmt.Println("Resolved Redis Configuration (Functional API):")
	fmt.Printf("  Host: %s\n", cfg.Redis().Host())
	fmt.Printf("  Port: %d\n", cfg.Redis().Port())
	fmt.Printf("  DB: %d\n", cfg.Redis().Db())
	fmt.Printf("  Pool Size: %d\n", cfg.Redis().PoolSize())
	fmt.Printf("  Dial Timeout: %d\n", cfg.Redis().DialTimeout())
	fmt.Printf("  Read Timeout: %d\n", cfg.Redis().ReadTimeout())
	fmt.Printf("  Write Timeout: %d\n", cfg.Redis().WriteTimeout())
	fmt.Printf("  TLS Enabled: %v\n", cfg.Redis().TlsEnabled())
	fmt.Printf("  Sentinel Addrs: %v\n", cfg.Redis().SentinelAddrs())

	// Reachability check choice: We perform a raw TCP dial using the loaded host, port, and dial timeout.
	// This avoids external third-party dependencies while keeping the reachability check buildable and honest.
	address := fmt.Sprintf("%s:%d", cfg.Redis().Host(), cfg.Redis().Port())
	dialTimeout := time.Duration(cfg.Redis().DialTimeout()) * time.Second

	fmt.Printf("\nAttempting connection to Redis at %s (timeout: %v)...\n", address, dialTimeout)
	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		// A failed connection is expected since we aren't guaranteed to have Redis running locally.
		// The loading and printing succeeded, proving the config pipeline works.
	} else {
		conn.Close()
		fmt.Println("Connection succeeded: Redis is reachable!")
	}
}
