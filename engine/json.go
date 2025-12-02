package engine

import (
	"io"
	"encoding/json"
)

// JSONEngine renders data as JSON
type JSONEngine struct {
	indent string
}

// Render serializes data to JSON
func (e *JSONEngine) Render(w io.Writer, name string, data interface{}, ctx interface{}) error {
	encoder := json.NewEncoder(w)
	if e.indent != "" {
		encoder.SetIndent("", e.indent)
	}
	return encoder.Encode(data)
}

// ContentType returns JSON content type
func (e *JSONEngine) ContentType() string {
	return "application/json; charset=utf-8"
}

// NewJSONEngine creates a new JSON engine
func NewJSONEngine(indent string) *JSONEngine {
	if indent == "" {
		indent = "  " // Default 2-space indent
	}
	return &JSONEngine{
		indent: indent,
	}
}
