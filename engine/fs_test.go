package engine

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHTMLEngineFS(t *testing.T) {
	fsys := fstest.MapFS{
		"views/layout.html": &fstest.MapFile{
			Data: []byte(`{{define "layout"}}<html><body>{{template "content" .}}</body></html>{{end}}`),
		},
		"views/page.html": &fstest.MapFile{
			Data: []byte(`{{template "layout" .}}{{define "content"}}Hello {{.Name}}{{end}}`),
		},
	}

	e := NewHTMLEngineFS(fsys, "views", "html", nil, true)

	var buf bytes.Buffer
	if err := e.Render(&buf, "page.html", map[string]string{"Name": "World"}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Hello World") {
		t.Fatalf("render = %q", buf.String())
	}

	// Cached second render should still work.
	buf.Reset()
	if err := e.Render(&buf, "page.html", map[string]string{"Name": "Again"}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Hello Again") {
		t.Fatalf("cached render = %q", buf.String())
	}
}

func TestPlainEngineFS(t *testing.T) {
	fsys := fstest.MapFS{
		"views/greeting.txt": &fstest.MapFile{Data: []byte(`Hello {{.}}`)},
	}

	e := NewPlainEngineFS(fsys, "views", "txt", nil, true)

	var buf bytes.Buffer
	if err := e.Render(&buf, "greeting.txt", "World", nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "Hello World" {
		t.Fatalf("render = %q", buf.String())
	}
}
