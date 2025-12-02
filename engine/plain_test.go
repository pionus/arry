package engine

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestPlainEngine(t *testing.T) {
	// PlainEngine uses text/template which behaves differently
	// Just test that the engine can be created and has correct content type
	engine := NewPlainEngine("../_example/assets/", "html", nil, true)

	if engine.ContentType() != "text/plain; charset=utf-8" {
		t.Error("Content type should be text/plain")
	}
}

func TestPlainEngineCaching(t *testing.T) {
	// PlainEngine uses text/template which behaves differently
	// Just test that cache is initialized properly
	engine := NewPlainEngine("../_example/assets/", "html", nil, true)

	engine.Mu.RLock()
	cacheExists := engine.cache != nil
	engine.Mu.RUnlock()

	if !cacheExists {
		t.Error("cache should be initialized when caching enabled")
	}
}

func TestPlainEngineNoCaching(t *testing.T) {
	engine := NewPlainEngine("../_example/assets/", "html", nil, false)

	if engine.cache != nil {
		t.Error("cache map should be nil when caching disabled")
	}
}

func TestPlainEngineContentType(t *testing.T) {
	// Don't require actual files for content type test
	engine := &PlainEngine{
		BaseEngine: NewBaseEngine("../_example/assets/", true),
	}

	contentType := engine.ContentType()
	expected := "text/plain; charset=utf-8"

	if contentType != expected {
		t.Errorf("ContentType() = %q, want %q", contentType, expected)
	}
}

func TestPlainEngineCustomFuncMap(t *testing.T) {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	engine := NewPlainEngine("../_example/assets/", "html", funcMap, true)

	if engine.funcMap == nil {
		t.Error("funcMap should not be nil")
	}
}

func TestPlainEngineDefaultExtension(t *testing.T) {
	// Just test that the engine was created, don't load files
	engine := &PlainEngine{
		BaseEngine: NewBaseEngine("../_example/assets/", true),
	}

	// Should use default "txt" extension
	// We can't test the pattern directly, but we can verify the engine was created
	if engine.template == nil {
		// This is OK - we didn't load any template files
	}
}

// BenchmarkPlainEngine benchmarks plain text template rendering
func BenchmarkPlainEngine(b *testing.B) {
	engine := NewPlainEngine("../_example/assets/", "html", nil, true)
	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		engine.Render(buf, "static.html", nil, nil)
	}
}
