package arry

import (
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pionus/arry/engine"
)

// Engine is the interface for template rendering
type Engine interface {
	Render(w io.Writer, name string, data interface{}, ctx interface{}) error
	ContentType() string
}

// EngineType represents the type of template engine
type EngineType string

const (
	EngineHTML  EngineType = "html"
	EngineJSON  EngineType = "json"
	EngineYAML  EngineType = "yaml"
	EngineXML   EngineType = "xml"
	EnginePlain EngineType = "plain"
)

// EngineConfig holds configuration for template engine
type EngineConfig struct {
	Type      EngineType          // Engine type (html, json, yaml, xml, plain)
	Dir       string              // Template directory
	Extension string              // File extension (auto-detect from Type if empty)
	FuncMap   template.FuncMap    // Custom template functions (for HTML/Plain engines)
	Cache     bool                // Enable template caching (production: true, dev: false)
	Indent    string              // Indentation for JSON/XML (optional, default: "  ")
}

// detectEngineType infers engine type from file extension
func detectEngineType(extension string) EngineType {
	ext := strings.ToLower(strings.TrimPrefix(extension, "."))
	switch ext {
	case "html", "htm":
		return EngineHTML
	case "json":
		return EngineJSON
	case "yaml", "yml":
		return EngineYAML
	case "xml":
		return EngineXML
	case "txt", "text":
		return EnginePlain
	default:
		return EngineHTML
	}
}

// resolveDir resolves template directory path
func resolveDir(dir string) string {
	if dir == "" {
		return ""
	}
	if path.IsAbs(dir) {
		return dir
	}
	base, _ := os.Getwd()
	return path.Join(base, dir)
}

// NewEngine creates a new HTML template engine with default settings
// Deprecated: Use NewEngineWithConfig for more control
func NewEngine(dir string, t string) Engine {
	return NewEngineWithConfig(EngineConfig{
		Type:      EngineHTML,
		Dir:       dir,
		Extension: "html",
		Cache:     true,
	})
}

// NewEngineWithConfig creates a template engine based on configuration
// Supports multiple engine types: HTML, JSON, XML, Plain Text
func NewEngineWithConfig(config EngineConfig) Engine {
	// Auto-detect extension from type
	if config.Extension == "" && config.Type != "" {
		switch config.Type {
		case EngineHTML:
			config.Extension = "html"
		case EngineJSON:
			config.Extension = "json"
		case EngineYAML:
			config.Extension = "yaml"
		case EngineXML:
			config.Extension = "xml"
		case EnginePlain:
			config.Extension = "txt"
		}
	}

	// Auto-detect type from extension
	if config.Type == "" && config.Extension != "" {
		config.Type = detectEngineType(config.Extension)
	}

	// Default to HTML
	if config.Type == "" {
		config.Type = EngineHTML
	}

	basedir := resolveDir(config.Dir)

	// Factory pattern: create appropriate engine
	switch config.Type {
	case EngineHTML:
		return engine.NewHTMLEngine(basedir, config.Extension, config.FuncMap, config.Cache)
	case EngineJSON:
		return engine.NewJSONEngine(config.Indent)
	case EngineXML:
		return engine.NewXMLEngine(config.Indent)
	case EnginePlain:
		return engine.NewPlainEngine(basedir, config.Extension, config.FuncMap, config.Cache)
	case EngineYAML:
		// YAML engine requires external package
		// For now, treat as plain text
		// TODO: Implement YAMLEngine with gopkg.in/yaml.v3
		return engine.NewPlainEngine(basedir, config.Extension, config.FuncMap, config.Cache)
	default:
		return engine.NewHTMLEngine(basedir, config.Extension, config.FuncMap, config.Cache)
	}
}
