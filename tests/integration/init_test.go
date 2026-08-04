package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_Init_Success(t *testing.T) {
	dir := t.TempDir()
	output, err := runCli("init", "--dir", dir)
	if err != nil {
		t.Fatalf("init failed: %v, output: %s", err, output)
	}

	metaPath := filepath.Join(dir, "metadata.yaml")
	configPath := filepath.Join(dir, "config.yaml")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("metadata.yaml was not created")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.yaml was not created")
	}

	if !strings.Contains(output, "Successfully generated") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify that the generated files pass validate
	valOut, valErr := runCli("validate", "-m", metaPath, "-c", configPath)
	if valErr != nil {
		t.Fatalf("validate failed on generated files: %v, output: %s", valErr, valOut)
	}
	if !strings.Contains(valOut, "config valid") {
		t.Errorf("expected config to be valid, got: %s", valOut)
	}
}

func TestIntegration_Init_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.yaml")
	configPath := filepath.Join(dir, "config.yaml")
	
	// Pre-create both files
	metaContent := []byte("existing metadata")
	configContent := []byte("existing config")
	os.WriteFile(metaPath, metaContent, 0644)
	os.WriteFile(configPath, configContent, 0644)

	output, err := runCli("init", "--dir", dir)
	if err == nil {
		t.Fatalf("expected init to fail due to existing file, but it succeeded")
	}

	if !strings.Contains(output, "metadata.yaml") || !strings.Contains(output, "config.yaml") || !strings.Contains(output, "use --force to overwrite") {
		t.Errorf("expected error mentioning both files and --force, got: %s", output)
	}

	// Verify byte-for-byte equality after failed init
	newMetaContent, _ := os.ReadFile(metaPath)
	if string(newMetaContent) != string(metaContent) {
		t.Errorf("metadata.yaml was modified without --force")
	}

	newConfigContent, _ := os.ReadFile(configPath)
	if string(newConfigContent) != string(configContent) {
		t.Errorf("config.yaml was modified without --force")
	}
}

func TestIntegration_Init_Force(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.yaml")

	// Pre-create metadata
	os.WriteFile(metaPath, []byte("existing metadata"), 0644)

	output, err := runCli("init", "--dir", dir, "--force")
	if err != nil {
		t.Fatalf("init failed: %v, output: %s", err, output)
	}

	content, _ := os.ReadFile(metaPath)
	if string(content) == "existing metadata" {
		t.Errorf("existing file was not overwritten despite --force")
	}
}

func TestIntegration_Init_Dir(t *testing.T) {
	// 1. Set up a pristine temporary "cwd" and a target dir
	cwdDir := t.TempDir()
	targetDir := t.TempDir()

	// 2. We change dir so cwd is clean
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(cwdDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// 3. Run init pointing to targetDir. We can't use runCli easily if runCli relies on relative paths to the binary
	// wait, runCli shells out to `go run ../../cmd/specforge`. If we chdir, that path breaks. 
	// The in-process fallback `cmd.ExecuteWithArgs` does not rely on working directory paths.
	// But it's better to stay in original dir, and just pass an absolute path to --dir, then check our temp cwdDir wasn't touched.
	
	err = os.Chdir(originalDir)
	if err != nil {
		t.Fatalf("failed to chdir back: %v", err)
	}

	// We'll simulate checking cwd by just checking if files were created in originalDir, but that's messy.
	// Actually, just pass targetDir. The test is running in tests/integration anyway.
	output, err := runCli("init", "--dir", targetDir)
	if err != nil {
		t.Fatalf("init failed: %v, output: %s", err, output)
	}

	// Ensure they exist in targetDir
	if _, err := os.Stat(filepath.Join(targetDir, "metadata.yaml")); os.IsNotExist(err) {
		t.Errorf("metadata.yaml not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "config.yaml")); os.IsNotExist(err) {
		t.Errorf("config.yaml not created in target dir")
	}

	// Ensure they do NOT exist in cwd (which is tests/integration)
	if _, err := os.Stat("metadata.yaml"); err == nil {
		t.Errorf("metadata.yaml incorrectly created in cwd")
		os.Remove("metadata.yaml")
	}
	if _, err := os.Stat("config.yaml"); err == nil {
		t.Errorf("config.yaml incorrectly created in cwd")
		os.Remove("config.yaml")
	}
}

