package arry

import (
	"net/http"
	"testing"
	"testing/fstest"
	"time"
)

func TestStaticFSWithConfig(t *testing.T) {
	fsys := fstest.MapFS{
		"app.css":   &fstest.MapFile{Data: []byte("body{}")},
		"dir/a.txt": &fstest.MapFile{Data: []byte("x")},
	}

	a := New()
	a.StaticFSWithConfig("/assets", fsys, StaticConfig{MaxAge: time.Hour})

	rec := serve(a, http.MethodGet, "/assets/app.css", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "body{}" {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Fatalf("Cache-Control=%q", cc)
	}

	// Directories are not listed.
	if rec := serve(a, http.MethodGet, "/assets/", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("dir status=%d", rec.Code)
	}
	if rec := serve(a, http.MethodGet, "/assets/dir/", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("nested dir status=%d", rec.Code)
	}
	if rec := serve(a, http.MethodGet, "/assets/dir/a.txt", nil); rec.Code != http.StatusOK {
		t.Fatalf("file status=%d", rec.Code)
	}
}
