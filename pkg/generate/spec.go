package generate

// Document is a simplified OpenAPI 3.x document.
type Document struct {
	Info    DocInfo
	Schemas []NamedSchema
	Paths   []PathEntry
}

// DocInfo holds document metadata.
type DocInfo struct {
	Title   string
	Version string
}

// NamedSchema is a named component schema.
type NamedSchema struct {
	Name   string
	Schema Schema
}

// Schema is a simplified OpenAPI schema.
type Schema struct {
	Types      []string
	Format     string
	Properties []NamedProperty
	Required   map[string]bool
	Items      *SchemaRef
	ReadOnly   bool
}

// HasType returns true if the schema includes the given type.
func (s Schema) HasType(t string) bool {
	for _, tt := range s.Types {
		if tt == t {
			return true
		}
	}
	return false
}

// NamedProperty is a named property within a schema.
type NamedProperty struct {
	Name   string
	Ref    string
	Schema Schema
}

// SchemaRef is a schema that may be a $ref.
type SchemaRef struct {
	Ref    string
	Schema *Schema
}

// PathEntry is a path with its operations.
type PathEntry struct {
	Path       string
	Operations []Operation
}

// Operation is an HTTP operation.
type Operation struct {
	Method      string
	OperationID string
	Tags        []string
	Parameters  []Parameter
	RequestBody *SchemaRef
	Responses   map[string]*SchemaRef
}

// Parameter is a path/query/header parameter.
type Parameter struct {
	Name     string
	In       string
	Required bool
	Schema   *Schema
}
