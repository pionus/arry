package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONEngine(t *testing.T) {
	engine := NewJSONEngine("  ")

	// Test ContentType
	if ct := engine.ContentType(); ct != "application/json; charset=utf-8" {
		t.Errorf("ContentType = %q, want application/json", ct)
	}

	// Test rendering
	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	buf := new(bytes.Buffer)
	err := engine.Render(buf, "test.json", data, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("name = %v, want John", result["name"])
	}

	if age, ok := result["age"].(float64); !ok || age != 30 {
		t.Errorf("age = %v, want 30", result["age"])
	}
}

func TestJSONEngineIndentation(t *testing.T) {
	// Test with custom indentation
	engine := NewJSONEngine("\t")

	data := map[string]string{"key": "value"}
	buf := new(bytes.Buffer)
	engine.Render(buf, "test.json", data, nil)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("\t")) {
		t.Error("JSON output should contain tab indentation")
	}
}

func TestJSONEngineDefaultIndent(t *testing.T) {
	// Test default indentation (empty string should use default)
	engine := NewJSONEngine("")

	if engine.indent != "  " {
		t.Errorf("default indent = %q, want \"  \"", engine.indent)
	}
}

// BenchmarkJSONEngine benchmarks JSON rendering
func BenchmarkJSONEngine(b *testing.B) {
	engine := NewJSONEngine("  ")

	data := map[string]interface{}{
		"name": "John",
		"age":  30,
		"tags": []string{"go", "web", "api"},
	}

	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		engine.Render(buf, "test.json", data, nil)
	}
}
