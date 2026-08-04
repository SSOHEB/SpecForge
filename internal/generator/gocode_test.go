package generator

import (
	"bytes"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	yamlparser "github.com/SSOHEB/SpecForge/internal/parser"
	"github.com/SSOHEB/SpecForge/internal/schema"
	"go/ast"
)

func TestGoCodeGeneration(t *testing.T) {
	// 1. Generation from examples/http-server/metadata.yaml compiles
	raw, err := yamlparser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata.yaml: %v", err)
	}

	ast, err := schema.Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	goCodeBytes, err := GenerateGoCode(ast, "config")
	if err != nil {
		t.Fatalf("failed to generate Go code: %v", err)
	}

	// Verify using go/parser.ParseFile
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "generated_config.go", goCodeBytes, parser.ParseComments)
	if err != nil {
		t.Fatalf("generated Go code failed compilation/parsing check: %v\nCode:\n%s", err, string(goCodeBytes))
	}

	if parsedFile.Name.Name != "config" {
		t.Errorf("expected package name 'config', got %s", parsedFile.Name.Name)
	}

	// 2. Deterministic output verification (generate twice, byte-identical)
	goCodeBytes2, err := GenerateGoCode(ast, "config")
	if err != nil {
		t.Fatalf("failed to generate Go code second time: %v", err)
	}
	if !bytes.Equal(goCodeBytes, goCodeBytes2) {
		t.Errorf("Go code generation is not deterministic (output mismatch)")
	}

	// 3. Verify all 7 types, description doc comments, and yaml tags
	allTypesYAML := `
all_types:
  boolean_field:
    type: bool
    description: "A boolean field"
  string_field:
    type: string
    description: "A string field"
  integer_field:
    type: int
    description: "An integer field"
  float_field:
    type: float
    description: "A float field"
  string_slice_field:
    type: string[]
    description: "A string slice field"
  int_slice_field:
    type: int[]
    description: "An int slice field"
  string_map_field:
    type: map[string]string
    description: "A string map field"
`
	r := strings.NewReader(allTypesYAML)
	rawAll, err := yamlparser.Parse(r)
	if err != nil {
		t.Fatalf("failed to parse all-types yaml: %v", err)
	}
	astAll, err := schema.Build(rawAll)
	if err != nil {
		t.Fatalf("failed to build all-types AST: %v", err)
	}
	allTypesCodeBytes, err := GenerateGoCode(astAll, "config")
	if err != nil {
		t.Fatalf("failed to generate Go code for all types: %v", err)
	}

	codeStr := string(allTypesCodeBytes)

	// Assert correct types
	assertContains(t, codeStr, "BooleanField bool `yaml:\"boolean_field\"`")
	assertContains(t, codeStr, "StringField string `yaml:\"string_field\"`")
	assertContains(t, codeStr, "IntegerField int `yaml:\"integer_field\"`")
	assertContains(t, codeStr, "FloatField float64 `yaml:\"float_field\"`")
	assertContains(t, codeStr, "StringSliceField []string `yaml:\"string_slice_field\"`")
	assertContains(t, codeStr, "IntSliceField []int `yaml:\"int_slice_field\"`")
	assertContains(t, codeStr, "StringMapField map[string]string `yaml:\"string_map_field\"`")

	// Assert descriptions become comments
	assertContains(t, codeStr, "// A boolean field")
	assertContains(t, codeStr, "// A string field")
	assertContains(t, codeStr, "// An integer field")
	assertContains(t, codeStr, "// A float field")
	assertContains(t, codeStr, "// A string slice field")
	assertContains(t, codeStr, "// An int slice field")
	assertContains(t, codeStr, "// A string map field")

	// Assert nested namespace type formatting
	nestedYAML := `
parent:
  child:
    type: string
`
	rNested := strings.NewReader(nestedYAML)
	rawNested, err := yamlparser.Parse(rNested)
	if err != nil {
		t.Fatalf("failed to parse nested yaml: %v", err)
	}
	astNested, err := schema.Build(rawNested)
	if err != nil {
		t.Fatalf("failed to build nested AST: %v", err)
	}
	nestedCodeBytes, err := GenerateGoCode(astNested, "config")
	if err != nil {
		t.Fatalf("failed to generate Go code for nested namespaces: %v", err)
	}

	nestedCodeStr := string(nestedCodeBytes)

	// Suffix "Config" only on root-level namespaces (parent becomes ParentConfig)
	// Nested ones just use the Go name (child is inside parent, but it is a leaf, so it's a field)
	assertContains(t, nestedCodeStr, "Parent ParentConfig `yaml:\"parent\"`")
	assertContains(t, nestedCodeStr, "type ParentConfig struct {")

	// What if child was a namespace? Let's verify that also works.
	deepNestedYAML := `
parent:
  child_ns:
    leaf_field:
      type: string
`
	rDeep := strings.NewReader(deepNestedYAML)
	rawDeep, err := yamlparser.Parse(rDeep)
	if err != nil {
		t.Fatalf("failed to parse deep nested yaml: %v", err)
	}
	astDeep, err := schema.Build(rawDeep)
	if err != nil {
		t.Fatalf("failed to build deep nested AST: %v", err)
	}
	deepCodeBytes, err := GenerateGoCode(astDeep, "config")
	if err != nil {
		t.Fatalf("failed to generate Go code for deep nested namespaces: %v", err)
	}

	deepCodeStr := string(deepCodeBytes)
	assertContains(t, deepCodeStr, "Parent ParentConfig `yaml:\"parent\"`")
	assertContains(t, deepCodeStr, "type ParentConfig struct {\n\tChildNs ChildNs `yaml:\"child_ns\"`")
	assertContains(t, deepCodeStr, "type ChildNs struct {\n\tLeafField string `yaml:\"leaf_field\"`")
}

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected generated code to contain %q, but it did not.\nFull code:\n%s", sub, s)
	}
}

func TestGoCodeGeneration_FunctionalAPI(t *testing.T) {
	raw, err := yamlparser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	astStruct, err := schema.Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	// 1. Verify WithFunctionalAPI: false produces output identical to default
	codeDefault, err := GenerateGoCode(astStruct, "config")
	if err != nil {
		t.Fatalf("failed to generate default: %v", err)
	}
	codeFalse, err := GenerateGoCode(astStruct, "config", GenOptions{WithFunctionalAPI: false})
	if err != nil {
		t.Fatalf("failed to generate WithFunctionalAPI: false: %v", err)
	}
	if !bytes.Equal(codeDefault, codeFalse) {
		t.Errorf("default generation differs from WithFunctionalAPI: false")
	}

	// 2. Generate with WithFunctionalAPI: true
	codeTrue, err := GenerateGoCode(astStruct, "config", GenOptions{WithFunctionalAPI: true})
	if err != nil {
		t.Fatalf("failed to generate WithFunctionalAPI: true: %v", err)
	}

	// Verify using go/parser.ParseFile
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "generated_config.go", codeTrue, parser.ParseComments)
	if err != nil {
		t.Fatalf("generated functional Go code failed parsing check: %v\nCode:\n%s", err, string(codeTrue))
	}

	var foundConfigStruct, foundHttpStruct bool
	var foundConfigInstrumentationMethod, foundHttpPortMethod bool

	for _, decl := range parsedFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if typeSpec.Name.Name == "Config" {
					foundConfigStruct = true
					var foundField bool
					for _, field := range structType.Fields.List {
						for _, name := range field.Names {
							if name.Name == "InstrumentationField" {
								foundField = true
							}
						}
					}
					if !foundField {
						t.Errorf("Config struct missing InstrumentationField")
					}
				}

				if typeSpec.Name.Name == "Http" {
					foundHttpStruct = true
					var foundField bool
					for _, field := range structType.Fields.List {
						for _, name := range field.Names {
							if name.Name == "PortField" {
								foundField = true
							}
						}
					}
					if !foundField {
						t.Errorf("Http struct missing PortField")
					}
				}
			}

		case *ast.FuncDecl:
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recvType := d.Recv.List[0].Type
				var recvName string
				if star, ok := recvType.(*ast.StarExpr); ok {
					if ident, ok := star.X.(*ast.Ident); ok {
						recvName = ident.Name
					}
				} else if ident, ok := recvType.(*ast.Ident); ok {
					recvName = ident.Name
				}

				if recvName == "Config" && d.Name.Name == "Instrumentation" {
					foundConfigInstrumentationMethod = true
					if star, ok := d.Type.Results.List[0].Type.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok {
							if ident.Name != "InstrumentationConfig" {
								t.Errorf("expected Config.Instrumentation to return *InstrumentationConfig, got *%s", ident.Name)
							}
						}
					} else {
						t.Errorf("expected Config.Instrumentation to return a pointer type")
					}
				}

				if recvName == "Http" && d.Name.Name == "Port" {
					foundHttpPortMethod = true
					if ident, ok := d.Type.Results.List[0].Type.(*ast.Ident); ok {
						if ident.Name != "int" {
							t.Errorf("expected Http.Port to return int, got %s", ident.Name)
						}
					} else {
						t.Errorf("expected Http.Port to return basic identifier type")
					}
				}
			}
		}
	}

	if !foundConfigStruct {
		t.Errorf("missing Config struct definition")
	}
	if !foundHttpStruct {
		t.Errorf("missing Http struct definition")
	}
	if !foundConfigInstrumentationMethod {
		t.Errorf("missing Config.Instrumentation() method definition")
	}
	if !foundHttpPortMethod {
		t.Errorf("missing Http.Port() method definition")
	}
}
