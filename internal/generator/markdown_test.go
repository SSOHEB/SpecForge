package generator

import (
	"bytes"
	"strings"
	"testing"

	yamlparser "github.com/SSOHEB/codrao/internal/parser"
	"github.com/SSOHEB/codrao/internal/schema"
)

func TestMarkdownDocsGeneration(t *testing.T) {
	// 1. Full doc generation from examples/http-server/metadata.yaml
	raw, err := yamlparser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata.yaml: %v", err)
	}

	ast, err := schema.Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	mdBytes, err := GenerateMarkdownDocs(ast)
	if err != nil {
		t.Fatalf("failed to generate Markdown: %v", err)
	}

	mdStr := string(mdBytes)

	// Check main headers
	assertContains(t, mdStr, "# Configuration Reference")
	assertContains(t, mdStr, "## Instrumentation")
	assertContains(t, mdStr, "### Http")

	// Verify "port" table contains default (8080) and bounds (1, 65535)
	assertContains(t, mdStr, "| `port` | integer | `8080` | No | Min: 1, Max: 65535 |")

	// Verify "log_level" table contains enum values
	assertContains(t, mdStr, "Enum: [debug, info, warn, error]")

	// 2. Deterministic output verification
	mdBytes2, err := GenerateMarkdownDocs(ast)
	if err != nil {
		t.Fatalf("failed to generate Markdown second time: %v", err)
	}
	if !bytes.Equal(mdBytes, mdBytes2) {
		t.Errorf("Markdown generation is not deterministic (output mismatch)")
	}

	// 3. Nested namespaces produce correctly leveled headings
	assertContains(t, mdStr, "### Http")
	assertContains(t, mdStr, "**Path:** `instrumentation.http`")

	// 4. A field with no description still renders a valid table row
	noDescYAML := `
test:
  no_desc_field:
    type: string
`
	r := strings.NewReader(noDescYAML)
	rawNoDesc, err := yamlparser.Parse(r)
	if err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}
	astNoDesc, err := schema.Build(rawNoDesc)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}
	mdNoDesc, err := GenerateMarkdownDocs(astNoDesc)
	if err != nil {
		t.Fatalf("failed to generate docs: %v", err)
	}
	// Row should end with |- | indicating empty/default description
	assertContains(t, string(mdNoDesc), "| `no_desc_field` | string | - | No | - | - |")

	// 5. Empty AST produces a minimal valid doc
	mdEmpty, err := GenerateMarkdownDocs(nil)
	if err != nil {
		t.Fatalf("failed on nil AST: %v", err)
	}
	if string(mdEmpty) != "# Configuration Reference\n" {
		t.Errorf("expected clean empty doc, got %q", string(mdEmpty))
	}
}
