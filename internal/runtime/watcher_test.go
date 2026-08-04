package runtime

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SSOHEB/SpecForge/internal/schema"
)

type testWatcherConfig struct {
	Port int `yaml:"port"`
}

func buildTestWatcherAST() *schema.AST {
	ast := &schema.AST{
		Root: &schema.Node{
			Name: "Config",
			Path: []string{},
			Fields: []*schema.Field{
				{Name: "Port", YAMLKey: "port", Path: []string{"port"}, Type: schema.TypeInt, Default: 80},
			},
		},
	}
	return ast
}

func TestWatcher_Success(t *testing.T) {
	ast := buildTestWatcherAST()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write initial config
	if err := os.WriteFile(configPath, []byte("port: 1000\n"), 0644); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	w, err := NewWatcher[testWatcherConfig](ast, configPath, RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if w.Current().Port != 1000 {
		t.Errorf("expected initial Port 1000, got %d", w.Current().Port)
	}

	reloadChan := make(chan int, 1)
	w.OnReload(func(cfg *testWatcherConfig) {
		reloadChan <- cfg.Port
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Start(ctx); err != nil {
			t.Errorf("watcher failed: %v", err)
		}
	}()

	// Wait for watcher to initialize directory add
	time.Sleep(50 * time.Millisecond)

	// Modify config
	if err := os.WriteFile(configPath, []byte("port: 2000\n"), 0644); err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	// Wait for reload
	select {
	case newPort := <-reloadChan:
		if newPort != 2000 {
			t.Errorf("expected reloaded port 2000, got %d", newPort)
		}
		if w.Current().Port != 2000 {
			t.Errorf("expected w.Current().Port 2000, got %d", w.Current().Port)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for reload")
	}
}

func TestWatcher_FailFast(t *testing.T) {
	ast := buildTestWatcherAST()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write invalid initial config
	if err := os.WriteFile(configPath, []byte("port: [invalid]\n"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	_, err := NewWatcher[testWatcherConfig](ast, configPath, RuntimeOptions{})
	if err == nil {
		t.Fatalf("expected NewWatcher to fail fast on invalid YAML, but it succeeded")
	}
}

func TestWatcher_InvalidReload(t *testing.T) {
	ast := buildTestWatcherAST()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("port: 1000\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	w, err := NewWatcher[testWatcherConfig](ast, configPath, RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	errChan := make(chan error, 1)
	w.OnError(func(err error) {
		errChan <- err
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Write invalid yaml
	if err := os.WriteFile(configPath, []byte("port: [invalid]\n"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	select {
	case err := <-errChan:
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		// Confirm last valid config is still returned
		if w.Current().Port != 1000 {
			t.Errorf("expected Current() to return last valid port 1000, got %d", w.Current().Port)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for error notification")
	}
}

func TestWatcher_Debounce(t *testing.T) {
	ast := buildTestWatcherAST()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("port: 1000\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	w, err := NewWatcher[testWatcherConfig](ast, configPath, RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	var reloadCount int32
	reloadChan := make(chan int, 1)
	w.OnReload(func(cfg *testWatcherConfig) {
		atomic.AddInt32(&reloadCount, 1)
		reloadChan <- cfg.Port
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Write three modifications in rapid succession
	for i := 1; i <= 3; i++ {
		_ = os.WriteFile(configPath, []byte(string([]byte{byte(48 + i)})), 0644) // writes: port values will reload to default since parsing single digit fails or fails YAML shape, wait: let's write valid port strings to avoid validation error
		_ = os.WriteFile(configPath, []byte(string([]byte("port: "))+string([]byte{byte(48 + i)})+string([]byte("\n"))), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for reload
	time.Sleep(200 * time.Millisecond)

	count := atomic.LoadInt32(&reloadCount)
	if count != 1 {
		t.Errorf("expected exactly 1 reload triggered due to debounce, got %d", count)
	}
}

func TestWatcher_Stop(t *testing.T) {
	ast := buildTestWatcherAST()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("port: 1000\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	w, err := NewWatcher[testWatcherConfig](ast, configPath, RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	var reloadCount int32
	w.OnReload(func(cfg *testWatcherConfig) {
		atomic.AddInt32(&reloadCount, 1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Stop watcher
	if err := w.Stop(); err != nil {
		t.Fatalf("failed to stop watcher: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Modify config again
	if err := os.WriteFile(configPath, []byte("port: 2000\n"), 0644); err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	count := atomic.LoadInt32(&reloadCount)
	if count != 0 {
		t.Errorf("expected 0 reloads after Stop(), got %d", count)
	}
}
