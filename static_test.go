package arry

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func serve(a *Arry, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	return rec
}

func TestStaticServesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.css"), []byte("body{}"), 0644); err != nil {
		t.Fatal(err)
	}

	a := New()
	a.Static("/assets", dir)

	rec := serve(a, http.MethodGet, "/assets/app.css", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "body{}" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestStaticRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.MkdirAll(assets, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}

	a := New()
	a.Static("/assets", assets)

	rec := serve(a, http.MethodGet, "/assets/../../secret.txt", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("traversal status = %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatal("traversal leaked file contents")
	}
}

func TestStaticHeadAndCacheControl(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	a := New()
	a.StaticWithConfig("/assets", StaticConfig{Dir: dir, MaxAge: time.Hour})

	rec := serve(a, http.MethodHead, "/assets/a.txt", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD body = %q, want empty", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Fatalf("Cache-Control = %q", cc)
	}
}

func TestHeadFallsBackToGet(t *testing.T) {
	a := New()
	a.Router().Get("/hello", func(ctx Context) {
		ctx.Text(http.StatusOK, "hello")
	})

	rec := serve(a, http.MethodHead, "/hello", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD body = %q, want empty", rec.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	a := New()
	a.Router().Get("/hello", func(ctx Context) {
		ctx.Text(http.StatusOK, "hello")
	})

	rec := serve(a, http.MethodPost, "/hello", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	allow := rec.Header().Get("Allow")
	for _, m := range []string{"GET", "HEAD", "OPTIONS"} {
		if !strings.Contains(allow, m) {
			t.Fatalf("Allow = %q, missing %s", allow, m)
		}
	}
}

func TestOptionsReturnsAllowedMethods(t *testing.T) {
	a := New()
	a.Router().Get("/hello", func(ctx Context) {
		ctx.Text(http.StatusOK, "hello")
	})

	rec := serve(a, http.MethodOptions, "/hello", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow == "" {
		t.Fatal("missing Allow header")
	}
}

func TestNoContent(t *testing.T) {
	a := New()
	a.Router().Get("/empty", func(ctx Context) {
		ctx.NoContent(http.StatusNoContent)
	})

	rec := serve(a, http.MethodGet, "/empty", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestCatchAllFallsBackForStaticPrefix(t *testing.T) {
	a := New()
	r := a.Router()
	r.Get("/index.php/search/:q", func(ctx Context) {
		ctx.Text(http.StatusOK, "search")
	})
	r.Get("/*", func(ctx Context) {
		ctx.Text(http.StatusOK, "fallback:"+ctx.Param("*"))
	})

	code, body := request("GET", "/index.php", a)
	if code != 200 || body != "fallback:index.php" {
		t.Fatalf("code=%d body=%q", code, body)
	}
}
