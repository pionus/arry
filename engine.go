package arry

import (
	"os"
	"io"
	"path"
	"text/template"
)

type Engine interface {
	Render(w io.Writer, name string, data interface{}, ctx Context) error
}

type HTMLEngine struct {
	template *template.Template
	basedir string
}

func (e *HTMLEngine) Render(w io.Writer, name string, data interface{}, ctx Context) error {
	t, _ := e.template.Clone()
	t, _ = t.ParseFiles(path.Join(e.basedir, name))
	return t.ExecuteTemplate(w, name, data)
}

func NewEngine(dir string, t string) Engine {
	base, _ := os.Getwd()
	basedir := path.Join(base, dir)
	p := path.Join(basedir, "*.html")

	engine := &HTMLEngine{
		template: template.Must(template.ParseGlob(p)),
		basedir: basedir,
	}

	return engine
}
