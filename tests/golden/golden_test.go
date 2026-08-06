package golden

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/SSOHEB/codrao/internal/generator"
	"github.com/SSOHEB/codrao/internal/parser"
	"github.com/SSOHEB/codrao/internal/schema"
)

var update = flag.Bool("update", false, "update golden files")

func TestGoldens(t *testing.T) {
	examples := []string{"http-server", "redis", "postgres"}

	for _, ex := range examples {
		t.Run(ex, func(t *testing.T) {
			metaPath := filepath.Join("../../examples", ex, "metadata.yaml")
			goldenDir := filepath.Join("testdata", ex)

			// 1. Parse metadata
			rawMeta, err := parser.ParseFile(metaPath)
			if err != nil {
				t.Fatalf("failed to parse metadata: %v", err)
			}

			// 2. Build AST
			ast, err := schema.Build(rawMeta)
			if err != nil {
				t.Fatalf("failed to build AST: %v", err)
			}

			// 3. Generate outputs
			schemaBytes, err := generator.GenerateJSONSchema(ast)
			if err != nil {
				t.Fatalf("failed to generate JSON schema: %v", err)
			}

			opts := generator.GenOptions{WithFunctionalAPI: true}
			goBytes, err := generator.GenerateGoCode(ast, "main", opts)
			if err != nil {
				t.Fatalf("failed to generate Go code: %v", err)
			}

			mdBytes, err := generator.GenerateMarkdownDocs(ast)
			if err != nil {
				t.Fatalf("failed to generate Markdown: %v", err)
			}

			// 4. Update golden files if -update is passed
			if *update {
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden directory: %v", err)
				}
				err = os.WriteFile(filepath.Join(goldenDir, "expected_schema.json"), schemaBytes, 0644)
				if err != nil {
					t.Fatalf("failed to write schema golden: %v", err)
				}
				err = os.WriteFile(filepath.Join(goldenDir, "expected_generated_config.go"), goBytes, 0644)
				if err != nil {
					t.Fatalf("failed to write gocode golden: %v", err)
				}
				err = os.WriteFile(filepath.Join(goldenDir, "expected_CONFIGURATION.md"), mdBytes, 0644)
				if err != nil {
					t.Fatalf("failed to write doc golden: %v", err)
				}
				t.Logf("Updated golden files for %s", ex)
				return
			}

			// 5. Compare actual vs golden
			compareGolden(t, filepath.Join(goldenDir, "expected_schema.json"), schemaBytes)
			compareGolden(t, filepath.Join(goldenDir, "expected_generated_config.go"), goBytes)
			compareGolden(t, filepath.Join(goldenDir, "expected_CONFIGURATION.md"), mdBytes)
		})
	}
}

func compareGolden(t *testing.T, goldenPath string, actual []byte) {
	t.Helper()
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v. Run with -update to generate.", goldenPath, err)
	}

	if !bytes.Equal(expected, actual) {
		t.Errorf("mismatch in golden file: %s", goldenPath)
	}
}
