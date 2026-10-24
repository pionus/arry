package engine

import (
	"io"
	"io/fs"
	"path"
	"text/template"
)

// PlainEngine renders using text/template (no HTML escaping)
type PlainEngine struct {
	BaseEngine                               // Embedded base engine
	template   *template.Template            // Root template (private)
	cache      map[string]*template.Template // Template cache (private)
	funcMap    template.FuncMap              // Custom functions (private)
}

// Render renders a plain text template
func (e *PlainEngine) Render(w io.Writer, name string, data interface{}, ctx interface{}) error {
	if e.CacheEnabled {
		e.RLock()
		t, ok := e.cache[name]
		e.RUnlock()

		if !ok {
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

func (e *PlainEngine) parseFile(t *template.Template, name string) (*template.Template, error) {
	if e.FS != nil {
		return t.ParseFS(e.FS, e.GetFullPath(name))
	}
	return t.ParseFiles(e.GetFullPath(name))
}

// ContentType returns plain text content type
func (e *PlainEngine) ContentType() string {
	return "text/plain; charset=utf-8"
}

// NewPlainEngine creates a new plain text template engine
func NewPlainEngine(basedir, extension string, funcMap template.FuncMap, cache bool) *PlainEngine {
	return newPlainEngine(nil, basedir, extension, funcMap, cache)
}

// NewPlainEngineFS creates a plain text engine that reads templates from fsys.
func NewPlainEngineFS(fsys fs.FS, basedir, extension string, funcMap template.FuncMap, cache bool) *PlainEngine {
	return newPlainEngine(fsys, basedir, extension, funcMap, cache)
}

func newPlainEngine(fsys fs.FS, basedir, extension string, funcMap template.FuncMap, cache bool) *PlainEngine {
	if extension == "" {
		extension = "txt"
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

	engine := &PlainEngine{
		BaseEngine: BaseEngine{Basedir: basedir, FS: fsys, CacheEnabled: cache},
		template:   tmpl,
		funcMap:    funcMap,
	}

	if cache {
		engine.cache = make(map[string]*template.Template)
	}

	return engine
}
