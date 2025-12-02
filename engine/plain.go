package engine

import (
	"io"
	"path"
	"text/template"
)

// PlainEngine renders using text/template (no HTML escaping)
type PlainEngine struct {
	BaseEngine                                  // Embedded base engine
	template   *template.Template               // Root template (private)
	cache      map[string]*template.Template    // Template cache (private)
	funcMap    template.FuncMap                 // Custom functions (private)
}

// Render renders a plain text template
func (e *PlainEngine) Render(w io.Writer, name string, data interface{}, ctx interface{}) error {
	if e.CacheEnabled {
		e.RLock()
		t, ok := e.cache[name]
		e.RUnlock()

		if !ok {
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

		return t.Execute(w, data)
	}

	t, err := e.template.Clone()
	if err != nil {
		return err
	}
	t, err = t.ParseFiles(e.GetFullPath(name))
	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

// ContentType returns plain text content type
func (e *PlainEngine) ContentType() string {
	return "text/plain; charset=utf-8"
}

// NewPlainEngine creates a new plain text template engine
func NewPlainEngine(basedir, extension string, funcMap template.FuncMap, cache bool) *PlainEngine {
	if extension == "" {
		extension = "txt"
	}

	pattern := path.Join(basedir, "*."+extension)

	tmpl := template.New("")
	if funcMap != nil {
		tmpl = tmpl.Funcs(funcMap)
	}
	tmpl = template.Must(tmpl.ParseGlob(pattern))

	engine := &PlainEngine{
		BaseEngine: NewBaseEngine(basedir, cache),
		template:   tmpl,
		funcMap:    funcMap,
	}

	if cache {
		engine.cache = make(map[string]*template.Template)
	}

	return engine
}
