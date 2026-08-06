// Package cmd contains unit tests for verifying the Cobra CLI commands.
package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	// Reset persistent flags to clear state between tests
	metadataPath = "./metadata.yaml"
	configPath = "./config.yaml"
	outputPath = "./generated"
	writeSchema = false
	functionalAPI = false

	_, err := rootCmd.ExecuteC()
	return buf.String(), err
}

func TestCli_Validate_Success(t *testing.T) {
	output, err := executeCommand("validate", "-m", "../../../examples/http-server/metadata.yaml", "-c", "../../../examples/http-server/config.yaml")
	if err != nil {
		t.Fatalf("expected validate success, got error: %v, output: %s", err, output)
	}
	if !strings.Contains(output, "config valid") {
		t.Errorf("expected output to contain 'config valid', got: %s", output)
	}
}

func TestCli_Validate_Failure(t *testing.T) {
	// Write a temp config with invalid port (0 is less than min 1)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "invalid_config.yaml")
	invalidConfig := `
instrumentation:
  http:
    port: 0
    api_key: "VALIDAPIKEY12345"
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := executeCommand("validate", "-m", "../../../examples/http-server/metadata.yaml", "-c", configPath)
	if err == nil {
		t.Fatal("expected validate to fail, but it succeeded")
	}

	var uErr *CliUserError
	if !errors.As(err, &uErr) {
		t.Errorf("expected CliUserError, got %T: %v", err, err)
	}
	if !strings.Contains(uErr.Msg, "validation failed") {
		t.Errorf("expected validation failed message, got: %s", uErr.Msg)
	}
}

func TestCli_Validate_MissingFiles(t *testing.T) {
	// Missing metadata
	_, err := executeCommand("validate", "-m", "non_existent_meta.yaml")
	if err == nil {
		t.Fatal("expected error for missing metadata, got nil")
	}
	var uErr *CliUserError
	if !errors.As(err, &uErr) || !strings.Contains(uErr.Msg, "metadata spec file not found") {
		t.Errorf("expected metadata file not found error, got: %v", err)
	}

	// Missing config
	_, err = executeCommand("validate", "-m", "../../../examples/http-server/metadata.yaml", "-c", "non_existent_config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
	if !errors.As(err, &uErr) || !strings.Contains(uErr.Msg, "configuration file not found") {
		t.Errorf("expected config file not found error, got: %v", err)
	}
}

func TestCli_Generate(t *testing.T) {
	dir := t.TempDir()
	output, err := executeCommand("generate", "-m", "../../../examples/http-server/metadata.yaml", "-o", dir)
	if err != nil {
		t.Fatalf("expected generate success, got: %v, output: %s", err, output)
	}

	schemaPath := filepath.Join(dir, "schema.json")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Errorf("expected schema.json to be created, but it was not")
	}

	goPath := filepath.Join(dir, "generated_config.go")
	if _, err := os.Stat(goPath); os.IsNotExist(err) {
		t.Errorf("expected generated_config.go to be created, but it was not")
	}
}

func TestCli_Schema(t *testing.T) {
	output, err := executeCommand("schema", "-m", "../../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("expected schema success, got: %v", err)
	}

	if !strings.Contains(output, "http://json-schema.org/draft-07/schema#") {
		t.Errorf("expected output to contain json schema header, got: %s", output)
	}
}

func TestCli_Defaults(t *testing.T) {
	output, err := executeCommand("defaults", "-m", "../../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("expected defaults success, got: %v", err)
	}

	expectedPaths := []string{
		"instrumentation.http.enabled",
		"instrumentation.http.host",
		"instrumentation.http.port",
		"instrumentation.http.api_key",
	}

	for _, path := range expectedPaths {
		if !strings.Contains(output, path) {
			t.Errorf("expected defaults output to contain path %q, got: %s", path, output)
		}
	}
}

func TestCli_Docs_Success(t *testing.T) {
	dir := t.TempDir()
	output, err := executeCommand("docs", "-m", "../../../examples/http-server/metadata.yaml", "-o", dir)
	if err != nil {
		t.Fatalf("expected docs success, got: %v, output: %s", err, output)
	}

	docPath := filepath.Join(dir, "CONFIGURATION.md")
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		t.Errorf("expected CONFIGURATION.md to be created")
	}
}

func TestCli_Version(t *testing.T) {
	output, err := executeCommand("version")
	if err != nil {
		t.Fatalf("expected version success, got: %v", err)
	}
	if !strings.Contains(output, "codrao version:") {
		t.Errorf("expected version output, got: %s", output)
	}
}
