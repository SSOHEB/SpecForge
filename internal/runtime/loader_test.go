package runtime

import (
	"errors"
	"strings"
	"testing"
)

type testConfig struct {
	Instrumentation struct {
		HTTP struct {
			Port           int      `yaml:"port"`
			Enabled        bool     `yaml:"enabled"`
			LogLevel       string   `yaml:"log_level"`
			Host           string   `yaml:"host"`
			Timeout        int      `yaml:"timeout"`
			RedactQuery    []string `yaml:"redact_query"`
			CaptureHeaders []string `yaml:"capture_headers"`
		} `yaml:"http"`
	} `yaml:"instrumentation"`
}

func TestLoadFile_Success(t *testing.T) {
	cfg, err := LoadFile[testConfig]("../../examples/http-server/config.yaml")
	if err != nil {
		t.Fatalf("failed to load valid config: %v", err)
	}

	if cfg.Instrumentation.HTTP.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Instrumentation.HTTP.Port)
	}
	if !cfg.Instrumentation.HTTP.Enabled {
		t.Errorf("expected enabled to be true")
	}
	if cfg.Instrumentation.HTTP.LogLevel != "info" {
		t.Errorf("expected log level 'info', got %s", cfg.Instrumentation.HTTP.LogLevel)
	}
}

func TestLoadReader_Success(t *testing.T) {
	yamlStr := `
instrumentation:
  http:
    port: 9090
    enabled: false
    log_level: "debug"
`
	r := strings.NewReader(yamlStr)
	cfg, err := LoadReader[testConfig](r)
	if err != nil {
		t.Fatalf("failed to load config from reader: %v", err)
	}

	if cfg.Instrumentation.HTTP.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Instrumentation.HTTP.Port)
	}
	if cfg.Instrumentation.HTTP.Enabled {
		t.Errorf("expected enabled to be false")
	}
	if cfg.Instrumentation.HTTP.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got %s", cfg.Instrumentation.HTTP.LogLevel)
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	_, err := LoadFile[testConfig]("nonexistent_file.yaml")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var fnfErr *ConfigFileNotFoundError
	if !errors.As(err, &fnfErr) {
		t.Errorf("expected ConfigFileNotFoundError, got %T: %v", err, err)
	}
}

func TestLoadReader_EmptyFile(t *testing.T) {
	r := strings.NewReader("")
	_, err := LoadReader[testConfig](r)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var emptyErr *ConfigEmptyError
	if !errors.As(err, &emptyErr) {
		t.Errorf("expected ConfigEmptyError, got %T: %v", err, err)
	}
}

func TestLoadReader_MalformedYAML(t *testing.T) {
	r := strings.NewReader(`
instrumentation:
  http:
    port
      enabled: true
`)
	_, err := LoadReader[testConfig](r)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var syntaxErr *ConfigSyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("expected ConfigSyntaxError, got %T: %v", err, err)
	}
}
