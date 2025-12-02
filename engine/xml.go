package engine

import (
	"io"
	"encoding/xml"
)

// XMLEngine renders data as XML
type XMLEngine struct {
	indent string
}

// Render serializes data to XML
func (e *XMLEngine) Render(w io.Writer, name string, data interface{}, ctx interface{}) error {
	encoder := xml.NewEncoder(w)
	if e.indent != "" {
		encoder.Indent("", e.indent)
	}
	// Write XML header
	w.Write([]byte(xml.Header))
	return encoder.Encode(data)
}

// ContentType returns XML content type
func (e *XMLEngine) ContentType() string {
	return "application/xml; charset=utf-8"
}

// NewXMLEngine creates a new XML engine
func NewXMLEngine(indent string) *XMLEngine {
	if indent == "" {
		indent = "  " // Default 2-space indent
	}
	return &XMLEngine{
		indent: indent,
	}
}
