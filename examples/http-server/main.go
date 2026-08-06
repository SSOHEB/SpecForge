package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SSOHEB/codrao/pkg/config"
)

func main() {
	watch := flag.Bool("watch", false, "watch config file for changes and reload")
	flag.Parse()

	configPath := os.Getenv("CODRAO_CONFIG_PATH")
	if configPath == "" {
		configPath = "examples/http-server/config.yaml"
		if _, err := os.Stat("config.yaml"); err == nil {
			configPath = "config.yaml"
		}
	}

	metadataFile := "examples/http-server/metadata.yaml"
	if _, err := os.Stat("metadata.yaml"); err == nil {
		metadataFile = "metadata.yaml"
	}

	if *watch {
		stop, err := config.Watch[Config](configPath, func(cfg *Config, err error) {
			if err != nil {
				fmt.Printf("config reload failed: %v\n", err)
				return
			}
			fmt.Printf("config reloaded: new port (method) = %d, new port (field) = %d\n",
				cfg.Instrumentation().Http().Port(),
				cfg.InstrumentationField.HttpField.PortField)
		}, config.WithMetadataPath(metadataFile))
		if err != nil {
			fmt.Printf("failed to initialize watcher: %v\n", err)
			os.Exit(1)
		}
		defer stop()

		fmt.Println("watching config file for changes... press Ctrl+C to stop")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		fmt.Println("\nstopping watcher...")
		os.Exit(0)
	}

	cfg, err := config.Load[Config](configPath, config.WithMetadataPath(metadataFile))
	if err != nil {
		fmt.Printf("Validation or loading failed:\n%v\n", err)
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
