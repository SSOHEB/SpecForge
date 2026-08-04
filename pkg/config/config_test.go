package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SSOHEB/SpecForge/pkg/config"
)

type testConfig struct {
	Instrumentation struct {
		Http struct {
			Port int `yaml:"port"`
		} `yaml:"http"`
	} `yaml:"instrumentation"`
}

func setupTestFiles(t *testing.T) (dir, metaPath, configPath string) {
	dir = t.TempDir()
	metaPath = filepath.Join(dir, "metadata.yaml")
	configPath = filepath.Join(dir, "config.yaml")

	metaYAML := `
instrumentation:
  http:
    port:
      type: int
      min: 1024
`
	if err := os.WriteFile(metaPath, []byte(metaYAML), 0644); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	configYAML := `
instrumentation:
  http:
    port: 8080
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return dir, metaPath, configPath
}

func TestLoad_Success(t *testing.T) {
	_, _, configPath := setupTestFiles(t)

	cfg, err := config.Load[testConfig](configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Instrumentation.Http.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Instrumentation.Http.Port)
	}
}

func TestLoad_ValidationError(t *testing.T) {
	_, _, configPath := setupTestFiles(t)

	// Write invalid config (port below minimum)
	if err := os.WriteFile(configPath, []byte("instrumentation:\n  http:\n    port: 80\n"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	_, err := config.Load[testConfig](configPath)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErrs config.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Fatalf("expected error to be ValidationErrors, got %T: %v", err, err)
	}

	if len(valErrs) == 0 {
		t.Fatalf("expected at least one validation error")
	}

	if valErrs[0].Rule != "min" {
		t.Errorf("expected rule 'min', got %q", valErrs[0].Rule)
	}
}

func TestLoad_MissingMetadataPath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := config.Load[testConfig](configPath)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := "config: no metadata path provided and no metadata.yaml found alongside"
	if err.Error()[:len(expectedErr)] != expectedErr {
		t.Errorf("expected error to start with %q, got %q", expectedErr, err.Error())
	}
}

func TestWatch_ReloadsOnChange(t *testing.T) {
	_, _, configPath := setupTestFiles(t)

	ch := make(chan *testConfig, 1)
	errCh := make(chan error, 1)

	stop, err := config.Watch[testConfig](configPath, func(cfg *testConfig, err error) {
		if err != nil {
			errCh <- err
			return
		}
		ch <- cfg
	})
	if err != nil {
		t.Fatalf("failed to start watch: %v", err)
	}
	defer stop()

	// Modify config
	newConfigYAML := `
instrumentation:
  http:
    port: 9090
`
	if err := os.WriteFile(configPath, []byte(newConfigYAML), 0644); err != nil {
		t.Fatalf("failed to modify config: %v", err)
	}

	select {
	case cfg := <-ch:
		if cfg.Instrumentation.Http.Port != 9090 {
			t.Errorf("expected port 9090, got %d", cfg.Instrumentation.Http.Port)
		}
	case err := <-errCh:
		t.Fatalf("unexpected watch error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for reload")
	}
}

func TestWatch_StopStopsWatching(t *testing.T) {
	_, _, configPath := setupTestFiles(t)

	ch := make(chan *testConfig, 1)
	stop, err := config.Watch[testConfig](configPath, func(cfg *testConfig, err error) {
		ch <- cfg
	})
	if err != nil {
		t.Fatalf("failed to start watch: %v", err)
	}
	
	stop()
	
	// Wait a moment for watcher to actually stop
	time.Sleep(100 * time.Millisecond)

	// Modify config
	if err := os.WriteFile(configPath, []byte("instrumentation:\n  http:\n    port: 9999\n"), 0644); err != nil {
		t.Fatalf("failed to modify config: %v", err)
	}

	select {
	case <-ch:
		t.Fatalf("received reload event after stop")
	case <-time.After(500 * time.Millisecond):
		// Success
	}
}

func TestValidate_Success(t *testing.T) {
	_, metaPath, configPath := setupTestFiles(t)

	err := config.Validate(configPath, metaPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_Failure(t *testing.T) {
	_, metaPath, configPath := setupTestFiles(t)

	if err := os.WriteFile(configPath, []byte("instrumentation:\n  http:\n    port: 80\n"), 0644); err != nil {
		t.Fatalf("failed to modify config: %v", err)
	}

	err := config.Validate(configPath, metaPath)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErrs config.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Fatalf("expected error to be ValidationErrors, got %T: %v", err, err)
	}
}

func TestWithEnvPrefix(t *testing.T) {
	_, _, configPath := setupTestFiles(t)
	
	os.Setenv("TESTPREFIX_INSTRUMENTATION_HTTP_PORT", "7777")
	defer os.Unsetenv("TESTPREFIX_INSTRUMENTATION_HTTP_PORT")

	cfg, err := config.Load[testConfig](configPath, config.WithEnvPrefix("TESTPREFIX_"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Instrumentation.Http.Port != 7777 {
		t.Errorf("expected env override port 7777, got %d", cfg.Instrumentation.Http.Port)
	}
}

func TestWithoutEnvOverrides(t *testing.T) {
	_, _, configPath := setupTestFiles(t)
	
	os.Setenv("SPECFORGE_INSTRUMENTATION_HTTP_PORT", "7777")
	defer os.Unsetenv("SPECFORGE_INSTRUMENTATION_HTTP_PORT")

	cfg, err := config.Load[testConfig](configPath, config.WithoutEnvOverrides())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Instrumentation.Http.Port != 8080 {
		t.Errorf("expected original port 8080 because env overrides disabled, got %d", cfg.Instrumentation.Http.Port)
	}
}

func TestWithMetadataPath(t *testing.T) {
	_, metaPath, _ := setupTestFiles(t)
	
	// Config file is in another dir
	dir2 := t.TempDir()
	configPath := filepath.Join(dir2, "config.yaml")
	configYAML := `
instrumentation:
  http:
    port: 8080
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// This should succeed because we explicitly passed the metadata path from dir1
	_, err := config.Load[testConfig](configPath, config.WithMetadataPath(metaPath))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
