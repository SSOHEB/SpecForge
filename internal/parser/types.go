package parser

// RawMetadata represents the top-level configuration metadata document.
type RawMetadata struct {
	Fields map[string]*RawNode
}

// RawNode represents either a leaf configuration field or a namespace containing child fields.
type RawNode struct {
	// If it is a leaf field:
	Type        *string       `yaml:"type,omitempty"`
	Default     interface{}   `yaml:"default,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Required    *bool         `yaml:"required,omitempty"`
	Min         *float64      `yaml:"min,omitempty"`
	Max         *float64      `yaml:"max,omitempty"`
	Enum        []interface{} `yaml:"enum,omitempty"`
	Pattern     *string       `yaml:"pattern,omitempty"`

	// If it is a namespace/group containing nested fields:
	Children map[string]*RawNode `yaml:"-"`
}
