package main

import (
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/SSOHEB/codrao/pkg/config"
)

func TestPostgresConfigLoadAndValidate(t *testing.T) {
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err != nil {
		t.Skipf("Skipping test: %s not found. Run from examples/postgres directory.", configPath)
	}
	
	metadataFile := "metadata.yaml"
	if _, err := os.Stat(metadataFile); err != nil {
		t.Skipf("Skipping test: %s not found.", metadataFile)
	}

	cfg, err := config.Load[Config](configPath, config.WithMetadataPath(metadataFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	t.Logf("config valid")
	t.Logf("Resolved Postgres Configuration (Functional API):")
	t.Logf("  Host: %s", cfg.Postgres().Host())
	t.Logf("  Port: %d", cfg.Postgres().Port())
	t.Logf("  Database: %s", cfg.Postgres().Database())
	t.Logf("  User: %s", cfg.Postgres().User())
	t.Logf("  SSL Mode: %s", cfg.Postgres().SslMode())
	t.Logf("  Max Open Conns: %d", cfg.Postgres().MaxOpenConns())
	t.Logf("  Max Idle Conns: %d", cfg.Postgres().MaxIdleConns())
	t.Logf("  Max Lifetime: %d", cfg.Postgres().ConnMaxLifetime())
	t.Logf("  Retry Attempts: %d", cfg.Postgres().RetryAttempts())
	t.Logf("  Retry Backoff: %d ms", cfg.Postgres().RetryBackoffMs())

	// Reconnection retry check: We wrap a TCP connection attempt to Postgres in a loop
	attempts := cfg.Postgres().RetryAttempts()
	backoff := time.Duration(cfg.Postgres().RetryBackoffMs()) * time.Millisecond
	address := net.JoinHostPort(cfg.Postgres().Host(), strconv.Itoa(cfg.Postgres().Port()))
	dialTimeout := 200 * time.Millisecond // short timeout for test speed

	t.Logf("Attempting connection to Postgres at %s...", address)
	for attempt := 0; attempt <= attempts; attempt++ {
		conn, err := net.DialTimeout("tcp", address, dialTimeout)
		if err == nil {
			conn.Close()
			t.Logf("Connection succeeded: Postgres is reachable!")
			return
		}

		t.Logf("Attempt %d/%d failed: %v", attempt+1, attempts+1, err)
		if attempt < attempts {
			t.Logf("Retrying in %v...", backoff)
			time.Sleep(backoff)
		}
	}

	t.Logf("All connection attempts failed: Postgres is unreachable.")
}
