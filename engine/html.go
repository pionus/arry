package engine

import (
	"html/template"
	"io"
	"io/fs"
	"path"
)

// HTMLEngine is the default HTML template engine
type HTMLEngine struct {
	BaseEngine                               // Embedded base engine
	template   *template.Template            // Root template (private)
	cache      map[string]*template.Template // Template cache (private)
	funcMap    template.FuncMap              // Custom functions (private)
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
			newTmpl, err := e.template.Clone()
			if err != nil {
				return err
			}
			newTmpl, err = e.parseFile(newTmpl, name)
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
	t, err = e.parseFile(t, name)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

func (e *HTMLEngine) parseFile(t *template.Template, name string) (*template.Template, error) {
	if e.FS != nil {
		return t.ParseFS(e.FS, e.GetFullPath(name))
	}
	return t.ParseFiles(e.GetFullPath(name))
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
	return newHTMLEngine(nil, basedir, extension, funcMap, cache)
}

// NewHTMLEngineFS creates an HTML engine that reads templates from fsys.
func NewHTMLEngineFS(fsys fs.FS, basedir, extension string, funcMap template.FuncMap, cache bool) *HTMLEngine {
	return newHTMLEngine(fsys, basedir, extension, funcMap, cache)
}

func newHTMLEngine(fsys fs.FS, basedir, extension string, funcMap template.FuncMap, cache bool) *HTMLEngine {
	if extension == "" {
		extension = "html"
	}

	pattern := path.Join(basedir, "*."+extension)

	tmpl := template.New("")
	if funcMap != nil {
		tmpl = tmpl.Funcs(funcMap)
	}
	if fsys != nil {
		tmpl = template.Must(tmpl.ParseFS(fsys, pattern))
	} else {
		tmpl = template.Must(tmpl.ParseGlob(pattern))
	}

	engine := &HTMLEngine{
		BaseEngine: BaseEngine{Basedir: basedir, FS: fsys, CacheEnabled: cache},
		template:   tmpl,
		funcMap:    funcMap,
	}

	if cache {
		engine.cache = make(map[string]*template.Template)
	}

	return engine
}
