package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"configforge/internal/parser"
	"configforge/internal/runtime"
	"configforge/internal/schema"
	"configforge/internal/validator"
)

func main() {
	watch := flag.Bool("watch", false, "watch config file for changes and reload")
	flag.Parse()

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

	if *watch {
		// Start Hot Reload Watcher
		w, err := runtime.NewWatcher[Config](ast, configPath)
		if err != nil {
			fmt.Printf("failed to initialize watcher: %v\n", err)
			os.Exit(1)
		}

		w.OnReload(func(cfg *Config) {
			fmt.Printf("config reloaded: new port (method) = %d, new port (field) = %d\n",
				cfg.Instrumentation().Http().Port(),
				cfg.InstrumentationField.HttpField.PortField)
		})

		w.OnError(func(err error) {
			fmt.Printf("config reload failed: %v\n", err)
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := w.Start(ctx); err != nil {
				fmt.Printf("watcher error: %v\n", err)
			}
		}()

		fmt.Println("watching config file for changes... press Ctrl+C to stop")

		// Block until SIGINT/SIGTERM
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		fmt.Println("\nstopping watcher...")
		_ = w.Stop()
		fmt.Println("stopped")
		os.Exit(0)
	}

	// One-shot behavior
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
	fmt.Println("Method Access:")
	fmt.Printf("  Port: %d\n", cfg.Instrumentation().Http().Port())
	fmt.Printf("  Host: %s\n", cfg.Instrumentation().Http().Host())
	fmt.Printf("  Enabled: %v\n", cfg.Instrumentation().Http().Enabled())
	fmt.Printf("  Timeout: %d\n", cfg.Instrumentation().Http().Timeout())
	fmt.Printf("  LogLevel: %s\n", cfg.Instrumentation().Http().LogLevel())
	fmt.Printf("  ApiKey: %s\n", cfg.Instrumentation().Http().ApiKey())

	fmt.Println("Field Access:")
	fmt.Printf("  Port: %d\n", cfg.InstrumentationField.HttpField.PortField)
	fmt.Printf("  Host: %s\n", cfg.InstrumentationField.HttpField.HostField)
	fmt.Printf("  Enabled: %v\n", cfg.InstrumentationField.HttpField.EnabledField)
	fmt.Printf("  Timeout: %d\n", cfg.InstrumentationField.HttpField.TimeoutField)
	fmt.Printf("  LogLevel: %s\n", cfg.InstrumentationField.HttpField.LogLevelField)
	fmt.Printf("  ApiKey: %s\n", cfg.InstrumentationField.HttpField.ApiKeyField)
}
