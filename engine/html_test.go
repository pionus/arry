package engine

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestHTMLEngine(t *testing.T) {
	file, _ := os.ReadFile("../_example/assets/static.html")

	engine := NewHTMLEngine("../_example/assets/", "html", nil, true)

	buf := new(bytes.Buffer)
	engine.Render(buf, "static.html", nil, nil)

	if buf.String() != string(file) {
		t.Errorf("engine rendering is not correct, %s", buf.String())
	}
}

// TestHTMLEngineCaching tests that templates are cached correctly
func TestHTMLEngineCaching(t *testing.T) {
	// Create engine with caching enabled
	engine := NewHTMLEngine("../_example/assets/", "html", nil, true)

	// First render should cache the template
	buf1 := new(bytes.Buffer)
	err := engine.Render(buf1, "static.html", nil, nil)
	if err != nil {
		t.Errorf("first render failed: %v", err)
	}

	// Check cache
	engine.Mu.RLock()
	_, cached := engine.cache["static.html"]
	engine.Mu.RUnlock()

	if !cached {
		t.Error("template was not cached after first render")
	}

	// Second render should use cache
	buf2 := new(bytes.Buffer)
	err = engine.Render(buf2, "static.html", nil, nil)
	if err != nil {
		t.Errorf("second render failed: %v", err)
	}

	// Results should be identical
	if buf1.String() != buf2.String() {
		t.Error("cached render produced different output")
	}
}

// TestHTMLEngineNoCaching tests dev mode without caching
func TestHTMLEngineNoCaching(t *testing.T) {
	engine := NewHTMLEngine("../_example/assets/", "html", nil, false)

	if engine.cache != nil {
		t.Error("cache map should be nil when caching disabled")
	}

	// Should still render correctly
	buf := new(bytes.Buffer)
	err := engine.Render(buf, "static.html", nil, nil)
	if err != nil {
		t.Errorf("render failed: %v", err)
	}
}

// TestHTMLEngineContentType tests ContentType method
func TestHTMLEngineContentType(t *testing.T) {
	engine := NewHTMLEngine("../_example/assets/", "html", nil, true)

	contentType := engine.ContentType()
	expected := "text/html; charset=utf-8"

	if contentType != expected {
		t.Errorf("ContentType() = %q, want %q", contentType, expected)
	}
}

// TestHTMLEngineCustomFuncMap tests custom template functions
func TestHTMLEngineCustomFuncMap(t *testing.T) {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	engine := NewHTMLEngine("../_example/assets/", "html", funcMap, true)

	if engine.funcMap == nil {
		t.Error("funcMap should not be nil")
	}
}

// TestHTMLEngineClearCache tests cache clearing
func TestHTMLEngineClearCache(t *testing.T) {
	engine := NewHTMLEngine("../_example/assets/", "html", nil, true)

	// Render to populate cache
	buf := new(bytes.Buffer)
	engine.Render(buf, "static.html", nil, nil)

	engine.Mu.RLock()
	cacheSize := len(engine.cache)
	engine.Mu.RUnlock()

	if cacheSize == 0 {
		t.Error("cache should not be empty after render")
	}

	// Clear cache
	engine.ClearCache()

	engine.Mu.RLock()
	newCacheSize := len(engine.cache)
	engine.Mu.RUnlock()

	if newCacheSize != 0 {
		t.Errorf("cache size after clear = %d, want 0", newCacheSize)
	}
}

// BenchmarkHTMLEngineWithCache benchmarks template rendering with cache
func BenchmarkHTMLEngineWithCache(b *testing.B) {
	engine := NewHTMLEngine("../_example/assets/", "html", nil, true)
	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		engine.Render(buf, "static.html", nil, nil)
	}
}

// BenchmarkHTMLEngineNoCache benchmarks template rendering without cache
func BenchmarkHTMLEngineNoCache(b *testing.B) {
	engine := NewHTMLEngine("../_example/assets/", "html", nil, false)
	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		engine.Render(buf, "static.html", nil, nil)
	}
}
