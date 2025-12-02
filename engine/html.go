package engine

import (
	"io"
	"path"
	"html/template"
)

// HTMLEngine is the default HTML template engine
type HTMLEngine struct {
	BaseEngine                                   // Embedded base engine
	template   *template.Template                // Root template (private)
	cache      map[string]*template.Template     // Template cache (private)
	funcMap    template.FuncMap                  // Custom functions (private)
}

// Render renders a template with data
func (e *HTMLEngine) Render(w io.Writer, name string, data interface{}, ctx interface{}) error {
	if e.CacheEnabled {
		// Cache enabled: check cache first
		e.RLock()
		t, ok := e.cache[name]
		e.RUnlock()

		if !ok {
			// Cache miss: parse and cache
			fullPath := e.GetFullPath(name)
			newTmpl, err := e.template.Clone()
			if err != nil {
				return err
			}
			newTmpl, err = newTmpl.ParseFiles(fullPath)
			if err != nil {
				return err
			}

			e.Lock()
			e.cache[name] = newTmpl
			e.Unlock()
			t = newTmpl
		}

		return t.ExecuteTemplate(w, name, data)
	}

	// Cache disabled: parse every time (development mode)
	t, err := e.template.Clone()
	if err != nil {
		return err
	}
	t, err = t.ParseFiles(e.GetFullPath(name))
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

// ContentType returns the content type for HTML templates
func (e *HTMLEngine) ContentType() string {
	return "text/html; charset=utf-8"
}

// ClearCache clears the template cache (useful for development hot-reload)
func (e *HTMLEngine) ClearCache() {
	e.Lock()
	e.cache = make(map[string]*template.Template)
	e.Unlock()
}

// NewHTMLEngine creates a new HTML template engine
func NewHTMLEngine(basedir, extension string, funcMap template.FuncMap, cache bool) *HTMLEngine {
	if extension == "" {
		extension = "html"
	}

	pattern := path.Join(basedir, "*."+extension)

	tmpl := template.New("")
	if funcMap != nil {
		tmpl = tmpl.Funcs(funcMap)
	}
	tmpl = template.Must(tmpl.ParseGlob(pattern))

	engine := &HTMLEngine{
		BaseEngine: NewBaseEngine(basedir, cache),
		template:   tmpl,
		funcMap:    funcMap,
	}

	if cache {
		engine.cache = make(map[string]*template.Template)
	}

	return engine
}
