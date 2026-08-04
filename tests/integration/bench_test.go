package integration

import (
	"os"
	"testing"

	"github.com/SSOHEB/configforge/internal/generator"
	"github.com/SSOHEB/configforge/internal/parser"
	"github.com/SSOHEB/configforge/internal/runtime"
	"github.com/SSOHEB/configforge/internal/schema"
)

func BenchmarkLoadAndPrepare(b *testing.B) {
	os.Setenv("SPECFORGE_CONFIG_PATH", "../../examples/http-server/config.yaml")
	defer os.Unsetenv("SPECFORGE_CONFIG_PATH")

	rawMeta, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		b.Fatalf("failed to parse metadata: %v", err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		b.Fatalf("failed to build AST: %v", err)
	}

	type httpConfig struct {
		Instrumentation struct {
			Http struct {
				Port           int      `yaml:"port"`
				Enabled        bool     `yaml:"enabled"`
				Host           string   `yaml:"host"`
				Timeout        int      `yaml:"timeout"`
				LogLevel       string   `yaml:"log_level"`
				ApiKey         string   `yaml:"api_key"`
				CaptureHeaders []string `yaml:"capture_headers"`
				RedactQuery    []string `yaml:"redact_query"`
			} `yaml:"http"`
		} `yaml:"instrumentation"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := runtime.LoadAndPrepare[httpConfig](ast)
		if err != nil {
			b.Fatalf("LoadAndPrepare failed: %v", err)
		}
	}
}

func BenchmarkASTBuild(b *testing.B) {
	rawMeta, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		b.Fatalf("failed to parse metadata: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := schema.Build(rawMeta)
		if err != nil {
			b.Fatalf("AST build failed: %v", err)
		}
	}
}

func BenchmarkJSONSchemaGen(b *testing.B) {
	rawMeta, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		b.Fatalf("failed to parse metadata: %v", err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		b.Fatalf("failed to build AST: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateJSONSchema(ast)
		if err != nil {
			b.Fatalf("JSON schema gen failed: %v", err)
		}
	}
}
