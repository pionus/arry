package arry

import (
	"bytes"
	"testing"

	enginepkg "github.com/pionus/arry/engine"
)

// TestNewEngine tests the legacy NewEngine factory function
func TestNewEngine(t *testing.T) {
	engine := NewEngine("_example/assets/", "html")

	if engine == nil {
		t.Fatal("NewEngine should not return nil")
	}

	// Should create an HTMLEngine
	if _, ok := engine.(*enginepkg.HTMLEngine); !ok {
		t.Error("NewEngine should create HTMLEngine by default")
	}

	// Should have correct content type
	if ct := engine.ContentType(); ct != "text/html; charset=utf-8" {
		t.Errorf("ContentType = %q, want text/html", ct)
	}

	// Should be able to render
	buf := new(bytes.Buffer)
	err := engine.Render(buf, "static.html", nil, nil)
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}
}

// TestEngineTypeDetection tests automatic type detection from extension
func TestEngineTypeDetection(t *testing.T) {
	tests := []struct {
		extension string
		wantType  EngineType
	}{
		{"html", EngineHTML},
		{"htm", EngineHTML},
		{"json", EngineJSON},
		{"xml", EngineXML},
		{"txt", EnginePlain},
		{"yaml", EngineYAML},
		{"yml", EngineYAML},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			got := detectEngineType(tt.extension)
			if got != tt.wantType {
				t.Errorf("detectEngineType(%q) = %v, want %v", tt.extension, got, tt.wantType)
			}
		})
	}
}

// TestEngineAutoDetect tests that engine type is auto-detected from extension
func TestEngineAutoDetect(t *testing.T) {
	// Create JSON engine by specifying extension only
	engine := NewEngineWithConfig(EngineConfig{
		Extension: "json",
	})

	if _, ok := engine.(*enginepkg.JSONEngine); !ok {
		t.Error("Should auto-detect JSON engine from extension")
	}

	// Create XML engine by specifying extension only
	engine = NewEngineWithConfig(EngineConfig{
		Extension: "xml",
	})

	if _, ok := engine.(*enginepkg.XMLEngine); !ok {
		t.Error("Should auto-detect XML engine from extension")
	}

	// Create HTML engine by default
	engine = NewEngineWithConfig(EngineConfig{
		Dir: "_example/assets/",
	})

	if _, ok := engine.(*enginepkg.HTMLEngine); !ok {
		t.Error("Should default to HTML engine when no type/extension specified")
	}
}

// TestNewEngineWithConfig tests the factory method with various configurations
func TestNewEngineWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config EngineConfig
		want   string // engine type name
	}{
		{
			name:   "HTML by type",
			config: EngineConfig{Type: EngineHTML, Dir: "_example/assets/"},
			want:   "HTMLEngine",
		},
		{
			name:   "JSON by type",
			config: EngineConfig{Type: EngineJSON},
			want:   "JSONEngine",
		},
		{
			name:   "XML by type",
			config: EngineConfig{Type: EngineXML},
			want:   "XMLEngine",
		},
		{
			name:   "Plain by type",
			config: EngineConfig{Type: EnginePlain, Dir: "_example/assets/", Extension: "html"},
			want:   "PlainEngine",
		},
		{
			name:   "HTML by extension",
			config: EngineConfig{Extension: "html", Dir: "_example/assets/"},
			want:   "HTMLEngine",
		},
		{
			name:   "JSON by extension",
			config: EngineConfig{Extension: "json"},
			want:   "JSONEngine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngineWithConfig(tt.config)
			if engine == nil {
				t.Fatal("NewEngineWithConfig returned nil")
			}

			// Check engine type by attempting type assertion
			var gotType string
			switch engine.(type) {
			case *enginepkg.HTMLEngine:
				gotType = "HTMLEngine"
			case *enginepkg.JSONEngine:
				gotType = "JSONEngine"
			case *enginepkg.XMLEngine:
				gotType = "XMLEngine"
			case *enginepkg.PlainEngine:
				gotType = "PlainEngine"
			default:
				gotType = "Unknown"
			}

			if gotType != tt.want {
				t.Errorf("NewEngineWithConfig created %s, want %s", gotType, tt.want)
			}
		})
	}
}

// TestEngineIntegration tests end-to-end engine functionality
func TestEngineIntegration(t *testing.T) {
	// Test that different engines can coexist
	htmlEngine := NewEngineWithConfig(EngineConfig{
		Type:  EngineHTML,
		Dir:   "_example/assets/",
		Cache: true,
	})

	jsonEngine := NewEngineWithConfig(EngineConfig{
		Type: EngineJSON,
	})

	xmlEngine := NewEngineWithConfig(EngineConfig{
		Type: EngineXML,
	})

	// All should be non-nil
	if htmlEngine == nil || jsonEngine == nil || xmlEngine == nil {
		t.Fatal("One or more engines failed to create")
	}

	// Each should have correct content type
	if htmlEngine.ContentType() != "text/html; charset=utf-8" {
		t.Error("HTMLEngine has wrong content type")
	}

	if jsonEngine.ContentType() != "application/json; charset=utf-8" {
		t.Error("JSONEngine has wrong content type")
	}

	if xmlEngine.ContentType() != "application/xml; charset=utf-8" {
		t.Error("XMLEngine has wrong content type")
	}
}
