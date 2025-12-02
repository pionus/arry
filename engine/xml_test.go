package engine

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func TestXMLEngine(t *testing.T) {
	engine := NewXMLEngine("  ")

	// Test ContentType
	if ct := engine.ContentType(); ct != "application/xml; charset=utf-8" {
		t.Errorf("ContentType = %q, want application/xml", ct)
	}

	// Test rendering
	type Person struct {
		XMLName xml.Name `xml:"person"`
		Name    string   `xml:"name"`
		Age     int      `xml:"age"`
	}

	data := Person{Name: "John", Age: 30}

	buf := new(bytes.Buffer)
	err := engine.Render(buf, "test.xml", data, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	// Should contain XML header
	if !strings.Contains(output, "<?xml version") {
		t.Error("Missing XML header")
	}

	// Should contain data
	if !strings.Contains(output, "<name>John</name>") {
		t.Error("Missing name element")
	}

	if !strings.Contains(output, "<age>30</age>") {
		t.Error("Missing age element")
	}
}

func TestXMLEngineIndentation(t *testing.T) {
	engine := NewXMLEngine("\t")

	type Simple struct {
		XMLName xml.Name `xml:"root"`
		Value   string   `xml:"value"`
	}

	data := Simple{Value: "test"}
	buf := new(bytes.Buffer)
	engine.Render(buf, "test.xml", data, nil)

	output := buf.String()
	if !strings.Contains(output, "\t") {
		t.Error("XML output should contain tab indentation")
	}
}

func TestXMLEngineDefaultIndent(t *testing.T) {
	engine := NewXMLEngine("")

	if engine.indent != "  " {
		t.Errorf("default indent = %q, want \"  \"", engine.indent)
	}
}

// BenchmarkXMLEngine benchmarks XML rendering
func BenchmarkXMLEngine(b *testing.B) {
	engine := NewXMLEngine("  ")

	type Person struct {
		XMLName xml.Name `xml:"person"`
		Name    string   `xml:"name"`
		Age     int      `xml:"age"`
	}

	data := Person{Name: "John", Age: 30}

	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		engine.Render(buf, "test.xml", data, nil)
	}
}
