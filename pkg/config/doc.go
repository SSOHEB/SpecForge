// Package config provides a stable, public-facing API for Codrao's configuration management layer.
//
// It wraps the internal parsing, schema compilation, validation, and runtime execution
// (defaults + environment variable overrides) into a single declarative interface.
//
// Quick Start
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//
//		"github.com/SSOHEB/codrao/pkg/config"
//	)
//
//	func main() {
//		// 1. Load config in a single shot (automatically resolves metadata.yaml in the same directory)
//		cfg, err := config.Load[Config]("config.yaml", config.WithEnvPrefix("MYAPP_"))
//		if err != nil {
//			log.Fatalf("failed to load config: %v", err)
//		}
//		fmt.Printf("Loaded config: %+v\n", cfg)
//
//		// 2. Or watch it for hot-reloads
//		stop, err := config.Watch[Config]("config.yaml", func(newCfg *Config, err error) {
//			if err != nil {
//				log.Printf("reload failed: %v", err)
//				return
//			}
//			fmt.Printf("Config reloaded! %+v\n", newCfg)
//		})
//		if err != nil {
//			log.Fatalf("failed to watch config: %v", err)
//		}
//		defer stop()
//	}
package config
