package schema

// FieldType represents the normalized type of a configuration field.
type FieldType int

const (
	TypeBool FieldType = iota
	TypeString
	TypeInt
	TypeFloat
	TypeStringSlice
	TypeIntSlice
	TypeStringMap
)

func (t FieldType) String() string {
	switch t {
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeStringSlice:
		return "string[]"
	case TypeIntSlice:
		return "int[]"
	case TypeStringMap:
		return "map[string]string"
	default:
		return "unknown"
	}
}

// Field represents a leaf configuration field.
type Field struct {
	Name        string
	YAMLKey     string
	Path        []string
	Type        FieldType
	Default     any
	Description string
	Required    bool
	Min         *float64
	Max         *float64
	Enum        []string
	Pattern     string
}

// Node represents a configuration namespace.
type Node struct {
	Name     string
	YAMLKey  string
	Path     []string
	Fields   []*Field
	Children []*Node
}

// AST represents the entire normalized AST structure.
type AST struct {
	Root *Node
}
