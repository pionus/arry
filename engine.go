package arry

import (
	"io"
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

func NewEngine(path string, t string) Engine {
	engine := &HTMLEngine{
		template: template.Must(template.ParseGlob(path + "*.html")),
	}

	return engine
}
