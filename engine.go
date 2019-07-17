package arry

import (
	"os"
	"io"
	"path"
	"html/template"
)

type Engine interface {
	Render(w io.Writer, name string, data interface{}, ctx Context) error
}

type HTMLEngine struct {
	template *template.Template
}

func (e *HTMLEngine) Render(w io.Writer, name string, data interface{}, ctx Context) error {
	return e.template.ExecuteTemplate(w, name, data)
}

func NewEngine(dir string, t string) Engine {
	base, _ := os.Getwd()
	p := path.Join(base, dir, "*.html")

	engine := &HTMLEngine{
		template: template.Must(template.ParseGlob(p)),
	}

	return engine
}
