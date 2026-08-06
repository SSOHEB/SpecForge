package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SSOHEB/codrao/cmd/codrao/cmd"
)

func runCli(args ...string) (string, error) {
	// 1. Try to shell out to "go run"
	cmdArgs := append([]string{"run", "../../cmd/codrao"}, args...)
	execCmd := exec.Command("go", cmdArgs...)
	outBytes, err := execCmd.CombinedOutput()
	if err == nil {
		return string(outBytes), nil
	}

	// 2. If blocked by Windows Defender Application Control policy, fall back to in-process execution.
	outStr := string(outBytes)
	if strings.Contains(outStr, "Application Control") || strings.Contains(err.Error(), "Access is denied") || strings.Contains(outStr, "blocked") {
		return cmd.ExecuteWithArgs(args)
	}

	return outStr, err
}

func TestIntegration_Generate(t *testing.T) {
	examples := []string{"http-server", "redis", "postgres"}

	for _, ex := range examples {
		t.Run(ex, func(t *testing.T) {
			dir := t.TempDir()
			metaPath := filepath.Join("../../examples", ex, "metadata.yaml")

			output, err := runCli("generate", "-m", metaPath, "-o", dir, "--functional-api")
			if err != nil {
				t.Fatalf("generate failed: %v, output: %s", err, output)
			}

			schemaPath := filepath.Join(dir, "schema.json")
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				t.Errorf("schema.json was not created")
			}

			goPath := filepath.Join(dir, "generated_config.go")
			if _, err := os.Stat(goPath); os.IsNotExist(err) {
				t.Errorf("generated_config.go was not created")
			}
		})
	}
}

func TestIntegration_Validate_Success(t *testing.T) {
	examples := []string{"http-server", "redis", "postgres"}

	for _, ex := range examples {
		t.Run(ex, func(t *testing.T) {
			metaPath := filepath.Join("../../examples", ex, "metadata.yaml")
			configPath := filepath.Join("../../examples", ex, "config.yaml")

			output, err := runCli("validate", "-m", metaPath, "-c", configPath)
			if err != nil {
				t.Fatalf("validate failed: %v, output: %s", err, output)
			}

			if !strings.Contains(output, "config valid") {
				t.Errorf("expected output to contain 'config valid', got: %s", output)
			}
		})
	}
}

func TestIntegration_Validate_Failure(t *testing.T) {
	tests := []struct {
		example       string
		invalidConfig string
		expectedError string
	}{
		{
			example: "http-server",
			invalidConfig: `
instrumentation:
  http:
    port: 0
    api_key: "VALIDAPIKEY12345"
`,
			expectedError: "instrumentation.http.port: value 0 is less than minimum 1",
		},
		{
			example: "redis",
			invalidConfig: `
redis:
  port: 70000
`,
			expectedError: "redis.port: value 70000 is greater than maximum 65535",
		},
		{
			example: "postgres",
			invalidConfig: `
postgres:
  host: "localhost"
`,
			expectedError: "validation failed", // validation error outputs will print
		},
	}

	for _, tc := range tests {
		t.Run(tc.example, func(t *testing.T) {
			dir := t.TempDir()
			configPath := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tc.invalidConfig), 0644); err != nil {
				t.Fatalf("failed to write invalid config: %v", err)
			}

			metaPath := filepath.Join("../../examples", tc.example, "metadata.yaml")

			output, err := runCli("validate", "-m", metaPath, "-c", configPath)
			if err == nil {
				t.Fatal("expected validate to fail, but it succeeded")
			}

			// Verify we got an error — either exec.ExitError (shell) or CliUserError (in-process fallback)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 1 {
					t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
				}
			}
			// If it's not an ExitError, it came from the in-process fallback — any non-nil error is acceptable.

			if !strings.Contains(output, tc.expectedError) {
				t.Errorf("expected output to contain %q, got: %s", tc.expectedError, output)
			}
		})
	}
}
